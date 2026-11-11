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

	api "terraform-provider-corax/internal/generated"
)

// fakeModelPoolAPI is an in-memory /v1/model-pools, so the resource lifecycle
// runs without a live Corax API. It stores what the provider sends and fills in
// no defaults, so a field the provider forgets to send shows up as a diff.
type fakeModelPoolAPI struct {
	*httptest.Server

	mu    sync.Mutex
	pools map[string]*api.ModelDeploymentPool
	next  int
	// calls records member endpoint hits, so tests can assert the reconcile path.
	calls []string
	// failAdd makes POST .../members fail with 422.
	failAdd bool
	// memberGoneOnDelete makes DELETE .../members/{id} answer 404, as if the
	// member was removed concurrently.
	memberGoneOnDelete bool
}

func newFakeModelPoolAPI(t *testing.T) *fakeModelPoolAPI {
	t.Helper()
	f := &fakeModelPoolAPI{pools: map[string]*api.ModelDeploymentPool{}}

	newID := func() string {
		f.next++
		return fmt.Sprintf("00000000-0000-0000-0000-%012d", f.next)
	}
	newMember := func(poolID string, c api.PoolMemberCreate) api.PoolMember {
		return api.PoolMember{
			Id: newID(), PoolId: poolID, ModelDeploymentId: c.ModelDeploymentId,
			Weight: c.Weight, Priority: c.Priority, IsActive: c.IsActive,
		}
	}
	writeJSON := func(w http.ResponseWriter, status int, body interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}
	decode := func(r *http.Request, into interface{}) {
		if err := json.NewDecoder(r.Body).Decode(into); err != nil {
			t.Errorf("bad request body: %v", err)
		}
	}
	notFound := func(w http.ResponseWriter) { http.Error(w, `{"detail":"not found"}`, http.StatusNotFound) }

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/model-pools", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		var create api.ModelDeploymentPoolCreate
		decode(r, &create)
		pool := &api.ModelDeploymentPool{
			Id: newID(), Name: create.Name, Strategy: create.Strategy, CooldownSeconds: create.CooldownSeconds,
			FallbackMode: create.FallbackMode, CreatedBy: "fake-user", CreatedAt: time.Now().UTC(),
			Members: []api.PoolMember{},
		}
		for _, c := range create.Members {
			pool.Members = append(pool.Members, newMember(pool.Id, c))
		}
		f.pools[pool.Id] = pool
		writeJSON(w, http.StatusCreated, pool)
	})
	mux.HandleFunc("/v1/model-pools/", func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/model-pools/"), "/")
		pool, ok := f.pools[parts[0]]
		if !ok {
			notFound(w)
			return
		}

		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				writeJSON(w, http.StatusOK, pool)
			case http.MethodPut:
				var u api.ModelDeploymentPoolUpdate
				decode(r, &u)
				if v := u.Name.Get(); v != nil {
					pool.Name = *v
				}
				if v := u.Strategy.Get(); v != nil {
					pool.Strategy = v
				}
				if v := u.CooldownSeconds.Get(); v != nil {
					pool.CooldownSeconds = v
				}
				if v := u.FallbackMode.Get(); v != nil {
					pool.FallbackMode = v
				}
				// Like the real PUT (as far as we know), answer without members:
				// the provider must not rely on them being here.
				resp := *pool
				resp.Members = nil
				writeJSON(w, http.StatusOK, resp)
			case http.MethodDelete:
				delete(f.pools, parts[0])
				w.WriteHeader(http.StatusNoContent)
			}
			return
		}

		f.calls = append(f.calls, r.Method+" member")
		if r.Method == http.MethodPost {
			if f.failAdd {
				http.Error(w, `{"detail":"invalid model deployment"}`, http.StatusUnprocessableEntity)
				return
			}
			var c api.PoolMemberCreate
			decode(r, &c)
			m := newMember(pool.Id, c)
			pool.Members = append(pool.Members, m)
			writeJSON(w, http.StatusCreated, m)
			return
		}
		for i := range pool.Members {
			m := &pool.Members[i]
			if len(parts) != 3 || m.Id != parts[2] {
				continue
			}
			if r.Method == http.MethodDelete {
				pool.Members = append(pool.Members[:i], pool.Members[i+1:]...)
				if f.memberGoneOnDelete {
					notFound(w)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			var u api.PoolMemberUpdate
			decode(r, &u)
			if v := u.Weight.Get(); v != nil {
				m.Weight = v
			}
			if v := u.Priority.Get(); v != nil {
				m.Priority = v
			}
			if v := u.IsActive.Get(); v != nil {
				m.IsActive = v
			}
			writeJSON(w, http.StatusOK, m)
			return
		}
		notFound(w)
	})

	f.Server = httptest.NewServer(mux)
	t.Cleanup(f.Close)
	return f
}

