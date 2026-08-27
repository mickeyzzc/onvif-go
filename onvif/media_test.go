package onvif

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGetProfiles tests GetProfiles operation.
func TestGetProfiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetProfilesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Profiles token="Profile1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">Main Profile</tt:Name>
				<tt:VideoEncoderConfiguration xmlns:tt="http://www.onvif.org/ver10/schema" token="VideoEnc1">
					<tt:Encoding>H264</tt:Encoding>
					<tt:Resolution>
						<tt:Width>1920</tt:Width>
						<tt:Height>1080</tt:Height>
					</tt:Resolution>
					<tt:Quality>5.0</tt:Quality>
				</tt:VideoEncoderConfiguration>
			</trt:Profiles>
		</trt:GetProfilesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	profiles, err := client.Media().GetProfiles(ctx)
	if err != nil {
		t.Fatalf("GetProfiles() failed: %v", err)
	}

	if len(profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(profiles))
	}

	if profiles[0].Token != "Profile1" {
		t.Errorf("Expected token Profile1, got %s", profiles[0].Token)
	}

	if profiles[0].Name != "Main Profile" {
		t.Errorf("Expected name 'Main Profile', got %s", profiles[0].Name)
	}
}

// TestGetProfile tests GetProfile operation.
func TestGetProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetProfileResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Profile token="Profile1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">Main Profile</tt:Name>
			</trt:Profile>
		</trt:GetProfileResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	profile, err := client.Media().GetProfile(ctx, "Profile1")
	if err != nil {
		t.Fatalf("GetProfile() failed: %v", err)
	}

	if profile.Token != "Profile1" {
		t.Errorf("Expected token Profile1, got %s", profile.Token)
	}
}

// TestSetProfile tests SetProfile operation.
func TestSetProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetProfileResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	profile := &Profile{
		Token: "Profile1",
		Name:  "Updated Profile",
	}

	err = client.Media().SetProfile(ctx, profile)
	if err != nil {
		t.Fatalf("SetProfile() failed: %v", err)
	}
}

// TestGetStreamURI tests GetStreamURI operation.
func TestGetStreamURI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:MediaUri>
				<tt:Uri xmlns:tt="http://www.onvif.org/ver10/schema">rtsp://192.168.1.100:554/stream1</tt:Uri>
				<tt:InvalidAfterConnect xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:InvalidAfterConnect>
				<tt:InvalidAfterReboot xmlns:tt="http://www.onvif.org/ver10/schema">true</tt:InvalidAfterReboot>
			</trt:MediaUri>
		</trt:GetStreamUriResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	uri, err := client.Media().GetStreamURI(ctx, "Profile1")
	if err != nil {
		t.Fatalf("GetStreamURI() failed: %v", err)
	}

	if uri.URI != "rtsp://192.168.1.100:554/stream1" {
		t.Errorf("Expected URI 'rtsp://192.168.1.100:554/stream1', got %s", uri.URI)
	}
}

// TestGetSnapshotURI tests GetSnapshotURI operation.
func TestGetSnapshotURI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetSnapshotUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:MediaUri>
				<tt:Uri xmlns:tt="http://www.onvif.org/ver10/schema">http://192.168.1.100/snapshot.jpg</tt:Uri>
			</trt:MediaUri>
		</trt:GetSnapshotUriResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	uri, err := client.Media().GetSnapshotURI(ctx, "Profile1")
	if err != nil {
		t.Fatalf("GetSnapshotURI() failed: %v", err)
	}

	if !strings.Contains(uri.URI, "snapshot") {
		t.Errorf("Expected snapshot URI, got %s", uri.URI)
	}
}

// TestGetVideoSources tests GetVideoSources operation.
func TestGetVideoSources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoSourcesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:VideoSources token="VideoSource1">
				<tt:Framerate xmlns:tt="http://www.onvif.org/ver10/schema">30.0</tt:Framerate>
				<tt:Resolution xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Width>1920</tt:Width>
					<tt:Height>1080</tt:Height>
				</tt:Resolution>
			</trt:VideoSources>
		</trt:GetVideoSourcesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	sources, err := client.Media().GetVideoSources(ctx)
	if err != nil {
		t.Fatalf("GetVideoSources() failed: %v", err)
	}

	if len(sources) != 1 {
		t.Errorf("Expected 1 video source, got %d", len(sources))
	}

	if sources[0].Token != "VideoSource1" {
		t.Errorf("Expected token VideoSource1, got %s", sources[0].Token)
	}
}

