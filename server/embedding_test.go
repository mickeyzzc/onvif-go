package server

// Tests for the embedding gaps filed from the MiBee Eye (rpi3b-cam)
// migration: per-stream RTSP port (#34), GetScopes (#37), and the
// snapshot-chain adaptations (#36).

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// staticStreamProvider pins one StreamInfo for every token.
type staticStreamProvider struct {
	info provider.StreamInfo
}

func (p staticStreamProvider) Stream(string) (provider.StreamInfo, error) {
	return p.info, nil
}

// TestGetStreamUriRTSPPort pins #34: the RTSP port comes from
// StreamInfo.RTSPPort (0 → the 8554 default); OverrideURI stays verbatim.
func TestGetStreamUriRTSPPort(t *testing.T) {
	config := createTestConfig()
	body := []byte(`<GetStreamURI><ProfileToken>profile_token_1</ProfileToken></GetStreamURI>`)

	t.Run("custom port", func(t *testing.T) {
		srv, err := New(config, WithStreamURIProvider(staticStreamProvider{
			info: provider.StreamInfo{RTSPPath: "/main", RTSPPort: 10554},
		}))
		if err != nil {
			t.Fatal(err)
		}

		resp, err := srv.HandleGetStreamUri(&soap.RequestContext{RemoteIP: "192.0.2.9"}, body)
		if err != nil {
			t.Fatal(err)
		}

		uri := resp.(*GetStreamUriResponse).MediaUri.URI
		if !strings.Contains(uri, "rtsp://192.0.2.9:10554/main") {
			t.Fatalf("URI = %q, want rtsp://192.0.2.9:10554/main", uri)
		}
	})

	t.Run("zero port falls back to 8554", func(t *testing.T) {
		srv, err := New(config, WithStreamURIProvider(staticStreamProvider{
			info: provider.StreamInfo{RTSPPath: "/main"},
		}))
		if err != nil {
			t.Fatal(err)
		}

		resp, err := srv.HandleGetStreamUri(&soap.RequestContext{RemoteIP: "192.0.2.9"}, body)
		if err != nil {
			t.Fatal(err)
		}

		uri := resp.(*GetStreamUriResponse).MediaUri.URI
		if !strings.Contains(uri, "rtsp://192.0.2.9:8554/main") {
			t.Fatalf("URI = %q, want rtsp://192.0.2.9:8554/main", uri)
		}
	})

	t.Run("override stays verbatim", func(t *testing.T) {
		srv, err := New(config, WithStreamURIProvider(staticStreamProvider{
			info: provider.StreamInfo{OverrideURI: "rtsp://elsewhere:1234/x", RTSPPort: 10554},
		}))
		if err != nil {
			t.Fatal(err)
		}

		resp, err := srv.HandleGetStreamUri(&soap.RequestContext{RemoteIP: "192.0.2.9"}, body)
		if err != nil {
			t.Fatal(err)
		}

		if uri := resp.(*GetStreamUriResponse).MediaUri.URI; uri != "rtsp://elsewhere:1234/x" {
			t.Fatalf("URI = %q, want verbatim override", uri)
		}
	})
}

// TestGetScopes pins #37: SOAP GetScopes returns the configured scope
// list; an empty config serves the conventional ONVIF default set.
func TestGetScopes(t *testing.T) {
	config := createTestConfig()
	config.Scopes = []string{
		"onvif://www.onvif.org/type/video_encoder",
		"onvif://www.onvif.org/name/MiBeeEye",
	}

	srv, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := srv.HandleGetScopes(nil, []byte(`<GetScopes/>`))
	if err != nil {
		t.Fatal(err)
	}

	scopesResp, ok := resp.(*GetScopesResponse)
	if !ok {
		t.Fatalf("response type = %T, want *GetScopesResponse", resp)
	}

	if len(scopesResp.Scopes) != 2 {
		t.Fatalf("scope count = %d, want 2", len(scopesResp.Scopes))
	}

	for i, want := range config.Scopes {
		if got := scopesResp.Scopes[i].ScopeItem; got != want {
			t.Errorf("scope[%d] = %q, want %q", i, got, want)
		}
	}

	// The wire form carries the scope as the tt:ScopeDefinition Scopeitem
	// attribute, the shape ONVIF clients parse.
	data, err := xml.Marshal(scopesResp)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), `Scopeitem="onvif://www.onvif.org/name/MiBeeEye"`) {
		t.Fatalf("marshaled scopes missing Scopeitem attribute: %s", data)
	}
}

