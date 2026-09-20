// Example: parsing a Profile M metadata stream.
//
// Cameras with analytics (Profile M) can emit an ONVIF metadata stream:
// XML documents describing tracked objects per frame. metadata.Parse
// decodes them offline — from a file, an RTP payload you demuxed, or
// (as here) an inline sample — into frames, objects, shapes, and
// class likelihoods. No camera needed to run this example.
package main

import (
	"fmt"
	"log"

	"github.com/mickeyzzc/onvif-go/v2/metadata"
)

const sample = `<tt:MetadataStream xmlns:tt="http://www.onvif.org/ver10/schema">
  <tt:VideoAnalytics>
    <tt:Frame UtcTime="2026-09-20T08:00:00Z">
      <tt:Object ObjectId="1">
        <tt:Appearance>
          <tt:Shape>
            <tt:BoundingBox bottom="0.85" left="0.10" right="0.42" top="0.15"/>
            <tt:CenterOfGravity x="0.26" y="0.50"/>
          </tt:Shape>
          <tt:Class>
            <tt:Type Likelihood="0.87">Human</tt:Type>
          </tt:Class>
        </tt:Appearance>
      </tt:Object>
      <tt:Object ObjectId="2">
        <tt:Appearance>
          <tt:Shape>
            <tt:BoundingBox bottom="0.55" left="0.60" right="0.95" top="0.30"/>
          </tt:Shape>
          <tt:Class>
            <tt:Type Likelihood="0.71">Vehicle</tt:Type>
          </tt:Class>
        </tt:Appearance>
        <tt:Behaviour><tt:Removed/></tt:Behaviour>
      </tt:Object>
    </tt:Frame>
  </tt:VideoAnalytics>
</tt:MetadataStream>`

func main() {
	stream, err := metadata.Parse([]byte(sample))
	if err != nil {
		log.Fatalf("Parse failed: %v", err)
	}

	for _, frame := range stream.Frames {
		fmt.Printf("frame %s — %d object(s)\n", frame.UtcTime.Format("15:04:05.000"), len(frame.Objects))

		for _, obj := range frame.Objects {
			line := fmt.Sprintf("  object %d", obj.ObjectID)

			if app := obj.Appearance; app != nil {
				if app.Shape != nil && app.Shape.BoundingBox != nil {
					b := app.Shape.BoundingBox
					line += fmt.Sprintf(" bbox[%.2f,%.2f–%.2f,%.2f]", b.Left, b.Top, b.Right, b.Bottom)
				}
				if app.Class != nil {
					for _, t := range app.Class.Types {
						likelihood := "?"
						if t.Likelihood != nil {
							likelihood = fmt.Sprintf("%.0f%%", *t.Likelihood*100)
						}
						line += fmt.Sprintf(" %s(%s)", t.Type, likelihood)
					}
				}
			}
			if obj.Behaviour != nil && obj.Behaviour.Removed != nil {
				line += " [removed]"
			}
			fmt.Println(line)
		}
	}

	// Namespace-lenient: documents that declare the schema namespace as
	// the default xmlns (no tt: prefix) parse identically. Polygon
	// outlines, geo locations, and the deprecated ClassCandidate form
	// are modeled too — see metadata.Stream.
}