// TestGetAudioSources tests GetAudioSources operation.
func TestGetAudioSources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioSourcesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:AudioSources token="AudioSource1">
				<tt:Channels xmlns:tt="http://www.onvif.org/ver10/schema">2</tt:Channels>
			</trt:AudioSources>
		</trt:GetAudioSourcesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	sources, err := client.Media().GetAudioSources(ctx)
	if err != nil {
		t.Fatalf("GetAudioSources() failed: %v", err)
	}

	if len(sources) != 1 {
		t.Errorf("Expected 1 audio source, got %d", len(sources))
	}
}

// TestGetAudioOutputs tests GetAudioOutputs operation.
func TestGetAudioOutputs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioOutputsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:AudioOutputs token="AudioOutput1"/>
		</trt:GetAudioOutputsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	outputs, err := client.Media().GetAudioOutputs(ctx)
	if err != nil {
		t.Fatalf("GetAudioOutputs() failed: %v", err)
	}

	if len(outputs) != 1 {
		t.Errorf("Expected 1 audio output, got %d", len(outputs))
	}
}

// TestCreateProfile tests CreateProfile operation.
func TestCreateProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:CreateProfileResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Profile token="NewProfile1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">New Profile</tt:Name>
			</trt:Profile>
		</trt:CreateProfileResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	profile, err := client.Media().CreateProfile(ctx, "New Profile", "")
	if err != nil {
		t.Fatalf("CreateProfile() failed: %v", err)
	}

	if profile.Token != "NewProfile1" {
		t.Errorf("Expected token NewProfile1, got %s", profile.Token)
	}
}

// TestDeleteProfile tests DeleteProfile operation.
func TestDeleteProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:DeleteProfileResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().DeleteProfile(ctx, "Profile1")
	if err != nil {
		t.Fatalf("DeleteProfile() failed: %v", err)
	}
}

// TestGetVideoEncoderConfiguration tests GetVideoEncoderConfiguration operation.
func TestGetVideoEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoEncoderConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="VideoEnc1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">H264 Config</tt:Name>
				<tt:Encoding xmlns:tt="http://www.onvif.org/ver10/schema">H264</tt:Encoding>
				<tt:Resolution xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Width>1920</tt:Width>
					<tt:Height>1080</tt:Height>
				</tt:Resolution>
				<tt:Quality xmlns:tt="http://www.onvif.org/ver10/schema">5.0</tt:Quality>
			</trt:Configuration>
		</trt:GetVideoEncoderConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config, err := client.Media().GetVideoEncoderConfiguration(ctx, "VideoEnc1")
	if err != nil {
		t.Fatalf("GetVideoEncoderConfiguration() failed: %v", err)
	}

	if config.Token != "VideoEnc1" {
		t.Errorf("Expected token VideoEnc1, got %s", config.Token)
	}

	if config.Encoding != "H264" {
		t.Errorf("Expected encoding H264, got %s", config.Encoding)
	}
}

// TestSetVideoEncoderConfiguration tests SetVideoEncoderConfiguration operation.
func TestSetVideoEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetVideoEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config := &VideoEncoderConfiguration{
		Token:    "VideoEnc1",
		Name:     "H264 Config",
		Encoding: "H264",
		Resolution: &VideoResolution{
			Width:  1920,
			Height: 1080,
		},
		Quality: 5.0,
	}

	err = client.Media().SetVideoEncoderConfiguration(ctx, config, true)
	if err != nil {
		t.Fatalf("SetVideoEncoderConfiguration() failed: %v", err)
	}
}