// TestGetScopesDefaultSet pins the default scope list for configs that
// do not configure any (discovery parity: same values a
// discovery.Responder advertises when Scopes is empty).
func TestGetScopesDefaultSet(t *testing.T) {
	srv, err := New(createTestConfig())
	if err != nil {
		t.Fatal(err)
	}

	resp, err := srv.HandleGetScopes(nil, []byte(`<GetScopes/>`))
	if err != nil {
		t.Fatal(err)
	}

	scopes := resp.(*GetScopesResponse).Scopes
	if len(scopes) == 0 {
		t.Fatal("default scope set is empty")
	}

	found := false

	for _, s := range scopes {
		if s.ScopeItem == "onvif://www.onvif.org/type/NetworkVideoTransmitter" {
			found = true
		}
	}

	if !found {
		t.Fatalf("default scopes missing NetworkVideoTransmitter: %+v", scopes)
	}
}

// contentTypeSnapshotProvider serves arbitrary bytes with a content type.
type contentTypeSnapshotProvider struct {
	result provider.SnapshotResult
}

func (p contentTypeSnapshotProvider) Snapshot(string) (provider.SnapshotResult, error) {
	return p.result, nil
}

// TestSnapshotEndpointAdaptations pins #36: the HTTP snapshot endpoint
// works without a ?profile= parameter (default profile), and serves the
// provider's content type instead of assuming JPEG.
func TestSnapshotEndpointAdaptations(t *testing.T) {
	config := createTestConfig()
	config.SnapshotPath = "/snap"
	config.SnapshotURIParameterless = true

	srv, err := New(config, WithSnapshotProvider(contentTypeSnapshotProvider{
		result: provider.SnapshotResult{Data: []byte("IDRBYTES"), ContentType: "video/H264"},
	}))
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	srv.handleSnapshot(rec, httptest.NewRequest(http.MethodGet, "/snap", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("parameterless snapshot: status = %d, want 200; body %s", rec.Code, rec.Body.String())
	}

	if got := rec.Header().Get("Content-Type"); got != "video/H264" {
		t.Fatalf("Content-Type = %q, want video/H264 (provider-supplied)", got)
	}

	if rec.Body.String() != "IDRBYTES" {
		t.Fatalf("body = %q, want provider bytes", rec.Body.String())
	}
}

// TestSnapshotJPEGDefaultContentType pins the JPEG default when the
// provider leaves ContentType empty.
func TestSnapshotJPEGDefaultContentType(t *testing.T) {
	srv, err := New(createTestConfig(), WithSnapshotProvider(contentTypeSnapshotProvider{
		result: provider.SnapshotResult{Data: []byte("JPEG")},
	}))
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	srv.handleSnapshot(rec, httptest.NewRequest(http.MethodGet, "/onvif/snapshot?profile=profile_token_1", nil))

	if got := rec.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Fatalf("Content-Type = %q, want image/jpeg default", got)
	}
}

// TestGetSnapshotUriForm pins #36: the advertised snapshot URI follows
// SnapshotPath and drops the ?profile= query when the parameterless
// form is configured (default: the historical query form).
func TestGetSnapshotUriForm(t *testing.T) {
	body := []byte(`<GetSnapshotURI><ProfileToken>profile_token_1</ProfileToken></GetSnapshotURI>`)
	rc := &soap.RequestContext{RemoteIP: "192.0.2.9"}

	t.Run("custom path without query", func(t *testing.T) {
		config := createTestConfig()
		config.SnapshotPath = "/snap"
		config.SnapshotURIParameterless = true

		srv, err := New(config)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := srv.HandleGetSnapshotUri(rc, body)
		if err != nil {
			t.Fatal(err)
		}

		if uri := resp.(*GetSnapshotUriResponse).MediaUri.URI; uri != "http://192.0.2.9:8080/snap" {
			t.Fatalf("URI = %q, want http://192.0.2.9:8080/snap", uri)
		}
	})

	t.Run("default keeps query form", func(t *testing.T) {
		srv, err := New(createTestConfig())
		if err != nil {
			t.Fatal(err)
		}

		resp, err := srv.HandleGetSnapshotUri(rc, body)
		if err != nil {
			t.Fatal(err)
		}

		want := "http://192.0.2.9:8080/onvif/snapshot?profile=profile_token_1"
		if uri := resp.(*GetSnapshotUriResponse).MediaUri.URI; uri != want {
			t.Fatalf("URI = %q, want %q", uri, want)
		}
	})
}
