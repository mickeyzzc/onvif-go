package onviftesting

// Tests for the parameter-aware V2 mock server and its helpers.

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

const v2Exchange = `{"version":"2.0","operation_name":"GetProfile","parameters":{"ProfileToken":"main"},
 "request_body":"<Body><GetProfile/></Body>","response_body":"<GetProfileResponse><main/></GetProfileResponse>","status_code":200}`

const v2SubExchange = `{"version":"2.0","operation_name":"GetProfile","parameters":{"ProfileToken":"sub"},
 "request_body":"<Body><GetProfile/></Body>","response_body":"<GetProfileResponse><sub/></GetProfileResponse>","status_code":200}`

func TestLoadCaptureFromArchiveV2(t *testing.T) {
	path := writeCaptureArchive(t, map[string]string{
		"metadata.json":   `{"version":"2.0","camera_info":{"manufacturer":"Acme","model":"CamX"},"total_exchanges":2}`,
		"001.json":        v2Exchange,
		"002.json":        v2SubExchange,
		"003_request.xml": "skipped",
	})

	capture, metadata, err := LoadCaptureFromArchiveV2(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(capture.Exchanges) != 2 {
		t.Fatalf("exchanges = %d, want 2 (_request members skipped)", len(capture.Exchanges))
	}

	if metadata == nil || metadata.CameraInfo.Manufacturer != "Acme" {
		t.Fatalf("metadata = %+v", metadata)
	}

	if capture.Metadata == nil || capture.Metadata.CameraInfo.Model != "CamX" {
		t.Errorf("capture.Metadata not attached: %+v", capture.Metadata)
	}
}

func TestLoadCaptureFromArchiveV2MixedV1(t *testing.T) {
	path := writeCaptureArchive(t, map[string]string{
		"001.json": `{"operation_name":"GetServices","request_body":"<Body><GetServices/></Body>","response_body":"<GetServicesResponse/>","status_code":200}`,
	})

	capture, metadata, err := LoadCaptureFromArchiveV2(path)
	if err != nil {
		t.Fatal(err)
	}

	if metadata != nil {
		t.Errorf("V1 archive must not produce metadata, got %+v", metadata)
	}

	if len(capture.Exchanges) != 1 {
		t.Fatalf("exchanges = %d", len(capture.Exchanges))
	}

	ex := capture.Exchanges[0]
	if ex.Parameters == nil || ex.Parameters["ProfileToken"] != nil && len(ex.Parameters) == 0 {
		// Parameters map exists (possibly empty for GetServices).
		t.Logf("parameters: %v", ex.Parameters)
	}
}

func TestMockSOAPServerV2ParameterMatching(t *testing.T) {
	path := writeCaptureArchive(t, map[string]string{
		"001.json": v2Exchange,
		"002.json": v2SubExchange,
	})

	mock, err := NewMockSOAPServerV2(path)
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	if mock.URL() == "" {
		t.Fatal("URL() empty")
	}

	if mock.GetExchangeCount() != 2 {
		t.Errorf("GetExchangeCount = %d", mock.GetExchangeCount())
	}

	ops := mock.GetOperations()
	if len(ops) != 1 || ops[0] != "GetProfile" {
		t.Errorf("GetOperations = %v", ops)
	}

	for _, tc := range []struct {
		token string
		want  string
	}{
		{"main", "<main/>"},
		{"sub", "<sub/>"},
	} {
		body := `<Body><GetProfile><ProfileToken>` + tc.token + `</ProfileToken></GetProfile></Body>`

		resp, err := http.Post(mock.URL(), "application/soap+xml", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}

		respBody, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if resp.StatusCode != http.StatusOK || !strings.Contains(string(respBody), tc.want) {
			t.Errorf("token %s: status %d body %s", tc.token, resp.StatusCode, respBody)
		}
	}
}

func TestMockSOAPServerV2FallbackAndMiss(t *testing.T) {
	path := writeCaptureArchive(t, map[string]string{
		"001.json": v2Exchange,
	})

	mock, err := NewMockSOAPServerV2(path)
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	// Unknown token still falls back to the single captured exchange.
	body := `<Body><GetProfile><ProfileToken>whatever</ProfileToken></GetProfile></Body>`

	resp, err := http.Post(mock.URL(), "application/soap+xml", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	respBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK || !strings.Contains(string(respBody), "<main/>") {
		t.Errorf("fallback: status %d body %s", resp.StatusCode, respBody)
	}

	// Uncaptured operation → 404.
	resp2, err := http.Post(mock.URL(), "application/soap+xml", strings.NewReader(`<Body><GetServices/></Body>`))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp2.Body.Close()

	if resp2.StatusCode != http.StatusNotFound {
		t.Errorf("uncaptured op status = %d, want 404", resp2.StatusCode)
	}

	// Unparseable body → 400.
	resp3, err := http.Post(mock.URL(), "application/soap+xml", strings.NewReader("garbage"))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp3.Body.Close()

	if resp3.StatusCode != http.StatusBadRequest {
		t.Errorf("garbage status = %d, want 400", resp3.StatusCode)
	}
}

func TestMockSOAPServerV2MetadataAccessor(t *testing.T) {
	path := writeCaptureArchive(t, map[string]string{
		"metadata.json": `{"version":"2.0","camera_info":{"manufacturer":"Acme"},"total_exchanges":1}`,
		"001.json":      v2Exchange,
	})

	mock, err := NewMockSOAPServerV2(path)
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	if mock.Metadata() == nil || mock.Metadata().CameraInfo.Manufacturer != "Acme" {
		t.Fatalf("Metadata() = %+v", mock.Metadata())
	}
}

func TestExtractParameters(t *testing.T) {
	body := `<Body><GetProfile><ProfileToken>main-1</ProfileToken></GetProfile></Body>`

	params := ExtractParameters("GetProfile", body)
	if params["ProfileToken"] != "main-1" {
		t.Errorf("params = %v", params)
	}

	empty := ExtractParameters("GetProfile", "<Body><GetProfile/></Body>")
	if len(empty) != 0 {
		t.Errorf("no-token body params = %v, want empty", empty)
	}
}

func TestExtractXMLElement(t *testing.T) {
	direct := ExtractXMLElement("<a><Manufacturer>Acme</Manufacturer></a>", "Manufacturer")
	if direct != "Acme" {
		t.Errorf("direct = %q", direct)
	}

	prefixed := ExtractXMLElement("<tds:GetDeviceInformationResponse><tds:Manufacturer>Acme</tds:Manufacturer></tds:GetDeviceInformationResponse>", "Manufacturer")
	if prefixed != "Acme" {
		t.Errorf("prefixed = %q", prefixed)
	}

	if got := ExtractXMLElement("<a/>", "Missing"); got != "" {
		t.Errorf("missing element = %q", got)
	}
}

func TestFaultHelpers(t *testing.T) {
	fault := GenerateFaultResponse(SOAPFault{Code: "ter:NotAuthorized", Reason: "denied", Detail: "extra"})
	if !IsFaultResponse(fault) {
		t.Error("generated fault not detected")
	}

	if !strings.Contains(fault, "ter:NotAuthorized") || !strings.Contains(fault, "<soap:Detail>extra</soap:Detail>") {
		t.Errorf("fault body malformed: %s", fault)
	}

	if IsFaultResponse("<Body><GetProfiles/></Body>") {
		t.Error("clean response misclassified as fault")
	}

	// Detail-less fault omits the Detail element.
	bare := GenerateFaultResponse(SOAPFault{Code: "ter:Receiver", Reason: "boom"})
	if strings.Contains(bare, "<soap:Detail>") {
		t.Errorf("bare fault has Detail: %s", bare)
	}
}