// TestGetMediaServiceCapabilities tests GetMediaServiceCapabilities operation.
func TestGetMediaServiceCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetServiceCapabilitiesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Capabilities SnapshotUri="true" Rotation="true" OSD="true">
				<trt:ProfileCapabilities MaximumNumberOfProfiles="10"/>
				<trt:StreamingCapabilities RTPMulticast="true" RTP_TCP="true"/>
			</trt:Capabilities>
		</trt:GetServiceCapabilitiesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	caps, err := client.Media().GetMediaServiceCapabilities(ctx)
	if err != nil {
		t.Fatalf("GetMediaServiceCapabilities() failed: %v", err)
	}

	if !caps.SnapshotURI {
		t.Error("Expected SnapshotURI to be true")
	}

	if caps.MaximumNumberOfProfiles != 10 {
		t.Errorf("Expected MaximumNumberOfProfiles 10, got %d", caps.MaximumNumberOfProfiles)
	}
}

// TestGetVideoEncoderConfigurationOptions tests GetVideoEncoderConfigurationOptions operation.
func TestGetVideoEncoderConfigurationOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoEncoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:QualityRange xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Min>1.0</tt:Min>
					<tt:Max>10.0</tt:Max>
				</tt:QualityRange>
				<tt:H264 xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:ResolutionsAvailable>
						<tt:Width>1920</tt:Width>
						<tt:Height>1080</tt:Height>
					</tt:ResolutionsAvailable>
					<tt:H264ProfilesSupported>Baseline</tt:H264ProfilesSupported>
				</tt:H264>
			</trt:Options>
		</trt:GetVideoEncoderConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.Media().GetVideoEncoderConfigurationOptions(ctx, "VideoEnc1")
	if err != nil {
		t.Fatalf("GetVideoEncoderConfigurationOptions() failed: %v", err)
	}

	if options.QualityRange == nil {
		t.Error("Expected QualityRange to be set")
	}

	if options.H264 == nil {
		t.Error("Expected H264 options to be set")
	}
}

// TestGetAudioEncoderConfiguration tests GetAudioEncoderConfiguration operation.
func TestGetAudioEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioEncoderConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="AudioEnc1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">AAC Config</tt:Name>
				<tt:Encoding xmlns:tt="http://www.onvif.org/ver10/schema">AAC</tt:Encoding>
				<tt:Bitrate xmlns:tt="http://www.onvif.org/ver10/schema">128000</tt:Bitrate>
				<tt:SampleRate xmlns:tt="http://www.onvif.org/ver10/schema">48000</tt:SampleRate>
			</trt:Configuration>
		</trt:GetAudioEncoderConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config, err := client.Media().GetAudioEncoderConfiguration(ctx, "AudioEnc1")
	if err != nil {
		t.Fatalf("GetAudioEncoderConfiguration() failed: %v", err)
	}

	if config.Token != "AudioEnc1" {
		t.Errorf("Expected token AudioEnc1, got %s", config.Token)
	}

	if config.Encoding != "AAC" {
		t.Errorf("Expected encoding AAC, got %s", config.Encoding)
	}
}

// TestSetAudioEncoderConfiguration tests SetAudioEncoderConfiguration operation.
func TestSetAudioEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetAudioEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config := &AudioEncoderConfiguration{
		Token:      "AudioEnc1",
		Name:       "AAC Config",
		Encoding:   "AAC",
		Bitrate:    128000,
		SampleRate: 48000,
	}

	err = client.Media().SetAudioEncoderConfiguration(ctx, config, true)
	if err != nil {
		t.Fatalf("SetAudioEncoderConfiguration() failed: %v", err)
	}
}

// TestGetMetadataConfiguration tests GetMetadataConfiguration operation.
func TestGetMetadataConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetMetadataConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="Metadata1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">Metadata Config</tt:Name>
				<tt:PTZStatus xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Status>true</tt:Status>
					<tt:Position>true</tt:Position>
				</tt:PTZStatus>
				<tt:Analytics xmlns:tt="http://www.onvif.org/ver10/schema">false</tt:Analytics>
			</trt:Configuration>
		</trt:GetMetadataConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config, err := client.Media().GetMetadataConfiguration(ctx, "Metadata1")
	if err != nil {
		t.Fatalf("GetMetadataConfiguration() failed: %v", err)
	}

	if config.Token != "Metadata1" {
		t.Errorf("Expected token Metadata1, got %s", config.Token)
	}

	if config.PTZStatus == nil {
		t.Error("Expected PTZStatus to be set")
	}
}

