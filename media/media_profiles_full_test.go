package media

// Full-mapping tests for the profile and encoder-options responses:
// every optional sub-configuration branch (source bounds, encoder
// resolution + rate control, PTZ node, JPEG/H264 option ranges) is
// exercised with a complete response body, plus the caller-failure path.

import (
	"context"
	"errors"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

func TestGetProfilesFullMapping(t *testing.T) {
	resp := `<GetProfilesResponse>
	<Profiles token="prof-1">
		<Name>Main</Name>
		<VideoSourceConfiguration token="vsc-1"><Name>Src</Name><UseCount>3</UseCount><SourceToken>src-0</SourceToken><Bounds x="0" y="0" width="2560" height="1440"/></VideoSourceConfiguration>
		<VideoEncoderConfiguration token="vec-1"><Name>H264</Name><UseCount>3</UseCount><Encoding>H264</Encoding><Quality>5</Quality><Resolution><Width>2560</Width><Height>1440</Height></Resolution><RateControl><FrameRateLimit>25</FrameRateLimit><EncodingInterval>1</EncodingInterval><BitrateLimit>4096</BitrateLimit></RateControl></VideoEncoderConfiguration>
		<PTZConfiguration token="ptz-1"><Name>PTZ</Name><UseCount>1</UseCount><NodeToken>node-0</NodeToken></PTZConfiguration>
	</Profiles>
	<Profiles token="prof-2"><Name>Bare</Name></Profiles>
	</GetProfilesResponse>`

	s, _ := newMediaOpsService(t, "GetProfiles", resp)

	profiles, err := s.GetProfiles(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(profiles) != 2 {
		t.Fatalf("profiles = %d, want 2", len(profiles))
	}

	main := profiles[0]
	if main.Token != "prof-1" || main.Name != "Main" {
		t.Errorf("profile identity = %+v", main)
	}

	vsc := main.VideoSourceConfiguration
	if vsc == nil || vsc.Token != "vsc-1" || vsc.SourceToken != "src-0" || vsc.UseCount != 3 {
		t.Errorf("video source config = %+v", vsc)
	}

	if vsc.Bounds == nil || vsc.Bounds.Width != 2560 || vsc.Bounds.Height != 1440 {
		t.Errorf("bounds = %+v", vsc.Bounds)
	}

	vec := main.VideoEncoderConfiguration
	if vec == nil || vec.Encoding != "H264" || vec.Quality != 5 {
		t.Errorf("video encoder config = %+v", vec)
	}

	if vec.Resolution == nil || vec.Resolution.Width != 2560 || vec.Resolution.Height != 1440 {
		t.Errorf("encoder resolution = %+v", vec.Resolution)
	}

	if vec.RateControl == nil || vec.RateControl.FrameRateLimit != 25 || vec.RateControl.BitrateLimit != 4096 {
		t.Errorf("rate control = %+v", vec.RateControl)
	}

	if main.PTZConfiguration == nil || main.PTZConfiguration.NodeToken != "node-0" {
		t.Errorf("ptz config = %+v", main.PTZConfiguration)
	}

	// Second profile carries none of the optional configurations.
	bare := profiles[1]
	if bare.VideoSourceConfiguration != nil || bare.VideoEncoderConfiguration != nil || bare.PTZConfiguration != nil {
		t.Errorf("bare profile = %+v", bare)
	}
}

func TestGetProfilesCallerFailure(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/media", func(string, string) (string, error) {
		return "", errors.New("device unreachable")
	})

	if _, err := New(caller).GetProfiles(context.Background()); err == nil {
		t.Fatal("caller failure not propagated")
	}
}

