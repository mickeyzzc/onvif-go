package onvif

// Coverage for the remaining client-core gaps: facade accessors, the
// AuthMode getter, and the download status error chain.

import (
	"errors"
	"testing"
)

func TestFacadeAccessors(t *testing.T) {
	client, err := NewClient("http://192.0.2.9:80/onvif/device_service")
	if err != nil {
		t.Fatal(err)
	}

	if client.PTZ() == nil {
		t.Error("PTZ() returned nil")
	}

	if client.Imaging() == nil {
		t.Error("Imaging() returned nil")
	}

	// Facades are long-lived: repeated calls return the same instance.
	first, second := client.PTZ(), client.PTZ()
	if first != second {
		t.Error("PTZ() must be a stable instance")
	}
}

func TestAuthModeGetter(t *testing.T) {
	client, err := NewClient("http://192.0.2.9:80/onvif/device_service")
	if err != nil {
		t.Fatal(err)
	}

	if got := client.AuthMode(); got != AuthDigest {
		t.Errorf("default AuthMode = %v, want AuthDigest", got)
	}
}

func TestDownloadStatusErrorChain(t *testing.T) {
	inner := errors.New("http 502 from snapshot endpoint")

	err := &downloadStatusError{status: 502, err: inner}
	if !errors.Is(err, inner) {
		t.Error("Unwrap chain broken")
	}

	if err.Error() != inner.Error() {
		t.Errorf("Error() = %q", err.Error())
	}
}