// TestSetMetadataConfiguration tests SetMetadataConfiguration operation.
func TestSetMetadataConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetMetadataConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config := &MetadataConfiguration{
		Token:     "Metadata1",
		Name:      "Metadata Config",
		Analytics: false,
		PTZStatus: &PTZFilter{
			Status:   true,
			Position: true,
		},
	}

	err = client.Media().SetMetadataConfiguration(ctx, config, true)
	if err != nil {
		t.Fatalf("SetMetadataConfiguration() failed: %v", err)
	}
}

// TestGetVideoSourceModes tests GetVideoSourceModes operation.
func TestGetVideoSourceModes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetVideoSourceModesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:VideoSourceModes token="Mode1">
				<tt:Enabled xmlns:tt="http://www.onvif.org/ver10/schema">true</tt:Enabled>
				<tt:Resolution xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Width>1920</tt:Width>
					<tt:Height>1080</tt:Height>
				</tt:Resolution>
			</trt:VideoSourceModes>
		</trt:GetVideoSourceModesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	modes, err := client.Media().GetVideoSourceModes(ctx, "VideoSource1")
	if err != nil {
		t.Fatalf("GetVideoSourceModes() failed: %v", err)
	}

	if len(modes) != 1 {
		t.Errorf("Expected 1 mode, got %d", len(modes))
	}

	if modes[0].Token != "Mode1" {
		t.Errorf("Expected token Mode1, got %s", modes[0].Token)
	}
}

// TestSetVideoSourceMode tests SetVideoSourceMode operation.
func TestSetVideoSourceMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetVideoSourceModeResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().SetVideoSourceMode(ctx, "VideoSource1", "Mode1")
	if err != nil {
		t.Fatalf("SetVideoSourceMode() failed: %v", err)
	}
}

// TestSetSynchronizationPoint tests SetSynchronizationPoint operation.
func TestSetSynchronizationPoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetSynchronizationPointResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().SetSynchronizationPoint(ctx, "Profile1")
	if err != nil {
		t.Fatalf("SetSynchronizationPoint() failed: %v", err)
	}
}

// TestGetOSDs tests GetOSDs operation.
func TestGetOSDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetOSDsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:OSDs token="OSD1"/>
			<trt:OSDs token="OSD2"/>
		</trt:GetOSDsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	osds, err := client.Media().GetOSDs(ctx, "")
	if err != nil {
		t.Fatalf("GetOSDs() failed: %v", err)
	}

	if len(osds) != 2 {
		t.Errorf("Expected 2 OSDs, got %d", len(osds))
	}
}

// TestGetOSD tests GetOSD operation.
func TestGetOSD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetOSDResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:OSD token="OSD1"/>
		</trt:GetOSDResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	osd, err := client.Media().GetOSD(ctx, "OSD1")
	if err != nil {
		t.Fatalf("GetOSD() failed: %v", err)
	}

	if osd.Token != "OSD1" {
		t.Errorf("Expected token OSD1, got %s", osd.Token)
	}
}

// TestSetOSD tests SetOSD operation.
func TestSetOSD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetOSDResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	osd := &OSDConfiguration{
		Token: "OSD1",
	}

	err = client.Media().SetOSD(ctx, osd)
	if err != nil {
		t.Fatalf("SetOSD() failed: %v", err)
	}
}

// TestCreateOSD tests CreateOSD operation.
func TestCreateOSD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:CreateOSDResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:OSD token="NewOSD1"/>
		</trt:CreateOSDResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	osd, err := client.Media().CreateOSD(ctx, "VideoSourceConfig1", nil)
	if err != nil {
		t.Fatalf("CreateOSD() failed: %v", err)
	}

	if osd.Token != "NewOSD1" {
		t.Errorf("Expected token NewOSD1, got %s", osd.Token)
	}
}

// TestDeleteOSD tests DeleteOSD operation.
func TestDeleteOSD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:DeleteOSDResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().DeleteOSD(ctx, "OSD1")
	if err != nil {
		t.Fatalf("DeleteOSD() failed: %v", err)
	}
}

