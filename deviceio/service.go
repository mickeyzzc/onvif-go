package deviceio

import "github.com/mickeyzzc/onvif-go/v2/internal/api"

// Service covers the ONVIF device IO service (tmd): relay outputs, digital I/O.
type Service struct {
	c api.Caller
}

// New creates a device-IO service bound to a caller.
func New(c api.Caller) *Service { return &Service{c: c} }
