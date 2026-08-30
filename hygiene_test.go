// Package hygiene_test holds regression guards for the library-hygiene
// rules established for v2: no default credentials, no private network data
// committed under testdata, and no direct stdout printing in library code.
package hygiene_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/server"
)

// TestDefaultConfigShipsWithoutCredentials pins the credential default:
// DefaultConfig must NOT ship a guessable username/password pair. Without
// credentials the server runs in its documented everything-open mode and
// embedders set their own pair explicitly.
func TestDefaultConfigShipsWithoutCredentials(t *testing.T) {
	cfg := server.DefaultConfig()
	if cfg.Username != "" || cfg.Password != "" {
		t.Fatalf("DefaultConfig credentials must be empty, got %q/%q — a library default must not ship guessable credentials", cfg.Username, cfg.Password)
	}
}

// libraryDirs are the importable packages of the module.
var libraryDirs = []string{
	"device", "deviceio", "discovery", "events", "imaging", "internal",
	"media", "onvif", "ptz", "security", "server", "types", "wsdiscovery",
}

// TestNoDefaultCredentialsInLibraryCode scans non-test library sources for
// credential-looking defaults.
func TestNoDefaultCredentialsInLibraryCode(t *testing.T) {
	for _, dir := range libraryDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			src := string(data)
			for _, banned := range []string{`Username: "admin"`, `Password: "admin"`, `"admin/admin"`} {
				if strings.Contains(src, banned) {
					t.Errorf("%s contains default credential literal %s — credentials come from host configuration", path, banned)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}
}

// TestTestdataContainsNoPrivateNetworkData enforces the testdata policy:
// committed fixtures must use RFC 5737 documentation ranges only, never
// real LAN addresses (a 2026-01 network-discovery dump with live IPs was
// committed once — never again).
func TestTestdataContainsNoPrivateNetworkData(t *testing.T) {
	err := filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		src := string(data)
		for _, line := range strings.Split(src, "\n") {
			// Private ranges that indicate real network data.
			for _, prefix := range []string{"192.168.", "10.0.", "172.16.", "172.17.", "172.18.", "172.19.", "172.20.", "172.21.", "172.22.", "172.23.", "172.24.", "172.25.", "172.26.", "172.27.", "172.28.", "172.29.", "172.30.", "172.31."} {
				if strings.Contains(line, prefix) {
					t.Errorf("%s contains private address %q — testdata must use RFC 5737 documentation ranges (192.0.2.x / 198.51.100.x); keep real captures in gitignored tmp/", path, strings.TrimSpace(line))
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk testdata: %v", err)
	}
}
