package soap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// RawXML is a pre-built response body. Returning it from a handler embeds
// the bytes verbatim inside the SOAP envelope, bypassing encoding/xml —
// the byte-level output channel for hosts that need exact control over
// element names, prefixes, or formatting. Prefix modes never rewrite it.
type RawXML []byte

// namespacePrefixes maps ONVIF XML namespaces to the conventional wire
// prefixes used in explicit-prefix responses.
var namespacePrefixes = map[string]string{
	"http://www.onvif.org/ver10/device/wsdl":    "tds",
	"http://www.onvif.org/ver10/media/wsdl":     "trt",
	"http://www.onvif.org/ver10/events/wsdl":    "tev",
	"http://www.onvif.org/ver10/schema":         "tt",
	"http://www.onvif.org/ver10/common/wsdl":    "trc",
	"http://www.onvif.org/ver20/device/wsdl":    "tds",
	"http://www.onvif.org/ver20/media/wsdl":     "trt",
	"http://www.onvif.org/ver20/ptz/wsdl":       "tptz",
	"http://www.onvif.org/ver20/imaging/wsdl":   "timg",
	"http://www.onvif.org/ver20/analytics/wsdl": "tan",
}

// sendResponse marshals the handler result and writes the SOAP envelope.
func (h *Handler) sendResponse(w http.ResponseWriter, response interface{}) {
	content, err := h.marshalContent(response)
	if err != nil {
		h.sendFault(w, "Receiver", "Failed to marshal response", err.Error())

		return
	}

	writeEnvelope(w, http.StatusOK, content, h.explicitPrefixes)
}

// marshalContent serializes a handler result into body-content bytes.
func (h *Handler) marshalContent(response interface{}) ([]byte, error) {
	switch v := response.(type) {
	case nil:
		return nil, nil
	case RawXML:
		return v, nil
	}

	data, err := xml.MarshalIndent(response, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	if h.explicitPrefixes {
		data, err = withExplicitPrefixes(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

// writeEnvelope writes a complete SOAP 1.2 response envelope around
// pre-serialized body content. The byte layout is stable and locked by
// golden tests:
//
//	default:   <Envelope xmlns="…soap-envelope"> / <Body> / content
//	prefixed:  <s:Envelope xmlns:s="…"> / <s:Body> / prefixed content
func writeEnvelope(w http.ResponseWriter, status int, content []byte, prefixes bool) {
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.WriteHeader(status)

	var b bytes.Buffer
	b.WriteString(xml.Header)

	if prefixes {
		b.WriteString(`<s:Envelope xmlns:s="` + soapEnvelopeNS + `">`)
		b.WriteString("\n  <s:Body>")
	} else {
		b.WriteString(`<Envelope xmlns="` + soapEnvelopeNS + `">`)
		b.WriteString("\n  <Body>")
	}

	if len(content) > 0 {
		b.WriteString("\n")
		indentLines(&b, content, "    ")
		b.WriteString("\n  ")
	}

	if prefixes {
		b.WriteString("</s:Body>\n</s:Envelope>")
	} else {
		b.WriteString("</Body>\n</Envelope>")
	}

	_, _ = w.Write(b.Bytes())
}

// indentLines writes data with every line prefixed by indent.
func indentLines(b *bytes.Buffer, data []byte, indent string) {
	for i, line := range bytes.Split(data, []byte("\n")) {
		if i > 0 {
			b.WriteString("\n")
		}

		if len(line) > 0 {
			b.WriteString(indent)
		}

		b.Write(line)
	}
}

// sendFault sends a SOAP fault response.
func (h *Handler) sendFault(w http.ResponseWriter, code, reason, detail string) {
	content := faultContent(code, reason, detail, h.explicitPrefixes)

	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	statusCode := http.StatusInternalServerError
	if code == "Sender" {
		statusCode = http.StatusBadRequest
	}

	writeEnvelope(w, statusCode, content, h.explicitPrefixes)
}

// faultContent builds the SOAP Fault body content. Layout mirrors the
// encoding/xml form the library has always produced (Code>Value,
// Reason>Text, optional Detail).
func faultContent(code, reason, detail string, prefixes bool) []byte {
	prefix := ""
	if prefixes {
		prefix = "s:"
	}

	var b bytes.Buffer
	b.WriteString("<" + prefix + "Fault>\n")
	b.WriteString("  <" + prefix + "Code>\n")
	b.WriteString("    <" + prefix + "Value>" + escapeXMLText(code) + "</" + prefix + "Value>\n")
	b.WriteString("  </" + prefix + "Code>\n")
	b.WriteString("  <" + prefix + "Reason>\n")
	b.WriteString("    <" + prefix + "Text>" + escapeXMLText(reason) + "</" + prefix + "Text>\n")
	b.WriteString("  </" + prefix + "Reason>\n")

	if detail != "" {
		b.WriteString("  <" + prefix + "Detail>" + escapeXMLText(detail) + "</" + prefix + "Detail>\n")
	}

	b.WriteString("</" + prefix + "Fault>")

	return b.Bytes()
}

// escapeXMLText escapes a text node value.
func escapeXMLText(s string) string {
	var b strings.Builder
	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		return ""
	}

	return b.String()
}

// withExplicitPrefixes rewrites marshaled XML so that every element in a
// known ONVIF namespace carries its conventional prefix (tds:, trt:, ...)
// with the binding declared on first use. Namespaces without a
// conventional prefix keep a default xmlns declaration.
//
// encoding/xml cannot emit prefixes natively (golang/go#14211); the
// prefix is placed in the element's local name, which the encoder writes
// verbatim.
func withExplicitPrefixes(data []byte) ([]byte, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")

	declared := make(map[string]bool)

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("parse marshaled body: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			t.Attr = stripNamespaceDecls(t.Attr)

			if prefix, ok := namespacePrefixes[t.Name.Space]; ok {
				namespace := t.Name.Space
				t.Name.Local = prefix + ":" + t.Name.Local
				t.Name.Space = ""

				if !declared[prefix] {
					t.Attr = append(t.Attr, xml.Attr{
						Name:  xml.Name{Local: "xmlns:" + prefix},
						Value: namespace,
					})
					declared[prefix] = true
				}
			} else if t.Name.Space != "" {
				// Unknown namespace: keep correctness with a local default
				// declaration.
				namespace := t.Name.Space
				t.Name.Space = ""
				t.Attr = append(t.Attr, xml.Attr{
					Name:  xml.Name{Local: "xmlns"},
					Value: namespace,
				})
			}

			token = t
		case xml.EndElement:
			if prefix, ok := namespacePrefixes[t.Name.Space]; ok {
				t.Name.Local = prefix + ":" + t.Name.Local
			}

			t.Name.Space = ""
			token = t
		case xml.CharData:
			// Drop MarshalIndent's inter-element whitespace; the encoder
			// re-indents. Element text is preserved.
			if len(strings.TrimSpace(string(t))) == 0 {
				continue
			}
		}

		if err := encoder.EncodeToken(token); err != nil {
			return nil, fmt.Errorf("re-encode body token: %w", err)
		}
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("flush re-encoded body: %w", err)
	}

	return buf.Bytes(), nil
}

// stripNamespaceDecls removes xmlns/xmlns:p attributes from a decoded
// element; namespace rewrites re-declare what they need.
func stripNamespaceDecls(attrs []xml.Attr) []xml.Attr {
	kept := attrs[:0]

	for _, attr := range attrs {
		if attr.Name.Space == "xmlns" || attr.Name.Local == "xmlns" {
			continue
		}

		kept = append(kept, attr)
	}

	return kept
}
