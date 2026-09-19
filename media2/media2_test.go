package media2_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
	"github.com/mickeyzzc/onvif-go/v2/media2"
)

func TestGetProfiles(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/media2", func(action, reqXML string) (string, error) {
		if action != "tr2:GetProfiles" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetProfilesResponse>
			<Profiles token="main" fixed="true">
				<Name>Main</Name>
				<Configurations>
					<VideoSource token="src-1"><Name>Source</Name><UseCount>1</UseCount></VideoSource>
					<VideoEncoder token="enc-1">
						<Name>Encoder</Name><UseCount>1</UseCount>
						<Encoding>H265</Encoding>
						<Resolution><Width>1920</Width><Height>1080</Height></Resolution>
						<RateControl><FrameRateLimit>15</FrameRateLimit><BitrateLimit>2048</BitrateLimit></RateControl>
						<Quality>5</Quality>
					</VideoEncoder>
					<Metadata token="meta-1"><Name>Metadata</Name><UseCount>1</UseCount></Metadata>
				</Configurations>
			</Profiles>
			</GetProfilesResponse>`, nil
	})

	profiles, err := media2.New(caller).GetProfiles(context.Background(), "", nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(profiles) != 1 {
		t.Fatalf("profiles = %d, want 1", len(profiles))
	}

	p := profiles[0]
	if p.Token != "main" || !p.Fixed || p.Name != "Main" {
		t.Errorf("profile header = %+v", p)
	}

	if p.VideoSource == nil || p.VideoSource.Token != "src-1" {
		t.Errorf("video source ref = %+v", p.VideoSource)
	}

	if p.Metadata == nil || p.Metadata.Token != "meta-1" {
		t.Errorf("metadata ref = %+v", p.Metadata)
	}

	enc := p.VideoEncoder
	if enc == nil {
		t.Fatal("inline video encoder missing")
	}

	if enc.Token != "enc-1" || enc.Encoding != "H265" ||
		enc.Resolution.Width != 1920 || enc.Resolution.Height != 1080 ||
		enc.RateControl == nil || enc.RateControl.FrameRateLimit != 15 {
		t.Errorf("encoder config = %+v", enc)
	}
}

func TestGetVideoEncoderConfigurationOptions(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/media2", func(action, _ string) (string, error) {
		if action != "tr2:GetVideoEncoderConfigurationOptions" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetVideoEncoderConfigurationOptionsResponse>
			<Options GovLengthRange="1 120" FrameRateRange="1 30">
				<Encoding>H264</Encoding>
				<QualityRange><Min>1</Min><Max>6</Max></QualityRange>
				<ResolutionsAvailable><Width>1920</Width><Height>1080</Height></ResolutionsAvailable>
				<BitrateRange><Min>64</Min><Max>8192</Max></BitrateRange>
			</Options>
			<Options GovLengthRange="1 240" FrameRateRange="1 25">
				<Encoding>H265</Encoding>
				<QualityRange><Min>1</Min><Max>6</Max></QualityRange>
				<ResolutionsAvailable><Width>2560</Width><Height>1440</Height></ResolutionsAvailable>
				<BitrateRange><Min>128</Min><Max>16384</Max></BitrateRange>
			</Options>
			</GetVideoEncoderConfigurationOptionsResponse>`, nil
	})

	options, err := media2.New(caller).GetVideoEncoderConfigurationOptions(context.Background(), "enc-1", "")
	if err != nil {
		t.Fatal(err)
	}

	if len(options) != 2 {
		t.Fatalf("options = %d, want one per codec", len(options))
	}

	h264, h265 := options[0], options[1]
	if h264.Encoding != "H264" || len(h264.GovLengthRange) != 2 || h264.GovLengthRange[1] != 120 {
		t.Errorf("H264 options = %+v", h264)
	}

	if h265.Encoding != "H265" || h265.BitrateRange == nil || h265.BitrateRange.Max != 16384 ||
		len(h265.FrameRateRange) != 2 || h265.Resolutions[0].Width != 2560 {
		t.Errorf("H265 options = %+v", h265)
	}
}

func TestSetVideoEncoderConfigurationH265Wire(t *testing.T) {
	var reqXML string

	caller := testutil.NewFakeCaller("http://fake/media2", func(action, body string) (string, error) {
		if action != "tr2:SetVideoEncoderConfiguration" {
			return "", errors.New("unexpected action " + action)
		}

		reqXML = body

		return `<SetVideoEncoderConfigurationResponse/>`, nil
	})

	err := media2.New(caller).SetVideoEncoderConfiguration(context.Background(), &media2.VideoEncoderConfiguration{
		Token:       "enc-1",
		Name:        "Main",
		Encoding:    "H265",
		Resolution:  media2.Resolution{Width: 1920, Height: 1080},
		RateControl: &media2.RateControl{FrameRateLimit: 15, BitrateLimit: 2048},
		Quality:     5,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		`<tr2:SetVideoEncoderConfiguration xmlns:tr2="http://www.onvif.org/ver20/media/wsdl"`,
		`<tt:Encoding>H265</tt:Encoding>`,
		`<tt:Width>1920</tt:Width>`,
		`<tt:FrameRateLimit>15</tt:FrameRateLimit>`,
	} {
		if !strings.Contains(reqXML, want) {
			t.Errorf("set payload missing %q:\n%s", want, reqXML)
		}
	}
}

func TestGetStreamUri(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/media2", func(action, _ string) (string, error) {
		if action != "tr2:GetStreamUri" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetStreamUriResponse><Uri>rtsp://cam:8554/main</Uri></GetStreamUriResponse>`, nil
	})

	uri, err := media2.New(caller).GetStreamUri(context.Background(), "RtspUnicast", "main")
	if err != nil {
		t.Fatal(err)
	}

	if uri != "rtsp://cam:8554/main" {
		t.Errorf("uri = %q", uri)
	}
}
