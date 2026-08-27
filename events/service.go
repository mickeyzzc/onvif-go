package events

import "github.com/mickeyzzc/onvif-go/v2/internal/api"

// Service covers the ONVIF events service (tev): pull-point primitives and the
// managed subscription loop.
type Service struct {
	c api.Caller
}

// New creates an events service bound to a caller.
func New(c api.Caller) *Service { return &Service{c: c} }