// TestStartMulticastStreaming tests StartMulticastStreaming operation.
func TestStartMulticastStreaming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:StartMulticastStreamingResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().StartMulticastStreaming(ctx, "Profile1")
	if err != nil {
		t.Fatalf("StartMulticastStreaming() failed: %v", err)
	}
}

// TestStopMulticastStreaming tests StopMulticastStreaming operation.
func TestStopMulticastStreaming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:StopMulticastStreamingResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().StopMulticastStreaming(ctx, "Profile1")
	if err != nil {
		t.Fatalf("StopMulticastStreaming() failed: %v", err)
	}
}

// TestAddVideoEncoderConfiguration tests AddVideoEncoderConfiguration operation.
func TestAddVideoEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddVideoEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().AddVideoEncoderConfiguration(ctx, "Profile1", "VideoEnc1")
	if err != nil {
		t.Fatalf("AddVideoEncoderConfiguration() failed: %v", err)
	}
}

// TestRemoveVideoEncoderConfiguration tests RemoveVideoEncoderConfiguration operation.
func TestRemoveVideoEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveVideoEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().RemoveVideoEncoderConfiguration(ctx, "Profile1")
	if err != nil {
		t.Fatalf("RemoveVideoEncoderConfiguration() failed: %v", err)
	}
}

// TestAddAudioEncoderConfiguration tests AddAudioEncoderConfiguration operation.
func TestAddAudioEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddAudioEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().AddAudioEncoderConfiguration(ctx, "Profile1", "AudioEnc1")
	if err != nil {
		t.Fatalf("AddAudioEncoderConfiguration() failed: %v", err)
	}
}

// TestRemoveAudioEncoderConfiguration tests RemoveAudioEncoderConfiguration operation.
func TestRemoveAudioEncoderConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveAudioEncoderConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().RemoveAudioEncoderConfiguration(ctx, "Profile1")
	if err != nil {
		t.Fatalf("RemoveAudioEncoderConfiguration() failed: %v", err)
	}
}

// TestAddAudioSourceConfiguration tests AddAudioSourceConfiguration operation.
func TestAddAudioSourceConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddAudioSourceConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().AddAudioSourceConfiguration(ctx, "Profile1", "AudioSourceConfig1")
	if err != nil {
		t.Fatalf("AddAudioSourceConfiguration() failed: %v", err)
	}
}

// TestRemoveAudioSourceConfiguration tests RemoveAudioSourceConfiguration operation.
func TestRemoveAudioSourceConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveAudioSourceConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().RemoveAudioSourceConfiguration(ctx, "Profile1")
	if err != nil {
		t.Fatalf("RemoveAudioSourceConfiguration() failed: %v", err)
	}
}

// TestAddVideoSourceConfiguration tests AddVideoSourceConfiguration operation.
func TestAddVideoSourceConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddVideoSourceConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().AddVideoSourceConfiguration(ctx, "Profile1", "VideoSourceConfig1")
	if err != nil {
		t.Fatalf("AddVideoSourceConfiguration() failed: %v", err)
	}
}

// TestRemoveVideoSourceConfiguration tests RemoveVideoSourceConfiguration operation.
func TestRemoveVideoSourceConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveVideoSourceConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().RemoveVideoSourceConfiguration(ctx, "Profile1")
	if err != nil {
		t.Fatalf("RemoveVideoSourceConfiguration() failed: %v", err)
	}
}

// TestAddPTZConfiguration tests AddPTZConfiguration operation.
func TestAddPTZConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddPTZConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().AddPTZConfiguration(ctx, "Profile1", "PTZConfig1")
	if err != nil {
		t.Fatalf("AddPTZConfiguration() failed: %v", err)
	}
}

// TestRemovePTZConfiguration tests RemovePTZConfiguration operation.
func TestRemovePTZConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemovePTZConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().RemovePTZConfiguration(ctx, "Profile1")
	if err != nil {
		t.Fatalf("RemovePTZConfiguration() failed: %v", err)
	}
}

// TestAddMetadataConfiguration tests AddMetadataConfiguration operation.
func TestAddMetadataConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:AddMetadataConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().AddMetadataConfiguration(ctx, "Profile1", "Metadata1")
	if err != nil {
		t.Fatalf("AddMetadataConfiguration() failed: %v", err)
	}
}

