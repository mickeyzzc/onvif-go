package device

import (
	"sync"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
)

// Service covers the ONVIF device service (tds): identity, capabilities,
// system, network, storage, WiFi — plus the cached-capabilities accessor.
type Service struct {
	c api.Caller

	mu                  sync.RWMutex
	capsCache           *Capabilities
	capsCached          bool
	capsFetching        bool
	capsReady           chan struct{}
	minimalCapsFallback bool
}

// New creates a device service bound to a caller.
func New(c api.Caller) *Service { return &Service{c: c} }

// NewWithFallback creates a device service with the minimal-capabilities
// fallback enabled (see the root package's WithMinimalCapsFallback).
func NewWithFallback(c api.Caller, minimalFallback bool) *Service {
	return &Service{c: c, minimalCapsFallback: minimalFallback}
}
