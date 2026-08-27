package media

import "github.com/mickeyzzc/onvif-go/v2/internal/api"

// Service covers the ONVIF media service (trt): profiles, stream/snapshot URIs,
// encoder/audio/OSD configuration.
type Service struct {
	c api.Caller
}

// New creates a media service bound to a caller.
func New(c api.Caller) *Service { return &Service{c: c} }
