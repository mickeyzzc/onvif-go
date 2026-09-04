package main

import "testing"

// The simulator must not ship a default password: an unset flag has to be
// an error (empty credentials would leave every action open), the env var
// is the fallback, and an explicit flag always wins.
func TestResolvePassword(t *testing.T) {
	t.Setenv("ONVIF_SERVER_PASSWORD", "")
	if _, err := resolvePassword(""); err == nil {
		t.Fatal("expected error for empty flag and empty env, got nil")
	}

	t.Setenv("ONVIF_SERVER_PASSWORD", "from-env")
	got, err := resolvePassword("")
	if err != nil || got != "from-env" {
		t.Fatalf("flag empty, env set: got %q, err %v; want from-env", got, err)
	}

	got, err = resolvePassword("from-flag")
	if err != nil || got != "from-flag" {
		t.Fatalf("flag set: got %q, err %v; want from-flag (flag wins over env)", got, err)
	}
}