// member returns the stored member for a deployment, or nil.
func (f *fakeModelPoolAPI) member(deploymentID string) *api.PoolMember {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, p := range f.pools {
		for i := range p.Members {
			if p.Members[i].ModelDeploymentId == deploymentID {
				return &p.Members[i]
			}
		}
	}
	return nil
}

func (f *fakeModelPoolAPI) set(apply func(*fakeModelPoolAPI)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	apply(f)
}

// expectCalls compares member calls ignoring order (members are a map), except
// that the last call must match: removals have to run last.
func (f *fakeModelPoolAPI) expectCalls(want ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		f.mu.Lock()
		got := append([]string(nil), f.calls...)
		f.calls = nil
		f.mu.Unlock()
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

func (f *fakeModelPoolAPI) checkDestroyed(*terraform.State) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.pools) != 0 {
		return fmt.Errorf("%d pool(s) left after destroy", len(f.pools))
	}
	return nil
}

func useFakeAPI(t *testing.T, url string) {
	t.Setenv("TF_ACC", "1")
	t.Setenv("CORAX_API_ENDPOINT", url)
	t.Setenv("CORAX_API_KEY", "fake-api-key")
}

const (
	deploymentA = "aaaaaaaa-0000-0000-0000-000000000000"
	deploymentB = "bbbbbbbb-0000-0000-0000-000000000000"
	deploymentC = "cccccccc-0000-0000-0000-000000000000"
	deploymentD = "dddddddd-0000-0000-0000-000000000000"
	poolName    = "corax_model_pool.test"
)

func poolConfig(settings, members string) string {
	return fmt.Sprintf(`
resource "corax_model_pool" "test" {
  name = "pool"
  %s
  members = {
    %s
  }
}`, settings, members)
}

