package ptz

import "github.com/mickeyzzc/onvif-go/v2/internal/api"

// Service covers the ONVIF PTZ service (tptz): moves, status, presets.
type Service struct {
	c api.Caller
}

// New creates a PTZ service bound to a caller.
func New(c api.Caller) *Service { return &Service{c: c} }