// TestRemoveMetadataConfiguration tests RemoveMetadataConfiguration operation.
func TestRemoveMetadataConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:RemoveMetadataConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	err = client.Media().RemoveMetadataConfiguration(ctx, "Profile1")
	if err != nil {
		t.Fatalf("RemoveMetadataConfiguration() failed: %v", err)
	}
}

// TestGetAudioEncoderConfigurationOptions tests GetAudioEncoderConfigurationOptions operation.
func TestGetAudioEncoderConfigurationOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioEncoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:EncodingOptions xmlns:tt="http://www.onvif.org/ver10/schema">AAC</tt:EncodingOptions>
				<tt:EncodingOptions xmlns:tt="http://www.onvif.org/ver10/schema">G711</tt:EncodingOptions>
				<tt:BitrateList xmlns:tt="http://www.onvif.org/ver10/schema">64000</tt:BitrateList>
				<tt:BitrateList xmlns:tt="http://www.onvif.org/ver10/schema">128000</tt:BitrateList>
				<tt:SampleRateList xmlns:tt="http://www.onvif.org/ver10/schema">44100</tt:SampleRateList>
				<tt:SampleRateList xmlns:tt="http://www.onvif.org/ver10/schema">48000</tt:SampleRateList>
			</trt:Options>
		</trt:GetAudioEncoderConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.Media().GetAudioEncoderConfigurationOptions(ctx, "AudioEnc1", "")
	if err != nil {
		t.Fatalf("GetAudioEncoderConfigurationOptions() failed: %v", err)
	}

	if len(options.EncodingOptions) == 0 {
		t.Error("Expected encoding options to be set")
	}
}

// TestGetMetadataConfigurationOptions tests GetMetadataConfigurationOptions operation.
func TestGetMetadataConfigurationOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetMetadataConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:PTZStatusFilterOptions xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:Status>true</tt:Status>
					<tt:Position>true</tt:Position>
				</tt:PTZStatusFilterOptions>
			</trt:Options>
		</trt:GetMetadataConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.Media().GetMetadataConfigurationOptions(ctx, "Metadata1", "")
	if err != nil {
		t.Fatalf("GetMetadataConfigurationOptions() failed: %v", err)
	}

	if options.PTZStatusFilterOptions == nil {
		t.Error("Expected PTZStatusFilterOptions to be set")
	}
}

// TestGetAudioOutputConfiguration tests GetAudioOutputConfiguration operation.
func TestGetAudioOutputConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioOutputConfigurationResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Configuration token="AudioOutputConfig1">
				<tt:Name xmlns:tt="http://www.onvif.org/ver10/schema">Audio Output Config</tt:Name>
				<tt:OutputToken xmlns:tt="http://www.onvif.org/ver10/schema">AudioOutput1</tt:OutputToken>
			</trt:Configuration>
		</trt:GetAudioOutputConfigurationResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config, err := client.Media().GetAudioOutputConfiguration(ctx, "AudioOutputConfig1")
	if err != nil {
		t.Fatalf("GetAudioOutputConfiguration() failed: %v", err)
	}

	if config.Token != "AudioOutputConfig1" {
		t.Errorf("Expected token AudioOutputConfig1, got %s", config.Token)
	}
}

// TestSetAudioOutputConfiguration tests SetAudioOutputConfiguration operation.
func TestSetAudioOutputConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<?xml version="1.0"?><soap:Envelope><soap:Body><trt:SetAudioOutputConfigurationResponse/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	config := &AudioOutputConfiguration{
		Token:       "AudioOutputConfig1",
		Name:        "Audio Output Config",
		OutputToken: "AudioOutput1",
	}

	err = client.Media().SetAudioOutputConfiguration(ctx, config, true)
	if err != nil {
		t.Fatalf("SetAudioOutputConfiguration() failed: %v", err)
	}
}

