package xmlstrict

import "testing"

func TestCleanDocumentPasses(t *testing.T) {
	doc := `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>` +
		`<tev:R xmlns:tev="http://www.onvif.org/ver10/events/wsdl">` +
		`<wsnt:A xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">1</wsnt:A>` +
		`<wsnt:B xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">2</wsnt:B>` +
		`</tev:R></s:Body></s:Envelope>`
	if err := Check([]byte(doc)); err != nil {
		t.Fatalf("clean document flagged: %v", err)
	}
}

func TestUnboundSiblingPrefixReported(t *testing.T) {
	doc := `<tev:R xmlns:tev="http://www.onvif.org/ver10/events/wsdl">` +
		`<wsnt:A xmlns:wsnt="http://docs.oasis-open.org/wsn/b-2">1</wsnt:A>` +
		`<wsnt:B>2</wsnt:B></tev:R>`
	if err := Check([]byte(doc)); err == nil {
		t.Fatal("unbound sibling prefix must be reported")
	}
}
