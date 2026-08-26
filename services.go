package onvif

// Service facades.
//
// The Client is a connection + configuration holder; every ONVIF operation
// lives on the service object that owns it, mirroring the ONVIF service
// model (Device, Media, PTZ, Imaging, Events, Device IO):
//
//	client.Device().GetDeviceInformation(ctx)
//	client.Media().GetProfiles(ctx)
//	client.PTZ().ContinuousMove(ctx, token, speed, timeout)
//
// The service objects are stateless views over the shared Client — accessors
// construct them on demand, which keeps a single Client safe to share across
// goroutines without extra locking at the facade layer.

// DeviceService covers the ONVIF Device service (tds): device information,
// capabilities, network/system configuration, storage, WiFi, certificates.
type DeviceService struct {
	client *Client
}

// MediaService covers the ONVIF Media service (trt): profiles, stream and
// snapshot URIs, encoder/audio/OSD configuration, multicast.
type MediaService struct {
	client *Client
}

// PTZService covers the ONVIF PTZ service (tptz): moves, status, presets.
type PTZService struct {
	client *Client
}

// ImagingService covers the ONVIF Imaging service (timg): exposure, focus,
// imaging settings and options.
type ImagingService struct {
	client *Client
}

// EventService covers the ONVIF Events service (tev): pull-point
// subscriptions, message pulling, renewal.
type EventService struct {
	client *Client
}

// DeviceIOService covers the ONVIF Device IO service (tmd): relay outputs,
// digital inputs/outputs.
type DeviceIOService struct {
	client *Client
}

// SecurityService covers user management and access-control operations
// (tds:GetUsers family — kept apart from DeviceService because many firmwares
// gate these behind stricter auth than the rest of the device service).
type SecurityService struct {
	client *Client
}

// Device returns the Device service facade.
func (c *Client) Device() *DeviceService { return &DeviceService{client: c} }

// Media returns the Media service facade.
func (c *Client) Media() *MediaService { return &MediaService{client: c} }

// PTZ returns the PTZ service facade.
func (c *Client) PTZ() *PTZService { return &PTZService{client: c} }

// Imaging returns the Imaging service facade.
func (c *Client) Imaging() *ImagingService { return &ImagingService{client: c} }

// Events returns the Events service facade.
func (c *Client) Events() *EventService { return &EventService{client: c} }

// DeviceIO returns the Device IO service facade.
func (c *Client) DeviceIO() *DeviceIOService { return &DeviceIOService{client: c} }

// Security returns the Security (user management) service facade.
func (c *Client) Security() *SecurityService { return &SecurityService{client: c} }