// TestGetAudioOutputConfigurationOptions tests GetAudioOutputConfigurationOptions operation.
func TestGetAudioOutputConfigurationOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioOutputConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:OutputTokensAvailable xmlns:tt="http://www.onvif.org/ver10/schema">AudioOutput1</tt:OutputTokensAvailable>
				<tt:OutputTokensAvailable xmlns:tt="http://www.onvif.org/ver10/schema">AudioOutput2</tt:OutputTokensAvailable>
			</trt:Options>
		</trt:GetAudioOutputConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.Media().GetAudioOutputConfigurationOptions(ctx, "")
	if err != nil {
		t.Fatalf("GetAudioOutputConfigurationOptions() failed: %v", err)
	}

	if len(options.OutputTokensAvailable) == 0 {
		t.Error("Expected output tokens to be available")
	}
}

// TestGetAudioDecoderConfigurationOptions tests GetAudioDecoderConfigurationOptions operation.
func TestGetAudioDecoderConfigurationOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetAudioDecoderConfigurationOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:AACDecOptions xmlns:tt="http://www.onvif.org/ver10/schema">
					<tt:BitrateList>64000</tt:BitrateList>
					<tt:BitrateList>128000</tt:BitrateList>
					<tt:SampleRateList>44100</tt:SampleRateList>
					<tt:SampleRateList>48000</tt:SampleRateList>
				</tt:AACDecOptions>
			</trt:Options>
		</trt:GetAudioDecoderConfigurationOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.Media().GetAudioDecoderConfigurationOptions(ctx, "")
	if err != nil {
		t.Fatalf("GetAudioDecoderConfigurationOptions() failed: %v", err)
	}

	if options.AACDecOptions == nil {
		t.Error("Expected AACDecOptions to be set")
	}
}

// TestGetGuaranteedNumberOfVideoEncoderInstances tests GetGuaranteedNumberOfVideoEncoderInstances operation.
func TestGetGuaranteedNumberOfVideoEncoderInstances(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetGuaranteedNumberOfVideoEncoderInstancesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:TotalNumber>4</trt:TotalNumber>
			<trt:JPEG>2</trt:JPEG>
			<trt:H264>2</trt:H264>
			<trt:MPEG4>0</trt:MPEG4>
		</trt:GetGuaranteedNumberOfVideoEncoderInstancesResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	instances, err := client.Media().GetGuaranteedNumberOfVideoEncoderInstances(ctx, "VideoEnc1")
	if err != nil {
		t.Fatalf("GetGuaranteedNumberOfVideoEncoderInstances() failed: %v", err)
	}

	if instances.TotalNumber != 4 {
		t.Errorf("Expected TotalNumber 4, got %d", instances.TotalNumber)
	}

	if instances.H264 != 2 {
		t.Errorf("Expected H264 2, got %d", instances.H264)
	}
}

// TestGetOSDOptions tests GetOSDOptions operation.
func TestGetOSDOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
	<soap:Body>
		<trt:GetOSDOptionsResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
			<trt:Options>
				<tt:MaximumNumberOfOSDs xmlns:tt="http://www.onvif.org/ver10/schema">10</tt:MaximumNumberOfOSDs>
			</trt:Options>
		</trt:GetOSDOptionsResponse>
	</soap:Body>
