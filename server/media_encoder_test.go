package server

import (
	"context"
	"net"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/onvif"
)

// The ver10 encoder-configuration family + SetSynchronizationPoint and
// the tr2 encoder list — the surfaces NVRs probe before touching a
// stream (parity with the onvif-rs twin's media completion batch).

func TestMediaEncoderConfigurationFamily(t *testing.T) {
	_, client, config := startExpansionServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfgs, err := client.Media().GetVideoEncoderConfigurations(ctx)
	if err != nil {
		t.Fatalf("GetVideoEncoderConfigurations: %v", err)
	}
	if len(cfgs) == 0 {
		t.Fatal("no encoder configurations advertised")
	}
	if len(cfgs) != len(config.Profiles) {
		t.Fatalf("configurations = %d, want %d (one per profile)", len(cfgs), len(config.Profiles))
	}
	first := cfgs[0]
	wantToken := config.Profiles[0].Token + "_encoder"
	if first.Token != wantToken {
		t.Errorf("first configuration token = %q, want %q", first.Token, wantToken)
	}
	if first.Encoding != config.Profiles[0].VideoEncoder.Encoding {
		t.Errorf("encoding = %q, want %q", first.Encoding, config.Profiles[0].VideoEncoder.Encoding)
	}

	opts, err := client.Media().GetVideoEncoderConfigurationOptions(ctx, "")
	if err != nil {
		t.Fatalf("GetVideoEncoderConfigurationOptions: %v", err)
	}
	if opts.H264 == nil {
		t.Fatal("no H264 options block (the profiles are H264)")
	}
	if len(opts.H264.ResolutionsAvailable) == 0 {
		t.Error("H264 options carry no available resolutions")
	}
	if opts.QualityRange == nil {
		t.Error("no quality range")
	}
}

func TestSetSynchronizationPointFiresKeyframeHookBothFaces(t *testing.T) {
	var fired int32

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	config := createTestConfig()
	config.Port = listener.Addr().(*net.TCPAddr).Port
	config.SupportMedia2 = true

	s, err := New(config, WithKeyframeHook(func() { atomic.AddInt32(&fired, 1) }))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	ts := httptest.NewUnstartedServer(s.Handler())
	ts.Listener = listener
	ts.Start()
	t.Cleanup(ts.Close)

	client, err := onvif.NewClient(ts.URL, onvif.WithCredentials("admin", "password"))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := client.Initialize(context.Background()); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Media().SetSynchronizationPoint(ctx, config.Profiles[0].Token); err != nil {
		t.Fatalf("ver10 SetSynchronizationPoint: %v", err)
	}
	if got := atomic.LoadInt32(&fired); got != 1 {
		t.Errorf("keyframe hook fired %d times after the ver10 call, want 1", got)
	}
	if err := client.Media2().SetSynchronizationPoint(ctx, config.Profiles[0].Token); err != nil {
		t.Fatalf("tr2 SetSynchronizationPoint: %v", err)
	}
	if got := atomic.LoadInt32(&fired); got != 2 {
		t.Errorf("keyframe hook fired %d times after both calls, want 2", got)
	}
}

func TestMedia2GetVideoEncoderConfigurations(t *testing.T) {
	_, client, config := startExpansionServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfgs, err := client.Media2().GetVideoEncoderConfigurations(ctx)
	if err != nil {
		t.Fatalf("Media2 GetVideoEncoderConfigurations: %v", err)
	}
	if len(cfgs) != len(config.Profiles) {
		t.Fatalf("tr2 configurations = %d, want %d (one tr2:Configurations block per profile)", len(cfgs), len(config.Profiles))
	}
}

func TestMedia2GetServiceCapabilities(t *testing.T) {
	_, client, config := startExpansionServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	caps, err := client.Media2().GetServiceCapabilities(ctx)
	if err != nil {
		t.Fatalf("Media2 GetServiceCapabilities: %v", err)
	}
	if !caps.RTSPStreaming {
		t.Error("RTSPStreaming = false, want true")
	}
	if caps.MaximumNumberOfProfiles != len(config.Profiles) {
		t.Errorf("MaximumNumberOfProfiles = %d, want %d", caps.MaximumNumberOfProfiles, len(config.Profiles))
	}
}
