package media

import "testing"

func TestLooseExtractURI(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "prefixed element",
			raw:  `<trt:MediaUri><tt:Uri>rtsp://cam/stream</tt:Uri></trt:MediaUri>`,
			want: "rtsp://cam/stream",
		},
		{
			name: "unprefixed element with whitespace",
			raw:  `<MediaUri><Uri>  http://cam/snap  </Uri></MediaUri>`,
			want: "http://cam/snap",
		},
		{
			name: "self-closed element",
			raw:  `<MediaUri><Uri/></MediaUri>`,
			want: "",
		},
		{
			name: "no Uri element at all",
			raw:  `<GetStreamUriResponse><Other>x</Other></GetStreamUriResponse>`,
			want: "",
		},
		{
			name: "garbage input",
			raw:  `<not-xml`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := looseExtractURI(tt.raw); got != tt.want {
				t.Errorf("looseExtractURI() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestGetStreamURIWithOptionsDefaultsToRTSP verifies that a nil Transport
// (or an empty protocol) defaults to RTSP instead of erroring.
func TestGetStreamURIWithOptionsDefaultsToRTSP(t *testing.T) {
	for _, setup := range []StreamSetup{
		{Stream: StreamRTPUnicast},
		{Stream: StreamRTPUnicast, Transport: &Transport{}},
	} {
		if err := setup.validate(); err != nil {
			t.Fatalf("validate(%+v) error = %v, want RTSP default", setup, err)
		}

		if setup.Transport.Protocol != ProtocolRTSP {
			t.Errorf("validate(%+v) left protocol %q, want %q", setup, setup.Transport.Protocol, ProtocolRTSP)
		}
	}
}
