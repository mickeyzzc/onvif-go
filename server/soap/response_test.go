package soap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/xmlstrict"
)

// goldenStreamUri mirrors the server's GetStreamUriResponse wire shape.
type goldenStreamUri struct {
	XMLName  xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetStreamUriResponse"`
	MediaUri struct {
		URI string `xml:"Uri"`
	} `xml:"MediaUri"`
}

func newGoldenStreamUri() *goldenStreamUri {
	var resp goldenStreamUri
	resp.MediaUri.URI = "rtsp://198.51.100.7:8554/stream0"

	return &resp
}

func TestGoldenEnvelopeDefaultForm(t *testing.T) {
	h := NewHandler("", "")
	h.RegisterContextHandler("GetStreamUri", func(_ *RequestContext, _ []byte) (interface{}, error) {
		return newGoldenStreamUri(), nil
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><GetStreamUri xmlns="http://www.onvif.org/ver10/media/wsdl"/></s:Body></s:Envelope>`))
	req.RemoteAddr = "192.0.2.10:1234"

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	want := `<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
  <Body>
    <GetStreamUriResponse xmlns="http://www.onvif.org/ver10/media/wsdl">
      <MediaUri>
        <Uri>rtsp://198.51.100.7:8554/stream0</Uri>
      </MediaUri>
    </GetStreamUriResponse>
  </Body>
</Envelope>`

	if got := w.Body.String(); got != want {
		t.Errorf("golden default envelope mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestGoldenEnvelopeExplicitPrefixes(t *testing.T) {
	h := NewHandlerWithOptions(HandlerOptions{ExplicitPrefixes: true})
	h.RegisterContextHandler("GetStreamUri", func(_ *RequestContext, _ []byte) (interface{}, error) {
		return newGoldenStreamUri(), nil
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><GetStreamUri xmlns="http://www.onvif.org/ver10/media/wsdl"/></s:Body></s:Envelope>`))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	want := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
  <s:Body>
    <trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:MediaUri>
        <trt:Uri>rtsp://198.51.100.7:8554/stream0</trt:Uri>
      </trt:MediaUri>
    </trt:GetStreamUriResponse>
  </s:Body>
</s:Envelope>`

	if got := w.Body.String(); got != want {
		t.Errorf("golden prefixed envelope mismatch\n got: %q\nwant: %q", got, want)
	}
	if err := xmlstrict.Check(w.Body.Bytes()); err != nil {
		t.Errorf("golden prefixed envelope not strictly parseable: %v", err)
	}
}

func TestGoldenRawXMLPassthrough(t *testing.T) {
	raw := `<trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl"><trt:MediaUri><trt:Uri>rtsp://raw/stream</trt:Uri></trt:MediaUri></trt:GetStreamUriResponse>`

	// Both encoding modes must embed the bytes verbatim.
	for _, prefixes := range []bool{false, true} {
		h := NewHandlerWithOptions(HandlerOptions{ExplicitPrefixes: prefixes})
		h.RegisterContextHandler("GetStreamUri", func(_ *RequestContext, _ []byte) (interface{}, error) {
			return RawXML(raw), nil
		})

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
			`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><GetStreamUri xmlns="http://www.onvif.org/ver10/media/wsdl"/></s:Body></s:Envelope>`))

		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		if !strings.Contains(w.Body.String(), raw) {
			t.Errorf("prefixes=%v: raw bytes not embedded verbatim: %s", prefixes, w.Body.String())
		}
	}
}

