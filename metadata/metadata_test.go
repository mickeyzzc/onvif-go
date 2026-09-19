package metadata_test

import (
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/metadata"
)

// The golden document mirrors the metadatastream.xsd shapes: a frame with
// a tracked vehicle (bounding box, polygon, class likelihoods, geo
// location) and a second frame whose object is being removed.
const golden = `<?xml version="1.0" encoding="UTF-8"?>
<tt:MetadataStream xmlns:tt="http://www.onvif.org/ver10/schema">
  <tt:VideoAnalytics>
    <tt:Frame UtcTime="2026-09-19T12:00:00.123Z">
      <tt:Transformation>
        <tt:Translate x="0" y="0"/>
        <tt:Scale x="0.5" y="0.5"/>
      </tt:Transformation>
      <tt:Object ObjectId="7" Parent="3">
        <tt:Appearance>
          <tt:Shape>
            <tt:BoundingBox bottom="0.8" top="0.4" right="0.6" left="0.2"/>
            <tt:Polygon>
              <tt:Point x="0.2" y="0.4"/>
              <tt:Point x="0.6" y="0.4"/>
              <tt:Point x="0.6" y="0.8"/>
              <tt:Point x="0.2" y="0.8"/>
            </tt:Polygon>
            <tt:CenterOfGravity x="0.4" y="0.6"/>
          </tt:Shape>
          <tt:Class>
            <tt:Type Likelihood="0.87">Vehicle</tt:Type>
            <tt:ClassCandidate>
              <tt:Type>Human</tt:Type>
              <tt:Likelihood>0.05</tt:Likelihood>
            </tt:ClassCandidate>
          </tt:Class>
          <tt:GeoLocation lon="116.397" lat="39.908"/>
        </tt:Appearance>
      </tt:Object>
    </tt:Frame>
    <tt:Frame UtcTime="2026-09-19T12:00:00.500Z">
      <tt:Object ObjectId="7">
        <tt:Behaviour>
          <tt:Removed/>
        </tt:Behaviour>
      </tt:Object>
    </tt:Frame>
  </tt:VideoAnalytics>
</tt:MetadataStream>`

func TestParseGolden(t *testing.T) {
	stream, err := metadata.Parse([]byte(golden))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if len(stream.Frames) != 2 {
		t.Fatalf("frames = %d, want 2", len(stream.Frames))
	}

	first := stream.Frames[0]
	if got := first.UtcTime.UTC(); got.Format(time.RFC3339Nano) != "2026-09-19T12:00:00.123Z" {
		t.Errorf("UtcTime = %v", got)
	}

	if first.Transformation == nil || first.Transformation.Scale == nil || first.Transformation.Scale.X != 0.5 {
		t.Errorf("frame transformation = %+v", first.Transformation)
	}

	if len(first.Objects) != 1 {
		t.Fatalf("objects = %d, want 1", len(first.Objects))
	}

	obj := first.Objects[0]
	if obj.ObjectID != 7 || obj.Parent == nil || *obj.Parent != 3 {
		t.Errorf("object id/parent = %d/%v", obj.ObjectID, obj.Parent)
	}

	appearance := obj.Appearance
	if appearance == nil {
		t.Fatal("appearance missing")
	}

	box := appearance.Shape.BoundingBox
	if box == nil || box.Left != 0.2 || box.Right != 0.6 || box.Top != 0.4 || box.Bottom != 0.8 {
		t.Errorf("bounding box = %+v", box)
	}

	if len(appearance.Shape.Polygon) != 4 || appearance.Shape.CenterOfGravity == nil {
		t.Errorf("polygon/cog = %+v / %+v", appearance.Shape.Polygon, appearance.Shape.CenterOfGravity)
	}

	if len(appearance.Class.Types) != 1 ||
		appearance.Class.Types[0].Type != "Vehicle" ||
		appearance.Class.Types[0].Likelihood == nil || *appearance.Class.Types[0].Likelihood != 0.87 {
		t.Errorf("class types = %+v", appearance.Class.Types)
	}

	if len(appearance.Class.Candidates) != 1 || appearance.Class.Candidates[0].Type != "Human" {
		t.Errorf("class candidates = %+v", appearance.Class.Candidates)
	}

	if appearance.GeoLocation == nil || appearance.GeoLocation.Lon == nil || *appearance.GeoLocation.Lon != 116.397 {
		t.Errorf("geo location = %+v", appearance.GeoLocation)
	}

	second := stream.Frames[1]
	if len(second.Objects) != 1 || second.Objects[0].Behaviour == nil || second.Objects[0].Behaviour.Removed == nil {
		t.Errorf("second frame object behaviour = %+v", second.Objects[0].Behaviour)
	}
}

// Prefix independence: the same document with default xmlns declarations
// must decode identically.
func TestParseDefaultNamespaceForm(t *testing.T) {
	const doc = `<MetadataStream xmlns="http://www.onvif.org/ver10/schema">
	<VideoAnalytics><Frame UtcTime="2026-09-19T12:00:00Z">
	<Object ObjectId="1"><Appearance><Shape><BoundingBox bottom="1" top="0" right="1" left="0"/></Shape></Appearance></Object>
	</Frame></VideoAnalytics></MetadataStream>`

	stream, err := metadata.Parse([]byte(doc))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	if len(stream.Frames) != 1 || len(stream.Frames[0].Objects) != 1 {
		t.Fatalf("stream = %+v", stream)
	}

	box := stream.Frames[0].Objects[0].Appearance.Shape.BoundingBox
	if box == nil || box.Left != 0 || box.Bottom != 1 {
		t.Errorf("bounding box = %+v", box)
	}
}

func TestParseGarbage(t *testing.T) {
	if _, err := metadata.Parse([]byte("not xml at all")); err == nil {
		t.Error("garbage input unexpectedly parsed")
	}

	if _, err := metadata.Parse([]byte(`<Frame UtcTime="2026-09-19T12:00:00Z"/>`)); err == nil {
		t.Error("non-MetadataStream root unexpectedly accepted")
	}
}