</soap:Envelope>`
		w.Header().Set("Content-Type", "application/soap+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL + "/onvif/media_service")
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}

	ctx := context.Background()
	options, err := client.Media().GetOSDOptions(ctx, "")
	if err != nil {
		t.Fatalf("GetOSDOptions() failed: %v", err)
	}

	if options.MaximumNumberOfOSDs != 10 {
		t.Errorf("Expected MaximumNumberOfOSDs 10, got %d", options.MaximumNumberOfOSDs)
	}
}

// profile helper: token, name, W×H (-1 resolution = no encoder info).
func p(token, name string, w, h int) *Profile {
	profile := &Profile{Token: token, Name: name}
	if w > 0 {
		profile.VideoEncoderConfiguration = &VideoEncoderConfiguration{
			Resolution: &VideoResolution{Width: w, Height: h},
		}
	}

	return profile
}

func TestSelectMainProfile(t *testing.T) {
	tests := []struct {
		name     string
		profiles []*Profile
		want     string
	}{
		{
			name:     "empty list",
			profiles: nil,
			want:     "",
		},
		{
			name:     "single profile",
			profiles: []*Profile{p("main", "Main", 1920, 1080)},
			want:     "main",
		},
		{
			name: "sub listed first (resolution rules)",
			profiles: []*Profile{
				p("sub_1", "Sub", 640, 360),
				p("main_1", "Main", 1920, 1080),
			},
			want: "main_1",
		},
		{
			name: "main listed first",
			profiles: []*Profile{
				p("main_1", "Main", 1920, 1080),
				p("sub_1", "Sub", 640, 360),
			},
			want: "main_1",
		},
		{
			name: "tie broken by main naming hint",
			profiles: []*Profile{
				p("profile_1", "Quality1", 1920, 1080),
				p("mainStream", "Quality2", 1920, 1080),
			},
			want: "mainStream",
		},
		{
			name: "tie: sub-hint loses to neutral",
			profiles: []*Profile{
				p("extra_1", "Extra", 1920, 1080),
				p("profile_1", "Quality1", 1920, 1080),
			},
			want: "profile_1",
		},
		{
			name: "chinese hints: 主码流 wins tie",
			profiles: []*Profile{
				p("tok_a", "高清子码流", 1280, 720),
				p("tok_b", "主码流", 1280, 720),
			},
			want: "tok_b",
		},
		{
			name: "no resolution info falls back to list order",
			profiles: []*Profile{
				p("first", "Whatever", 0, 0),
				p("second", "Main", 0, 0),
			},
			want: "first",
		},
		{
			name:     "nil profile entries are safe",
			profiles: []*Profile{nil, p("main", "Main", 640, 480)},
			want:     "main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectMainProfile(tt.profiles); got != tt.want {
				t.Errorf("SelectMainProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSelectSubProfile(t *testing.T) {
	main1920 := p("main_1", "Main", 1920, 1080)

	tests := []struct {
		name      string
		profiles  []*Profile
		mainToken string
		want      string
	}{
		{
			name:      "picks largest strictly-smaller profile",
			profiles:  []*Profile{main1920, p("sub_hd", "HD", 1280, 720), p("sub_sd", "SD", 640, 360)},
			mainToken: "main_1",
			want:      "sub_hd",
		},
		{
			name:      "same resolution as main is not a substream (dual token)",
			profiles:  []*Profile{main1920, p("token_b", "MainB", 1920, 1080), p("sub", "Sub", 640, 360)},
			mainToken: "main_1",
			want:      "sub",
		},
		{
			name:      "only same-resolution profiles means no substream",
			profiles:  []*Profile{main1920, p("token_b", "MainB", 1920, 1080)},
			mainToken: "main_1",
			want:      "",
		},
		{
			name:      "single profile has no substream",
			profiles:  []*Profile{main1920},
			mainToken: "main_1",
			want:      "",
		},
		{
			name:      "empty list",
			profiles:  nil,
			mainToken: "main_1",
			want:      "",
		},
		{
			name:      "unknown main resolution falls back to list order",
			profiles:  []*Profile{p("first", "A", 0, 0), p("second", "B", 0, 0)},
			mainToken: "first",
			want:      "second",
		},
		{
			name:      "unknown resolutions besides main return nothing rankable",
			profiles:  []*Profile{p("main", "M", 1920, 1080), p("other", "O", 0, 0)},
			mainToken: "main",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SelectSubProfile(tt.profiles, tt.mainToken); got != tt.want {
				t.Errorf("SelectSubProfile() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSelectProfilesRoundTrip mirrors the documented usage: pick main, then
// sub, on a realistic 3-profile camera.
func TestSelectProfilesRoundTrip(t *testing.T) {
	profiles := []*Profile{
		p("profile_2", "substream", 640, 360),
		p("profile_1", "hq", 2560, 1440),
		p("profile_3", "mq", 1280, 720),
	}

	if got := SelectMainProfile(profiles); got != "profile_1" {
		t.Fatalf("SelectMainProfile() = %q, want profile_1", got)
	}

	if got := SelectSubProfile(profiles, "profile_1"); got != "profile_3" {
		t.Fatalf("SelectSubProfile() = %q, want profile_3", got)
	}
}
