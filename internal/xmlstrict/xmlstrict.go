// Package xmlstrict reports XML elements whose name prefix has no
// in-scope xmlns declaration — a well-formedness violation that Go's
// encoding/xml tolerates but strict stacks (expat, libxml, .NET)
// reject wholesale. Used by wire tests to keep serialized responses
// parseable by every real client.
package xmlstrict

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// conventionalPrefixes are the wire prefixes this library emits; a
// decoded element whose Space equals one of them was NOT resolved to a
// URI, i.e. the prefix was unbound at that position.
var conventionalPrefixes = map[string]bool{
	"tds": true, "trt": true, "tev": true, "tt": true, "trc": true,
	"tptz": true, "timg": true, "tan": true, "wsnt": true,
	"wstop": true, "wsa": true, "s": true, "soap": true,
}

// Check scans serialized XML and returns the first element using a
// prefix without an in-scope declaration, or nil when clean.
func Check(data []byte) error {
	dec := xml.NewDecoder(strings.NewReader(string(data)))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}
		if el, ok := tok.(xml.StartElement); ok {
			if conventionalPrefixes[el.Name.Space] {
				return fmt.Errorf("element <%s:%s> uses prefix %q with no in-scope xmlns declaration (strict parsers reject the document):\n%s",
					el.Name.Space, el.Name.Local, el.Name.Space, data)
			}
		}
	}
}