func TestGetVideoEncoderConfigurationOptionsFull(t *testing.T) {
	resp := `<GetVideoEncoderConfigurationOptionsResponse>
	<Options>
		<QualityRange><Min>1</Min><Max>6</Max></QualityRange>
		<JPEG>
			<ResolutionsAvailable><Width>1920</Width><Height>1080</Height></ResolutionsAvailable>
			<FrameRateRange><Min>1</Min><Max>30</Max></FrameRateRange>
			<EncodingIntervalRange><Min>1</Min><Max>3</Max></EncodingIntervalRange>
		</JPEG>
		<H264>
			<ResolutionsAvailable><Width>2560</Width><Height>1440</Height></ResolutionsAvailable>
			<GovLengthRange><Min>1</Min><Max>120</Max></GovLengthRange>
			<FrameRateRange><Min>1</Min><Max>25</Max></FrameRateRange>
			<EncodingIntervalRange><Min>1</Min><Max>2</Max></EncodingIntervalRange>
			<H264ProfilesSupported>Main</H264ProfilesSupported>
			<H264ProfilesSupported>High</H264ProfilesSupported>
		</H264>
	</Options>
	</GetVideoEncoderConfigurationOptionsResponse>`

	s, caller := newMediaOpsService(t, "GetVideoEncoderConfigurationOptions", resp)

	opts, err := s.GetVideoEncoderConfigurationOptions(context.Background(), "vec-1")
	if err != nil {
		t.Fatal(err)
	}

	if opts.QualityRange == nil || opts.QualityRange.Min != 1 || opts.QualityRange.Max != 6 {
		t.Errorf("quality range = %+v", opts.QualityRange)
	}

	if opts.JPEG == nil {
		t.Fatal("JPEG options missing")
	}

	if len(opts.JPEG.ResolutionsAvailable) != 1 || opts.JPEG.ResolutionsAvailable[0].Width != 1920 {
		t.Errorf("JPEG resolutions = %+v", opts.JPEG.ResolutionsAvailable)
	}

	if opts.JPEG.FrameRateRange == nil || opts.JPEG.FrameRateRange.Max != 30 {
		t.Errorf("JPEG frame rate = %+v", opts.JPEG.FrameRateRange)
	}

	if opts.JPEG.EncodingIntervalRange == nil || opts.JPEG.EncodingIntervalRange.Max != 3 {
		t.Errorf("JPEG interval = %+v", opts.JPEG.EncodingIntervalRange)
	}

	if opts.H264 == nil {
		t.Fatal("H264 options missing")
	}

	if opts.H264.GovLengthRange == nil || opts.H264.GovLengthRange.Max != 120 {
		t.Errorf("H264 gov length = %+v", opts.H264.GovLengthRange)
	}

	if len(opts.H264.ResolutionsAvailable) != 1 || opts.H264.ResolutionsAvailable[0].Height != 1440 {
		t.Errorf("H264 resolutions = %+v", opts.H264.ResolutionsAvailable)
	}

	if len(opts.H264.H264ProfilesSupported) != 2 {
		t.Errorf("H264 profiles = %v", opts.H264.H264ProfilesSupported)
	}

	if caller.CountAction("trt:GetVideoEncoderConfigurationOptions") != 1 {
		t.Error("configuration token request not sent exactly once")
	}
}

func TestGetVideoEncoderConfigurationOptionsEmpty(t *testing.T) {
	s, _ := newMediaOpsService(t, "GetVideoEncoderConfigurationOptions", `<GetVideoEncoderConfigurationOptionsResponse><Options/></GetVideoEncoderConfigurationOptionsResponse>`)

	opts, err := s.GetVideoEncoderConfigurationOptions(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}

	if opts.QualityRange != nil || opts.JPEG != nil || opts.H264 != nil {
		t.Errorf("empty options = %+v", opts)
	}
}

func TestGetMediaServiceCapabilitiesFull(t *testing.T) {
	resp := `<GetServiceCapabilitiesResponse>
	<Capabilities SnapshotUri="true" Rotation="false" VideoSourceMode="true" OSD="true" TemporaryOSDText="false" EXICompression="false">
		<ProfileCapabilities MaximumNumberOfProfiles="10"/>
		<StreamingCapabilities RTPMulticast="false" RTP_TCP="true" RTP_RTSP_TCP="true"/>
	</Capabilities>
	</GetServiceCapabilitiesResponse>`

	s, _ := newMediaOpsService(t, "GetServiceCapabilities", resp)

	caps, err := s.GetMediaServiceCapabilities(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !caps.SnapshotURI || caps.Rotation || !caps.VideoSourceMode || !caps.OSD {
		t.Errorf("capabilities attrs = %+v", caps)
	}

	if caps.MaximumNumberOfProfiles != 10 {
		t.Errorf("maximum profiles = %d, want 10", caps.MaximumNumberOfProfiles)
	}

	if !caps.RTPTCP || !caps.RTPRTSPTCP || caps.RTPMulticast {
		t.Errorf("streaming capabilities = %+v", caps)
	}
}
