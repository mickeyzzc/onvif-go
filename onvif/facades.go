package onvif

import (
	"github.com/mickeyzzc/onvif-go/v2/device"
	"github.com/mickeyzzc/onvif-go/v2/deviceio"
	"github.com/mickeyzzc/onvif-go/v2/events"
	"github.com/mickeyzzc/onvif-go/v2/imaging"
	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/media"
	"github.com/mickeyzzc/onvif-go/v2/ptz"
	"github.com/mickeyzzc/onvif-go/v2/security"
)

// Service facades (v2): the Client is a connection + configuration holder
// that implements internal/api.Caller; every ONVIF operation lives on the
// service object that owns it, mirroring the ONVIF service model:
//
//	client.Device().GetDeviceInformation(ctx)
//	client.Media().GetProfiles(ctx)
//	client.PTZ().ContinuousMove(ctx, token, speed, timeout)
//
// The service objects are long-lived instances created once by the client —
// accessors return the same pointers, so a service may hold its own state
// (the capabilities cache lives on the device service). Sharing a single
// Client across goroutines remains safe without facade-level locking.

// Device returns the Device service facade.
func (c *Client) Device() *device.Service { return c.deviceSvc }

// Media returns the Media service facade.
func (c *Client) Media() *media.Service { return c.mediaSvc }

// PTZ returns the PTZ service facade.
func (c *Client) PTZ() *ptz.Service { return c.ptzSvc }

// Imaging returns the Imaging service facade.
func (c *Client) Imaging() *imaging.Service { return c.imagingSvc }

// Events returns the Events service facade.
func (c *Client) Events() *events.Service { return c.eventsSvc }

// DeviceIO returns the Device IO service facade.
func (c *Client) DeviceIO() *deviceio.Service { return c.deviceioSvc }

// Security returns the Security (user management) service facade.
func (c *Client) Security() *security.Service { return c.securitySvc }

// EndpointFor returns the resolved endpoint for a service, falling back to
// the device endpoint when Initialize has not resolved (or the device does
// not advertise) a service-specific address. Implements internal/api.Caller.
func (c *Client) EndpointFor(svc api.Service) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var endpoint string
	switch svc {
	case api.ServiceMedia:
		endpoint = c.mediaEndpoint
	case api.ServicePTZ:
		endpoint = c.ptzEndpoint
	case api.ServiceImaging:
		endpoint = c.imagingEndpoint
	case api.ServiceEvents:
		endpoint = c.eventEndpoint
	default:
		return c.endpoint
	}

	if endpoint == "" {
		return c.endpoint
	}

	return endpoint
}

// SetServiceEndpoint pins the endpoint for a service, bypassing endpoint
// resolution. The v1 per-service setters (e.g. SetEventEndpoint) collapse
// into this one method; pinning the device service replaces the whole
// client endpoint.
func (c *Client) SetServiceEndpoint(svc api.Service, endpoint string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch svc {
	case api.ServiceMedia:
		c.mediaEndpoint = endpoint
	case api.ServicePTZ:
		c.ptzEndpoint = endpoint
	case api.ServiceImaging:
		c.imagingEndpoint = endpoint
	case api.ServiceEvents:
		c.eventEndpoint = endpoint
	default:
		c.endpoint = endpoint
	}
}