func TestGoldenFaultDefaultForm(t *testing.T) {
	h := NewHandler("admin", "secret")
	h.RegisterContextHandler("SetScopes", func(_ *RequestContext, _ []byte) (interface{}, error) {
		return RawXML("<unreachable/>"), nil
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><SetScopes xmlns="http://www.onvif.org/ver10/device/wsdl"/></s:Body></s:Envelope>`))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	want := `<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
  <Body>
    <Fault>
      <Code>
        <Value>Sender</Value>
      </Code>
      <Reason>
        <Text>Sender not authorized</Text>
      </Reason>
      <Detail>Invalid username or password</Detail>
    </Fault>
  </Body>
</Envelope>`

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}

	if got := w.Body.String(); got != want {
		t.Errorf("golden fault mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestFaultEscapesDetail(t *testing.T) {
	h := NewHandler("", "")
	h.RegisterContextHandler("Boom", func(_ *RequestContext, _ []byte) (interface{}, error) {
		return nil, errGolden //nolint:wrapcheck // exercising escaping through the transport
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><Boom/></s:Body></s:Envelope>`))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "a&lt;b &amp; c") {
		t.Errorf("fault detail not escaped as expected: %s", body)
	}
}

// errGolden carries characters that must be XML-escaped inside Detail.
var errGolden = &goldenError{}

type goldenError struct{}

func (*goldenError) Error() string { return "a<b & c" }

func TestNilResponseYieldsEmptyBody(t *testing.T) {
	h := NewHandler("", "")
	h.RegisterContextHandler("Nop", func(_ *RequestContext, _ []byte) (interface{}, error) {
		return nil, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><Nop/></s:Body></s:Envelope>`))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	want := `<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://www.w3.org/2003/05/soap-envelope">
  <Body></Body>
</Envelope>`

	if got := w.Body.String(); got != want {
		t.Errorf("nil response envelope mismatch\n got: %q\nwant: %q", got, want)
	}
}

// prefixURIs inverts namespacePrefixes: decoding with encoding/xml
// resolves a BOUND prefix to its URI, but leaves an UNBOUND prefix as
// the literal Space — so an element whose Space equals a conventional
// wire prefix is a well-formedness violation strict parsers reject.
func conventionalPrefixes() map[string]bool {
	set := make(map[string]bool)
	for _, prefix := range namespacePrefixes {
		set[prefix] = true
	}
	return set
}

// assertNoUnboundPrefixes walks the serialized form and fails when any
// element name uses a prefix without an in-scope xmlns declaration.
func assertNoUnboundPrefixes(t *testing.T, data []byte) {
	t.Helper()
	prefixes := conventionalPrefixes()
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("parse serialized form: %v", err)
		}
		if el, ok := tok.(xml.StartElement); ok {
			if prefixes[el.Name.Space] {
				t.Fatalf("element <%s:%s> uses prefix %q with no in-scope xmlns declaration (strict parsers reject the whole document):\n%s",
					el.Name.Space, el.Name.Local, el.Name.Space, data)
			}
		}
	}
}

// Sibling elements sharing a namespace each need the declaration in
// scope: a declaration rides on ONE element and siblings do not inherit
// it. Caught live on rpi3b-cam (.118): wsnt:TerminationTime next to a
// self-declaring wsnt:CurrentTime produced an unbound prefix that expat
// (and every strict stack) rejects.
func TestWithExplicitPrefixesSiblingScopes(t *testing.T) {
	type siblingPair struct {
		XMLName         xml.Name `xml:"http://www.onvif.org/ver10/events/wsdl CreatePullPointSubscriptionResponse"`
		CurrentTime     string   `xml:"http://docs.oasis-open.org/wsn/b-2 CurrentTime"`
		TerminationTime string   `xml:"http://docs.oasis-open.org/wsn/b-2 TerminationTime"`
	}
	raw, err := xml.MarshalIndent(siblingPair{
		CurrentTime:     "2026-09-21T02:08:56Z",
		TerminationTime: "2026-09-21T02:09:56Z",
	}, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out, err := withExplicitPrefixes(raw)
	if err != nil {
		t.Fatalf("withExplicitPrefixes: %v", err)
	}
	assertNoUnboundPrefixes(t, out)
	// Both siblings must carry their own declaration bytes.
	if got := strings.Count(string(out), `xmlns:wsnt=`); got != 2 {
		t.Fatalf("xmlns:wsnt declarations = %d, want 2 (one per sibling):\n%s", got, out)
	}
}
