package soap

// Over-the-wire metrics semantics (issue #66), mirroring the onvif-rs
// MetricsHooks contract: SoapRequest counts dispatched requests only,
// SoapFault counts handler errors (the fault-rate numerator), AuthFail
// counts failed UsernameTokens, AuthLockout counts lockout refusals.

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/metrics"
)

type countingHooks struct {
	metrics.NoopHooks // DiscoveryProbeAnswered never fires in the SOAP layer

	mu        sync.Mutex
	requests  []string
	faults    []string
	authFails int
	lockouts  int
}

func (c *countingHooks) SoapRequest(action string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requests = append(c.requests, action)
}

func (c *countingHooks) SoapFault(action string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.faults = append(c.faults, action)
}

func (c *countingHooks) AuthFail()    { c.mu.Lock(); c.authFails++; c.mu.Unlock() }
func (c *countingHooks) AuthLockout() { c.mu.Lock(); c.lockouts++; c.mu.Unlock() }

func (c *countingHooks) snapshot() (requests, faults []string, authFails, lockouts int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.requests, c.faults, c.authFails, c.lockouts
}

func newMetricsTestHandler(t *testing.T, hooks metrics.Hooks) *Handler {
	t.Helper()

	h := NewHandlerWithOptions(HandlerOptions{
		Username: "admin",
		Password: "secret",
		Metrics:  hooks,
		// Small limit so the lockout case fits in a tight loop.
		AuthFailureLimit: 2,
		AuthLockout:      0,
	})
	return h
}

func TestMetricsDispatchedRequestCountsByAction(t *testing.T) {
	hooks := &countingHooks{}
	h := newMetricsTestHandler(t, hooks)
	h.RegisterHandler("GetProfiles", func([]byte) (interface{}, error) {
		return "ok", nil
	})

	req := buildAuthedRequest(t, "GetProfiles", "admin", "secret", "digest")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	requests, faults, authFails, lockouts := hooks.snapshot()
	if len(requests) != 1 || requests[0] != "GetProfiles" {
		t.Errorf("requests = %v, want [GetProfiles]", requests)
	}
	if len(faults) != 0 || authFails != 0 || lockouts != 0 {
		t.Errorf("faults=%v authFails=%d lockouts=%d, want zeros", faults, authFails, lockouts)
	}
}

func TestMetricsHandlerErrorCountsSoapFault(t *testing.T) {
	hooks := &countingHooks{}
	h := newMetricsTestHandler(t, hooks)
	h.RegisterHandler("SetUsers", func([]byte) (interface{}, error) {
		return nil, errors.New("boom")
	})

	req := buildAuthedRequest(t, "SetUsers", "admin", "secret", "text")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	requests, faults, _, _ := hooks.snapshot()
	if len(requests) != 1 || requests[0] != "SetUsers" {
		t.Errorf("requests = %v, want [SetUsers]", requests)
	}
	if len(faults) != 1 || faults[0] != "SetUsers" {
		t.Errorf("faults = %v, want [SetUsers]", faults)
	}
}

func TestMetricsBadCredentialsCountAuthFailOnly(t *testing.T) {
	hooks := &countingHooks{}
	h := newMetricsTestHandler(t, hooks)
	h.RegisterHandler("SetUsers", func([]byte) (interface{}, error) {
		return nil, nil
	})

	req := buildAuthedRequest(t, "SetUsers", "admin", "wrong", "text")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}

	requests, faults, authFails, lockouts := hooks.snapshot()
	if authFails != 1 {
		t.Errorf("authFails = %d, want 1", authFails)
	}
	if len(requests) != 0 || len(faults) != 0 || lockouts != 0 {
		t.Errorf("requests=%v faults=%v lockouts=%d, want zeros — undelivered requests are not dispatch metrics", requests, faults, lockouts)
	}
}

func TestMetricsLockoutRefusalCountsAuthLockout(t *testing.T) {
	hooks := &countingHooks{}
	h := newMetricsTestHandler(t, hooks) // AuthFailureLimit = 2
	h.RegisterHandler("SetUsers", func([]byte) (interface{}, error) {
		return nil, nil
	})

	// Trip the limit, then one more request arrives while locked out.
	for i := range 3 {
		req := buildAuthedRequest(t, "SetUsers", "admin", "wrong", "text")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("request %d: status = %d, want 400", i, rec.Code)
		}
	}

	_, _, authFails, lockouts := hooks.snapshot()
	if authFails != 2 {
		t.Errorf("authFails = %d, want 2 (the third is refused, not authenticated)", authFails)
	}
	if lockouts != 1 {
		t.Errorf("lockouts = %d, want 1", lockouts)
	}
}

func TestMetricsNonPostCountsNothing(t *testing.T) {
	hooks := &countingHooks{}
	h := newMetricsTestHandler(t, hooks)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", http.NoBody))

	requests, faults, authFails, lockouts := hooks.snapshot()
	if len(requests) != 0 || len(faults) != 0 || authFails != 0 || lockouts != 0 {
		t.Errorf("hooks fired for a non-POST: requests=%v faults=%v authFails=%d lockouts=%d",
			requests, faults, authFails, lockouts)
	}
}