func TestAccModelPoolResource_againstFakeAPI(t *testing.T) {
	fake := newFakeModelPoolAPI(t)
	useFakeAPI(t, fake.URL)
	m := func(id string) string { return "members." + id + "." }

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             fake.checkDestroyed,
		Steps: []resource.TestStep{
			// Create with defaults and non-defaults: everything goes in the
			// create payload, no member calls.
			{
				Config: fmt.Sprintf(`
resource "corax_model_pool" "test" {
  name             = "pool"
  cooldown_seconds = 30
  fallback_mode    = "fail_fast"
  members = {
    %q = {}
    %q = { weight = 2, priority = 3, is_active = false }
    %q = {}
  }
}`, deploymentA, deploymentB, deploymentD),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(poolName, "id"),
					resource.TestCheckResourceAttr(poolName, "strategy", "weighted_random"),
					resource.TestCheckResourceAttr(poolName, "cooldown_seconds", "30"),
					resource.TestCheckResourceAttr(poolName, "fallback_mode", "fail_fast"),
					resource.TestCheckResourceAttr(poolName, "members.%", "3"),
					resource.TestCheckResourceAttr(poolName, m(deploymentA)+"weight", "1"),
					resource.TestCheckResourceAttr(poolName, m(deploymentA)+"priority", "1"),
					resource.TestCheckResourceAttr(poolName, m(deploymentA)+"is_active", "true"),
					resource.TestCheckResourceAttr(poolName, m(deploymentB)+"weight", "2"),
					resource.TestCheckResourceAttr(poolName, m(deploymentB)+"priority", "3"),
					resource.TestCheckResourceAttr(poolName, m(deploymentB)+"is_active", "false"),
					fake.expectCalls(),
				),
			},
			{ResourceName: poolName, ImportState: true, ImportStateVerify: true},
			// Pool settings, a weight-only change on B, add C with
			// non-defaults, drop D (which the API already lost: 404 is fine).
			// A is unchanged and must not be touched.
			{
				PreConfig: func() { fake.set(func(f *fakeModelPoolAPI) { f.memberGoneOnDelete = true }) },
				Config: fmt.Sprintf(`
resource "corax_model_pool" "test" {
  name             = "pool-renamed"
  cooldown_seconds = 45
  members = {
    %q = {}
    %q = { weight = 5, priority = 3, is_active = false }
    %q = { weight = 4, priority = 2, is_active = false }
  }
}`, deploymentA, deploymentB, deploymentC),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(poolName, "name", "pool-renamed"),
					resource.TestCheckResourceAttr(poolName, "cooldown_seconds", "45"),
					resource.TestCheckResourceAttr(poolName, "fallback_mode", "try_next"),
					resource.TestCheckResourceAttr(poolName, "members.%", "3"),
					resource.TestCheckResourceAttr(poolName, m(deploymentB)+"weight", "5"),
					resource.TestCheckResourceAttr(poolName, m(deploymentC)+"weight", "4"),
					resource.TestCheckResourceAttr(poolName, m(deploymentC)+"priority", "2"),
					resource.TestCheckResourceAttr(poolName, m(deploymentC)+"is_active", "false"),
					resource.TestCheckNoResourceAttr(poolName, m(deploymentD)+"weight"),
					fake.expectCalls("PUT member", "POST member", "DELETE member"),
				),
			},
			// A priority-only change on A and an is_active-only change on B.
			{
				PreConfig: func() { fake.set(func(f *fakeModelPoolAPI) { f.memberGoneOnDelete = false }) },
				Config: fmt.Sprintf(`
resource "corax_model_pool" "test" {
  name             = "pool-renamed"
  cooldown_seconds = 45
  members = {
    %q = { priority = 2 }
    %q = { weight = 5, priority = 3 }
    %q = { weight = 4, priority = 2, is_active = false }
  }
}`, deploymentA, deploymentB, deploymentC),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(poolName, m(deploymentA)+"priority", "2"),
					resource.TestCheckResourceAttr(poolName, m(deploymentB)+"is_active", "true"),
					fake.expectCalls("PUT member", "PUT member"),
				),
			},
		},
	})
}

// A pool deleted outside Terraform is dropped from state and planned for
// re-creation, and destroying it then is not an error.
func TestAccModelPoolResource_deletedOutsideTerraform(t *testing.T) {
	fake := newFakeModelPoolAPI(t)
	useFakeAPI(t, fake.URL)
	config := poolConfig("", fmt.Sprintf("%q = {}", deploymentA))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config},
			{
				PreConfig: func() {
					fake.set(func(f *fakeModelPoolAPI) {
						for id := range f.pools {
							delete(f.pools, id)
						}
					})
				},
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Replacing the only active member: when adding the new one fails, the old
// one must still be active. Deactivations run after additions.
func TestAccModelPoolResource_failedAddKeepsServingMember(t *testing.T) {
	fake := newFakeModelPoolAPI(t)
	useFakeAPI(t, fake.URL)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: poolConfig("", fmt.Sprintf("%q = {}", deploymentA))},
			{
				PreConfig:   func() { fake.set(func(f *fakeModelPoolAPI) { f.failAdd = true }) },
				Config:      poolConfig("", fmt.Sprintf("%q = { is_active = false }\n    %q = {}", deploymentA, deploymentC)),
				ExpectError: regexp.MustCompile(`Unable to add model deployment`),
			},
			{
				PreConfig: func() {
					if a := fake.member(deploymentA); a == nil || !a.GetIsActive() {
						t.Errorf("member A was deactivated before its replacement was added: %+v", a)
					}
					fake.set(func(f *fakeModelPoolAPI) { f.failAdd = false })
				},
				Config: poolConfig("", fmt.Sprintf("%q = {}", deploymentA)),
			},
		},
	})
}

