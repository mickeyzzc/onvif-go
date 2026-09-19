package media

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

// TestGetVideoEncoderConfigurationOptionsMPEG4 pins the MPEG-4 options
// branch of the ver10 media options decode (Resolutions/GovLength/
// FrameRate/EncodingInterval ranges + profiles) — the fourth codec family
// of tt:VideoEncoderConfigurationOptions.
func TestGetVideoEncoderConfigurationOptionsMPEG4(t *testing.T) {
	resp := `<GetVideoEncoderConfigurationOptionsResponse>
	<Options>
		<QualityRange><Min>1</Min><Max>6</Max></QualityRange>
		<MPEG4>
			<ResolutionsAvailable><Width>1280</Width><Height>720</Height></ResolutionsAvailable>
			<GovLengthRange><Min>1</Min><Max>60</Max></GovLengthRange>
			<FrameRateRange><Min>1</Min><Max>30</Max></FrameRateRange>
			<EncodingIntervalRange><Min>1</Min><Max>4</Max></EncodingIntervalRange>
			<Mpeg4ProfilesSupported>Simple</Mpeg4ProfilesSupported>
			<Mpeg4ProfilesSupported>ASP</Mpeg4ProfilesSupported>
		</MPEG4>
	</Options>
	</GetVideoEncoderConfigurationOptionsResponse>`

	s, _ := newMediaOpsService(t, "GetVideoEncoderConfigurationOptions", resp)

	opts, err := s.GetVideoEncoderConfigurationOptions(context.Background(), "vec-1")
	if err != nil {
		t.Fatal(err)
	}

	if opts.MPEG4 == nil {
		t.Fatal("MPEG4 options missing")
	}

	mpeg4 := opts.MPEG4
	switch {
	case len(mpeg4.ResolutionsAvailable) != 1 || mpeg4.ResolutionsAvailable[0].Width != 1280:
		t.Errorf("MPEG4 resolutions = %+v", mpeg4.ResolutionsAvailable)
	case mpeg4.GovLengthRange == nil || mpeg4.GovLengthRange.Max != 60:
		t.Errorf("MPEG4 gov length = %+v", mpeg4.GovLengthRange)
	case mpeg4.FrameRateRange == nil || mpeg4.FrameRateRange.Max != 30:
		t.Errorf("MPEG4 frame rate = %+v", mpeg4.FrameRateRange)
	case mpeg4.EncodingIntervalRange == nil || mpeg4.EncodingIntervalRange.Max != 4:
		t.Errorf("MPEG4 interval = %+v", mpeg4.EncodingIntervalRange)
	case len(mpeg4.Mpeg4ProfilesSupported) != 2 || mpeg4.Mpeg4ProfilesSupported[1] != "ASP":
		t.Errorf("MPEG4 profiles = %+v", mpeg4.Mpeg4ProfilesSupported)
	}
}

// TestSetVideoEncoderConfigurationEncodingPassthrough pins that the
// Encoding field is transmitted verbatim — a device advertising H265 (or
// any ver10-Extension/Media2 encoding name) can be configured without the
// client rewriting the enum.
func TestSetVideoEncoderConfigurationEncodingPassthrough(t *testing.T) {
	var reqXML string

	caller := testutil.NewFakeCaller("http://fake/media", func(action, body string) (string, error) {
		if action != "trt:SetVideoEncoderConfiguration" {
			return "", errors.New("unexpected action " + action)
		}

		reqXML = body

		return `<SetVideoEncoderConfigurationResponse/>`, nil
	})

	s := New(caller)
	err := s.SetVideoEncoderConfiguration(context.Background(), &VideoEncoderConfiguration{
		Token:      "enc-1",
		Name:       "Main",
		Encoding:   "H265",
		Resolution: &VideoResolution{Width: 1280, Height: 720},
	}, false)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(reqXML, "<tt:Encoding>H265</tt:Encoding>") {
		t.Errorf("request did not carry the Encoding verbatim:\n%s", reqXML)
	}
}
