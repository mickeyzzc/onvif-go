package soap

import (
	"encoding/xml"
	"testing"
)

// Fuzz targets pin the "hostile device response never panics" invariant for
// the client-side SOAP decode path (issue #61). They replay the exact
// sequence Client.Call applies to a response body: envelope body extraction,
// fault detection, response unmarshal. The seed corpus runs on every normal
// `go test`; extended fuzzing is `go test -fuzz`.

// fuzzResponse is a representative decode target: a plain response struct
// with the field kinds the generated types use (string + int).
type fuzzResponse struct {
	Uri  string `xml:"Uri"`
	TTL  int    `xml:"TTL"`
	Body string `xml:",innerxml"`
}

func FuzzClientDecodePath(f *testing.F) {
	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?><s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl"><trt:MediaUri><tt:Uri xmlns:tt="http://www.onvif.org/ver10/schema">rtsp://192.168.63.118:8554/live</tt:Uri><tt:InvalidAfterConnect>false</tt:InvalidAfterConnect></trt:MediaUri></trt:GetStreamUriResponse></s:Body></s:Envelope>`))
	f.Add([]byte(`<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body><s:Fault><s:Code><s:Value>s:Sender</s:Value></s:Code><s:Reason><s:Text xml:lang="en">ter:ActionNotSupported</s:Text></s:Reason></s:Fault></s:Body></s:Envelope>`))
	f.Add([]byte(`<soapenv:Fault xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"><faultcode>soapenv:Client</faultcode><faultstring>bad request</faultstring></soapenv:Fault>`))
	f.Add([]byte("<Envelope><Body>"))
	f.Add([]byte{0x00, 0xff, 0xfe})
	f.Fuzz(func(t *testing.T, respBody []byte) {
		// Mirrors Client.Call's post-read pipeline.
		var envelopeResp struct {
			Body struct {
				Content []byte `xml:",innerxml"`
			} `xml:"Body"`
		}
		if err := xml.Unmarshal(respBody, &envelopeResp); err != nil {
			return
		}
		fault := parseFault(envelopeResp.Body.Content, 200)
		if fault == nil {
			var response fuzzResponse
			_ = xml.Unmarshal(envelopeResp.Body.Content, &response)
		}
	})
}
