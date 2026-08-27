package imaging

import "github.com/mickeyzzc/onvif-go/v2/internal/api"

// Service covers the ONVIF imaging service (timg): exposure, focus, imaging settings.
type Service struct {
	c api.Caller
}

// New creates an imaging service bound to a caller.
func New(c api.Caller) *Service { return &Service{c: c} }
