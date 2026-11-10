// Copyright (c) Trifork

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// newFakeModelPoolAPI is an in-memory /v1/model-pools that applies the API's
// defaults, so the resource lifecycle runs without a live Corax API. calls
// records the member endpoint hits so tests can assert the reconcile path.
func newFakeModelPoolAPI(t *testing.T) (*httptest.Server, *[]string) {
	t.Helper()

	var (
		mu    sync.Mutex
		pools = map[string]map[string]interface{}{}
		next  = 0
		calls []string
	)
	newID := func() string {
		next++
		return fmt.Sprintf("00000000-0000-0000-0000-%012d", next)
	}
	withMemberDefaults := func(poolID string, m map[string]interface{}) map[string]interface{} {
		for k, v := range map[string]interface{}{"weight": 1, "priority": 1, "is_active": true} {
			if _, ok := m[k]; !ok {
				m[k] = v
			}
		}
		m["id"] = newID()
		m["pool_id"] = poolID
		return m
	}
	writeJSON := func(w http.ResponseWriter, status int, body interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}
	decode := func(r *http.Request) map[string]interface{} {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("bad request body: %v", err)
		}
		return body
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/model-pools", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		body := decode(r)
		id := newID()
		members := []interface{}{}
		if raw, ok := body["members"].([]interface{}); ok {
			for _, m := range raw {
				members = append(members, withMemberDefaults(id, m.(map[string]interface{})))
			}
		}
		pool := map[string]interface{}{
			"id": id, "name": body["name"], "strategy": "weighted_random", "cooldown_seconds": 60,
			"fallback_mode": "try_next", "created_by": "fake-user", "created_at": time.Now().UTC().Format(time.RFC3339),
			"members": members,
		}
		for _, k := range []string{"strategy", "cooldown_seconds", "fallback_mode"} {
			if v, ok := body[k]; ok {
				pool[k] = v
			}
		}
		pools[id] = pool
		writeJSON(w, http.StatusCreated, pool)
	})
	mux.HandleFunc("/v1/model-pools/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/model-pools/"), "/")
		pool, ok := pools[parts[0]]
		if !ok {
			http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
			return
		}

		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				writeJSON(w, http.StatusOK, pool)
			case http.MethodPut:
				for k, v := range decode(r) {
					if v != nil {
						pool[k] = v
					}
				}
				writeJSON(w, http.StatusOK, pool)
			case http.MethodDelete:
				delete(pools, parts[0])
				w.WriteHeader(http.StatusNoContent)
			}
			return
		}

		members := pool["members"].([]interface{})
		calls = append(calls, r.Method+" member")
		switch {
		case r.Method == http.MethodPost:
			m := withMemberDefaults(parts[0], decode(r))
			pool["members"] = append(members, m)
			writeJSON(w, http.StatusCreated, m)
		case len(parts) == 3:
			for i, raw := range members {
				m := raw.(map[string]interface{})
				if m["id"] != parts[2] {
					continue
				}
				if r.Method == http.MethodDelete {
					pool["members"] = append(members[:i], members[i+1:]...)
					w.WriteHeader(http.StatusNoContent)
					return
				}
				for k, v := range decode(r) {
					if v != nil {
						m[k] = v
					}
				}
				writeJSON(w, http.StatusOK, m)
				return
			}
			http.Error(w, `{"detail":"not found"}`, http.StatusNotFound)
		}
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, &calls
}

const (
	deploymentA = "aaaaaaaa-0000-0000-0000-000000000000"
	deploymentB = "bbbbbbbb-0000-0000-0000-000000000000"
	deploymentC = "cccccccc-0000-0000-0000-000000000000"
)

func TestAccModelPoolResource_againstFakeAPI(t *testing.T) {
	server, calls := newFakeModelPoolAPI(t)
	t.Setenv("TF_ACC", "1")
	t.Setenv("CORAX_API_ENDPOINT", server.URL)
	t.Setenv("CORAX_API_KEY", "fake-api-key")

	name := "corax_model_pool.test"
	// expectCalls compares member calls ignoring order (members are a map),
	// except that the last call must match: removals have to run last.
	expectCalls := func(want ...string) resource.TestCheckFunc {
		return func(*terraform.State) error {
			got := append([]string(nil), *calls...)
			*calls = nil
			if len(got) > 0 && len(want) > 0 && got[len(got)-1] != want[len(want)-1] {
				return fmt.Errorf("last member call = %q, want %q (all: %v)", got[len(got)-1], want[len(want)-1], got)
			}
			sort.Strings(got)
			sort.Strings(want)
			if strings.Join(got, ",") != strings.Join(want, ",") {
				return fmt.Errorf("member calls = %v, want %v", got, want)
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with defaults: members go in the create payload, no member calls.
			{
				Config: fmt.Sprintf(`
resource "corax_model_pool" "test" {
  name = "pool"
  members = {
    %q = {}
    %q = { weight = 3 }
  }
}`, deploymentA, deploymentB),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(name, "id"),
					resource.TestCheckResourceAttr(name, "strategy", "weighted_random"),
					resource.TestCheckResourceAttr(name, "cooldown_seconds", "60"),
					resource.TestCheckResourceAttr(name, "fallback_mode", "try_next"),
					resource.TestCheckResourceAttr(name, "members.%", "2"),
					resource.TestCheckResourceAttr(name, "members."+deploymentA+".weight", "1"),
					resource.TestCheckResourceAttr(name, "members."+deploymentB+".weight", "3"),
					resource.TestCheckResourceAttr(name, "members."+deploymentB+".priority", "1"),
					resource.TestCheckResourceAttr(name, "members."+deploymentB+".is_active", "true"),
					expectCalls(),
				),
			},
			{ResourceName: name, ImportState: true, ImportStateVerify: true},
			// Update: change A, drop B, add C, change pool settings.
			{
				Config: fmt.Sprintf(`
resource "corax_model_pool" "test" {
  name             = "pool-renamed"
  fallback_mode    = "fail_fast"
  cooldown_seconds = 30
  members = {
    %q = { priority = 2, is_active = false }
    %q = {}
  }
}`, deploymentA, deploymentC),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "pool-renamed"),
					resource.TestCheckResourceAttr(name, "fallback_mode", "fail_fast"),
					resource.TestCheckResourceAttr(name, "cooldown_seconds", "30"),
					resource.TestCheckResourceAttr(name, "members.%", "2"),
					resource.TestCheckResourceAttr(name, "members."+deploymentA+".priority", "2"),
					resource.TestCheckResourceAttr(name, "members."+deploymentA+".is_active", "false"),
					resource.TestCheckResourceAttr(name, "members."+deploymentC+".weight", "1"),
					resource.TestCheckNoResourceAttr(name, "members."+deploymentB+".weight"),
					// Removal runs last so the pool is never empty.
					expectCalls("PUT member", "POST member", "DELETE member"),
				),
			},
		},
	})
}

func TestAccCompletionCapabilityResource_modelIDConflictsWithPool(t *testing.T) {
	t.Setenv("TF_ACC", "1")
	t.Setenv("CORAX_API_ENDPOINT", "http://127.0.0.1:1")
	t.Setenv("CORAX_API_KEY", "fake-api-key")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`
resource "corax_completion_capability" "test" {
  name              = "conflict"
  model_id          = %q
  model_pool_id     = %q
  system_prompt     = "s"
  completion_prompt = "c"
  output_type       = "text"
}`, deploymentA, deploymentB),
			ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
		}},
	})
}