func TestAccModelPoolResource_validation(t *testing.T) {
	useFakeAPI(t, "http://127.0.0.1:1")
	cases := map[string]struct{ settings, members, err string }{
		"empty members":  {"", "", `members.*at least 1`},
		"weight 0":       {"", fmt.Sprintf("%q = { weight = 0 }", deploymentA), `weight.*between 1 and 2147483647`},
		"bad fallback":   {`fallback_mode = "bogus"`, fmt.Sprintf("%q = {}", deploymentA), `fallback_mode.*value must be one of`},
		"cooldown -1":    {`cooldown_seconds = -1`, fmt.Sprintf("%q = {}", deploymentA), `cooldown_seconds.*between 0 and 2147483647`},
		"weight > int32": {"", fmt.Sprintf("%q = { weight = 4294967298 }", deploymentA), `weight.*between 1 and 2147483647`},
		"upper-case key": {"", fmt.Sprintf("%q = {}", strings.ToUpper(deploymentA)), `members.*lower-case UUID`},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{{
					Config:      poolConfig(c.settings, c.members),
					PlanOnly:    true,
					ExpectError: regexp.MustCompile(`(?s)` + c.err),
				}},
			})
		})
	}
}

func TestAccCapabilityResource_modelIDConflictsWithPool(t *testing.T) {
	useFakeAPI(t, "http://127.0.0.1:1")
	for _, extra := range map[string]string{
		"chat":       `resource "corax_chat_capability" "test" {`,
		"completion": `resource "corax_completion_capability" "test" {` + "\n  completion_prompt = \"c\"\n  output_type = \"text\"",
	} {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{{
				Config: fmt.Sprintf(`%s
  name          = "conflict"
  system_prompt = "s"
  model_id      = %q
  model_pool_id = %q
}`, extra, deploymentA, deploymentB),
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			}},
		})
	}
}

// Switching a capability between model_id and model_pool_id, and removing
// both, sends the unused field as an explicit null so the server clears it.
func TestAccCapabilityResource_modelRoutingAgainstFakeAPI(t *testing.T) {
	for _, c := range []struct{ capType, header string }{
		{"chat", `resource "corax_chat_capability" "test" {`},
		{"completion", `resource "corax_completion_capability" "test" {` + "\n  completion_prompt = \"c\"\n  output_type = \"text\""},
	} {
		t.Run(c.capType, func(t *testing.T) {
			server, puts := newFakeCapabilityAPIOfType(t, c.capType)
			useFakeAPI(t, server.URL)
			name := "corax_" + c.capType + "_capability.test"
			config := func(routing string) string {
				return fmt.Sprintf("%s\n  name = \"routing\"\n  system_prompt = \"s\"\n  %s\n}", c.header, routing)
			}
			// lastPut checks the most recent PUT body: want maps key -> value,
			// where nil means the key must be present with JSON null.
			lastPut := func(want map[string]interface{}) resource.TestCheckFunc {
				return func(*terraform.State) error {
					if len(*puts) == 0 {
						return fmt.Errorf("no PUT recorded")
					}
					body := (*puts)[len(*puts)-1]
					for k, v := range want {
						got, ok := body[k]
						if !ok || got != v {
							return fmt.Errorf("PUT %s = %v (present %v), want %v", k, got, ok, v)
						}
					}
					return nil
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: config(fmt.Sprintf("model_pool_id = %q", deploymentB)),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(name, "model_pool_id", deploymentB),
							resource.TestCheckNoResourceAttr(name, "model_id"),
						),
					},
					{
						Config: config(fmt.Sprintf("model_id = %q", deploymentA)),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(name, "model_id", deploymentA),
							resource.TestCheckNoResourceAttr(name, "model_pool_id"),
							lastPut(map[string]interface{}{"model_id": deploymentA, "model_pool_id": nil}),
						),
					},
					{
						Config: config(fmt.Sprintf("model_pool_id = %q", deploymentB)),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr(name, "model_pool_id", deploymentB),
							resource.TestCheckNoResourceAttr(name, "model_id"),
							lastPut(map[string]interface{}{"model_id": nil, "model_pool_id": deploymentB}),
						),
					},
					{
						Config: config(""),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckNoResourceAttr(name, "model_pool_id"),
							lastPut(map[string]interface{}{"model_id": nil, "model_pool_id": nil}),
						),
					},
				},
			})
		})
	}
}
