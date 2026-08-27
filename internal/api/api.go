// Package api defines the contract between the root Client and the
// per-domain service packages. It exists to break the v1 import cycle:
// services cannot import the root package (the root provides the facade
// accessors), so they depend on this tiny leaf interface instead.
package api

import "context"

// Service identifies an ONVIF service for endpoint resolution.
type Service string

const (
	// ServiceDevice is the device service (tds).
	ServiceDevice Service = "device"
	// ServiceMedia is the media service (trt).
	ServiceMedia Service = "media"
	// ServicePTZ is the PTZ service (tptz).
	ServicePTZ Service = "ptz"
	// ServiceImaging is the imaging service (timg).
	ServiceImaging Service = "imaging"
	// ServiceEvents is the events service (tev).
	ServiceEvents Service = "events"
)

// Caller performs an authenticated SOAP call through the client's auth
// configuration (ladder, clock skew) and resolves service endpoints.
// The root Client implements it.
type Caller interface {
	// Call performs one SOAP exchange, applying the auth fallback ladder
	// and wrapping auth-class failures so errors.Is(err, ErrUnauthorized)
	// holds.
	Call(ctx context.Context, endpoint, action string, request, response interface{}) error

	// EndpointFor returns the resolved endpoint for a service, falling back
	// to the device endpoint when Initialize has not resolved (or the device
	// does not advertise) a service-specific address.
	EndpointFor(svc Service) string
}
