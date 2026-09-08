package metrics

import (
	"reflect"
	"testing"
)

func TestNoopHooksImplementsHooks(t *testing.T) {
	// Compile-time interface compliance plus runtime sanity: every hook
	// method is callable and returns without side effects.
	var h Hooks = NoopHooks{}
	h.SoapRequest("GetProfiles")
	h.SoapFault("GetProfiles")
	h.AuthFail()
	h.AuthLockout()
	h.DiscoveryProbeAnswered()
}

func TestOrNoop(t *testing.T) {
	if got := OrNoop(nil); reflect.TypeOf(got) != reflect.TypeOf(NoopHooks{}) {
		t.Fatalf("OrNoop(nil) = %T, want NoopHooks", got)
	}

	custom := NoopHooks{}
	if got := OrNoop(custom); got != custom {
		t.Fatalf("OrNoop must pass through non-nil hooks, got %T", got)
	}
}
