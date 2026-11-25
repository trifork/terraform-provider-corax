// Copyright (c) Trifork

package coraxclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

const (
	mountNotConfigured     = "Certificate authentication is not configured. Ask your platform administrator to configure the certificate mount directory."
	certificateNotReadable = "No readable mounted certificate is available for this client ID. Ask your platform administrator to mount <client-id>.pem."
)

type modelProviderRoundTripper func(*http.Request) (*http.Response, error)

func (f modelProviderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func modelProviderResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

func readinessErrorBody(detail string) string {
	body, _ := json.Marshal(map[string]string{"detail": detail})
	return string(body)
}

func newModelProviderRetryClient(t *testing.T, transport modelProviderRoundTripper) *Client {
	t.Helper()
	client, err := NewClient("https://example.invalid", "test-api-key")
	if err != nil {
		t.Fatal(err)
	}
	client.httpClient.Transport = transport
	return client
}

func mutateModelProvider(ctx context.Context, client *Client, method, providerType string) (*ModelProvider, error) {
	configuration := map[string]string{
		"api_endpoint": "https://example.cognitiveservices.azure.com/",
		"tenant_id":    "test-tenant", "client_id": "test-client", "certificate_password": "test-password",
	}
	if method == http.MethodPost {
		return client.CreateModelProvider(ctx, ModelProviderCreate{
			Name: "Test provider", ProviderType: providerType, Configuration: configuration,
		})
	}
	return client.UpdateModelProvider(ctx, "00000000-0000-0000-0000-000000000001", ModelProviderUpdate{
		ID:   "00000000-0000-0000-0000-000000000001",
		Name: "Test provider", ProviderType: providerType, Configuration: configuration,
	})
}

func TestModelProviderRetriesCertificateReadiness(t *testing.T) {
	for _, providerType := range []string{
		"azure_service_principal_certificate",
		"azure_ai_service_principal_certificate",
		"azure_speech_service_principal_certificate",
	} {
		for _, method := range []string{http.MethodPost, http.MethodPut} {
			t.Run(providerType+"/"+method, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					started := time.Now()
					var times []time.Duration
					var bodies [][]byte
					client := newModelProviderRetryClient(t, func(req *http.Request) (*http.Response, error) {
						var apiKey string
						for name, values := range req.Header {
							if strings.EqualFold(name, apiKeyHeader) {
								apiKey = strings.Join(values, "")
							}
						}
						if req.Method != method || apiKey != "test-api-key" {
							t.Fatalf("Unexpected request: %s, auth %q", req.Method, apiKey)
						}
						body, err := io.ReadAll(req.Body)
						if err != nil {
							t.Fatal(err)
						}
						bodies = append(bodies, body)
						times = append(times, time.Since(started))
						if len(times) <= 2 {
							detail := mountNotConfigured
							if len(times) == 2 {
								detail = certificateNotReadable
							}
							return modelProviderResponse(req, http.StatusBadRequest, readinessErrorBody(detail)), nil
						}
						var payload map[string]interface{}
						if err := json.Unmarshal(body, &payload); err != nil {
							t.Fatal(err)
						}
						payload["id"] = "00000000-0000-0000-0000-000000000001"
						payload["created_at"] = "2026-01-01T00:00:00Z"
						payload["created_by"] = "terraform-test"
						response, err := json.Marshal(payload)
						if err != nil {
							t.Fatal(err)
						}
						return modelProviderResponse(req, http.StatusOK, string(response)), nil
					})
					result, err := mutateModelProvider(t.Context(), client, method, providerType)
					if err != nil {
						t.Fatal(err)
					}
					if result.ProviderType != providerType || result.Configuration["certificate_password"] != "test-password" {
						t.Fatalf("Retry lost provider configuration: %#v", result)
					}
					if !reflect.DeepEqual(times, []time.Duration{0, time.Second, 3 * time.Second}) {
						t.Fatalf("Unexpected retry times: %v", times)
					}
					for _, body := range bodies[1:] {
						if string(body) != string(bodies[0]) {
							t.Fatal("Retry changed request payload")
						}
					}
				})
			})
		}
	}
}

func TestModelProviderRetryTimeoutAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name           string
		deadline       time.Duration
		callerDeadline bool
		cancelAfter    time.Duration
		wantError      error
		inFlight       bool
	}{
		{name: "readiness timeout", deadline: 10 * time.Minute, wantError: context.DeadlineExceeded},
		{name: "caller deadline", deadline: 1500 * time.Millisecond, callerDeadline: true, wantError: context.DeadlineExceeded},
		{name: "caller cancellation", cancelAfter: 500 * time.Millisecond, wantError: context.Canceled},
		{name: "deadline during request", deadline: 1500 * time.Millisecond, callerDeadline: true, wantError: context.DeadlineExceeded, inFlight: true},
		{name: "cancellation during request", cancelAfter: 1500 * time.Millisecond, wantError: context.Canceled, inFlight: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				started := time.Now()
				var attempts []time.Duration
				client := newModelProviderRetryClient(t, func(req *http.Request) (*http.Response, error) {
					attempts = append(attempts, time.Since(started))
					if tc.inFlight && len(attempts) == 2 {
						<-req.Context().Done()
						return nil, req.Context().Err()
					}
					return modelProviderResponse(req, http.StatusBadRequest, readinessErrorBody(certificateNotReadable)), nil
				})
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				if tc.cancelAfter > 0 {
					go func() { time.Sleep(tc.cancelAfter); cancel() }()
				} else if tc.callerDeadline {
					var cancelDeadline context.CancelFunc
					ctx, cancelDeadline = context.WithTimeout(ctx, tc.deadline)
					defer cancelDeadline()
				}
				_, err := mutateModelProvider(ctx, client, http.MethodPut, "azure_speech_service_principal_certificate")
				if !errors.Is(err, tc.wantError) {
					t.Fatalf("Expected %v, got %v", tc.wantError, err)
				}
				var apiErr *APIError
				if !errors.As(err, &apiErr) || string(apiErr.Body) != readinessErrorBody(certificateNotReadable) {
					t.Fatalf("Last readiness error lost: %v", err)
				}
				elapsed := tc.deadline
				if tc.cancelAfter > 0 {
					elapsed = tc.cancelAfter
				}
				if time.Since(started) != elapsed {
					t.Fatalf("Operation stopped after %v, expected %v", time.Since(started), elapsed)
				}
				for i, attempt := range attempts {
					if attempt >= elapsed {
						t.Fatal("Request made after deadline")
					}
					if i > 0 && attempt-attempts[i-1] > 10*time.Second {
						t.Fatal("Backoff exceeded ten seconds")
					}
				}
			})
		})
	}
}

func TestModelProviderDoesNotRetryOtherFailures(t *testing.T) {
	for _, tc := range []struct {
		name           string
		providerType   string
		status         int
		body           string
		transportError bool
	}{
		{"API key", "azure", 400, readinessErrorBody(mountNotConfigured), false},
		{"client secret", "azure_service_principal", 400, readinessErrorBody(mountNotConfigured), false},
		{"non-Azure certificate suffix", "other_certificate", 400, readinessErrorBody(mountNotConfigured), false},
		{"bad password", "azure_service_principal_certificate", 400, readinessErrorBody("Unable to load the mounted PEM certificate and private key. Check the file format and certificate password."), false},
		{"bad configuration", "azure_service_principal_certificate", 400, readinessErrorBody("Invalid tenant_id"), false},
		{"unauthorized", "azure_service_principal_certificate", 401, readinessErrorBody(mountNotConfigured), false},
		{"forbidden", "azure_service_principal_certificate", 403, readinessErrorBody(mountNotConfigured), false},
		{"not found", "azure_service_principal_certificate", 404, readinessErrorBody(mountNotConfigured), false},
		{"server failure", "azure_service_principal_certificate", 500, readinessErrorBody(mountNotConfigured), false},
		{"unstructured response", "azure_service_principal_certificate", 400, mountNotConfigured, false},
		{"nested detail", "azure_service_principal_certificate", 400, `{"detail":[{"msg":"Certificate authentication is not configured."}]}`, false},
		{"unknown detail", "azure_service_principal_certificate", 400, readinessErrorBody("Certificate authentication is not configured. Unknown reason."), false},
		{"uncertain transport failure", "azure_service_principal_certificate", 0, "", true},
	} {
		for _, method := range []string{http.MethodPost, http.MethodPut} {
			t.Run(tc.name+"/"+method, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					attempts := 0
					started := time.Now()
					client := newModelProviderRetryClient(t, func(req *http.Request) (*http.Response, error) {
						attempts++
						if tc.transportError {
							return nil, fmt.Errorf("connection closed after request")
						}
						return modelProviderResponse(req, tc.status, tc.body), nil
					})
					_, err := mutateModelProvider(t.Context(), client, method, tc.providerType)
					if err == nil || attempts != 1 || time.Since(started) != 0 {
						t.Fatalf("Unexpected retry: attempts=%d, elapsed=%v, error=%v", attempts, time.Since(started), err)
					}
				})
			})
		}
	}
}
