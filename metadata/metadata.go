// Package metadata parses ONVIF metadata streams — tt:MetadataStream
// documents (metadatastream.xsd, target namespace ver10/schema), the
// Profile M analytics output channel delivered inside the media stream.
// Decoding is lenient by design (unprefixed tags match any namespace), so
// both default-xmlns and prefixed documents from real devices parse.
package metadata

import (
	"encoding/xml"
	"fmt"
	"time"
)

// Parse decodes one tt:MetadataStream document. VideoAnalytics frames
// are modeled; PTZ, Event, and SensorData stream members are not (yet).
func Parse(data []byte) (*Stream, error) {
	var stream Stream
	if err := xml.Unmarshal(data, &stream); err != nil {
		return nil, fmt.Errorf("parse metadata stream: %w", err)
	}

	return &stream, nil
}

// Stream is the decoded MetadataStream: the analytics frames it carried.
type Stream struct {
	XMLName xml.Name `xml:"http://www.onvif.org/ver10/schema MetadataStream"`
	Frames  []Frame  `xml:"VideoAnalytics>Frame"`
}

// Frame is one tt:Frame: a timestamped scene description.
type Frame struct {
	UtcTime        time.Time       `xml:"UtcTime,attr"`
	Transformation *Transformation `xml:"Transformation"`
	Objects        []Object        `xml:"Object"`
}

// Object is one tracked object (tt:Object extends tt:ObjectId).
type Object struct {
	ObjectID   int64       `xml:"ObjectId,attr"`
	UUID       string      `xml:"UUID,attr"`
	Parent     *int64      `xml:"Parent,attr"`
	Appearance *Appearance `xml:"Appearance"`
	Behaviour  *Behaviour  `xml:"Behaviour"`
}

// Appearance is the tt:Appearance descriptor set.
type Appearance struct {
	Transformation *Transformation `xml:"Transformation"`
	Shape          *Shape          `xml:"Shape"`
	Class          *Class          `xml:"Class"`
	GeoLocation    *GeoLocation    `xml:"GeoLocation"`
}

// Shape is tt:ShapeDescriptor: the object's geometric representation.
type Shape struct {
	BoundingBox     *Rectangle `xml:"BoundingBox"`
	Polygon         []Point    `xml:"Polygon>Point"`
	CenterOfGravity *Point     `xml:"CenterOfGravity"`
}

// Rectangle is tt:Rectangle (axis-aligned, pixel or normalized units per
// the frame's transformation).
type Rectangle struct {
	Bottom float64 `xml:"bottom,attr"`
	Top    float64 `xml:"top,attr"`
	Right  float64 `xml:"right,attr"`
	Left   float64 `xml:"left,attr"`
}

// Point is tt:Vector.
type Point struct {
	X float32 `xml:"x,attr"`
	Y float32 `xml:"y,attr"`
}

// Transformation is tt:Transformation (translate + scale).
type Transformation struct {
	Translate *Point `xml:"Translate"`
	Scale     *Point `xml:"Scale"`
}

// Class is tt:ClassDescriptor: the recognized object types with
// likelihoods. Types is the current form; Candidates the deprecated one.
type Class struct {
	Types      []ClassType      `xml:"Type"`
	Candidates []ClassCandidate `xml:"ClassCandidate"`
}

// ClassType is one tt:StringLikelihood entry.
type ClassType struct {
	Type       string   `xml:",chardata"`
	Likelihood *float32 `xml:"Likelihood,attr"`
}

// ClassCandidate is the deprecated tt:ClassCandidate form.
type ClassCandidate struct {
	Type       string   `xml:"Type"`
	Likelihood *float32 `xml:"Likelihood"`
}

// GeoLocation is tt:GeoLocation (angles in degrees).
type GeoLocation struct {
	Lon *float64 `xml:"lon,attr"`
	Lat *float64 `xml:"lat,attr"`
}

// Behaviour is tt:Behaviour: the object's lifecycle state. Removed/Idle
// are marker elements (empty by schema), so presence is the signal —
// nil means absent.
type Behaviour struct {
	Removed *struct{} `xml:"Removed"`
	Idle    *struct{} `xml:"Idle"`
}
