package onvif

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/mickeyzzc/onvif-go/internal/soap"
)

// Default client configuration constants.
const (
	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second
	// DefaultIdleConnTimeout is the default idle connection timeout.
	DefaultIdleConnTimeout = 90 * time.Second
	// DefaultMaxIdleConns is the default maximum idle connections.
	DefaultMaxIdleConns = 10
	// DefaultMaxIdleConnsPerHost is the default maximum idle connections per host.
	DefaultMaxIdleConnsPerHost = 5
	// NonceSize is the size of the nonce for digest authentication.
	NonceSize = 16
)

// Client represents an ONVIF client for communicating with IP cameras.
type Client struct {
	endpoint   string
	username   string
	password   string
	httpClient *http.Client
	mu         sync.RWMutex

	// Service endpoints
	mediaEndpoint   string
	ptzEndpoint     string
	imagingEndpoint string
	eventEndpoint   string

	// clockSkew is the offset (deviceTime - localTime) applied to WS-Security
	// digest timestamps. Set via SetClockSkew after measuring the device's clock
	// via GetSystemDateAndTime. Fixes Hikvision time-skew auth rejections.
	clockSkew time.Duration
}

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithInsecureSkipVerify disables TLS certificate verification.
// WARNING: Only use this for testing or with trusted cameras on private networks.
func WithInsecureSkipVerify() ClientOption {
	return func(c *Client) {
		if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
			if transport.TLSClientConfig == nil {
				transport.TLSClientConfig = &tls.Config{ //nolint:gosec // InsecureSkipVerify is intentional for testing
				}
			}
			transport.TLSClientConfig.InsecureSkipVerify = true
		}
	}
}

// WithCredentials sets the authentication credentials.
func WithCredentials(username, password string) ClientOption {
	return func(c *Client) {
		c.username = username
		c.password = password
	}
}

// NewClient creates a new ONVIF client
// The endpoint can be provided in multiple formats:
//   - Full URL: "http://192.168.1.100/onvif/device_service"
//   - IP with port: "192.168.1.100:80" (http assumed, /onvif/device_service added)
//   - IP only: "192.168.1.100" (http://IP:80/onvif/device_service used)
func NewClient(endpoint string, opts ...ClientOption) (*Client, error) {
	// Normalize endpoint to full URL
	normalizedEndpoint, err := normalizeEndpoint(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	client := &Client{
		endpoint: normalizedEndpoint,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        DefaultMaxIdleConns,
				MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
				IdleConnTimeout:     DefaultIdleConnTimeout,
			},
			// Don't follow redirects automatically
			// This prevents http:// from being silently upgraded to https://
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// normalizeEndpoint converts various endpoint formats to a full ONVIF URL.
func normalizeEndpoint(endpoint string) (string, error) {
	// Check if endpoint starts with a scheme
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		// Parse as full URL
		parsedURL, err := url.Parse(endpoint)
		if err != nil {
			return "", fmt.Errorf("failed to parse endpoint URL: %w", err)
		}
		if parsedURL.Host == "" {
			return "", fmt.Errorf("%w", ErrURLMissingHost)
		}
		// If path is empty or just "/", add default ONVIF path
		if parsedURL.Path == "" || parsedURL.Path == "/" {
			parsedURL.Path = "/onvif/device_service"
		}

		return parsedURL.String(), nil
	}

	// No scheme - treat as IP, IP:port, hostname, or hostname:port
	// Add http:// scheme and validate
	fullURL := "http://" + endpoint + "/onvif/device_service"
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return "", fmt.Errorf("invalid IP address or hostname: %w", err)
	}

	if parsedURL.Host == "" {
		return "", fmt.Errorf("%w", ErrInvalidEndpointFormat)
	}

	return fullURL, nil
}

// fixServiceURL corrects capability XAddr URLs that a camera reports
// incorrectly. Two cases are handled:
//
//  1. Localhost/loopback (127.0.0.1, 0.0.0.0, localhost, ::1): some cameras
//     report these instead of their real IP.
//
//  2. Stale routable IP: after a camera's IP changes (DHCP reassignment, AP
//     roaming), the camera's internal network config may lag and it continues
//     to advertise its OLD IP in GetCapabilities XAddrs (e.g. it is reachable
//     at .199 but reports media_service at .200). Since we reached the camera
//     via c.endpoint (the device_service URL), c.endpoint's host is guaranteed
//     to be the current correct address — so when a capability XAddr host
//     disagrees with it, we rewrite the XAddr host to match c.endpoint's host.
//
// In both cases the port from the service URL is preserved (the camera's
// service-specific port is authoritative); if the service URL omits a port,
// the client endpoint's port is used as a fallback.
func (c *Client) fixServiceURL(serviceURL string) string {
	if serviceURL == "" {
		return serviceURL
	}

	// Parse the service URL.
	parsedService, err := url.Parse(serviceURL)
	if err != nil {
		return serviceURL // Return original if parsing fails.
	}

	// Parse the client's endpoint to get the authoritative camera address.
	parsedClient, err := url.Parse(c.endpoint)
	if err != nil {
		return serviceURL // Cannot determine the correct host.
	}

	serviceHost := parsedService.Hostname()
	clientHost := parsedClient.Hostname()

	// Determine whether the service URL host needs correction. It needs
	// correction if it is a loopback address OR if it differs from the
	// device_service host we used to reach the camera (stale advertised IP).
	needsFix := isLoopbackHost(serviceHost) || serviceHost != clientHost
	if !needsFix {
		return serviceURL
	}

	// Replace the host with the client endpoint's (authoritative) host,
	// preserving the service URL's port if specified.
	servicePort := parsedService.Port()
	if servicePort != "" {
		parsedService.Host = clientHost + ":" + servicePort
	} else {
		parsedService.Host = clientHost
		// Use client's port if the service doesn't specify one.
		if clientPort := parsedClient.Port(); clientPort != "" {
			parsedService.Host = clientHost + ":" + clientPort
		}
	}

	return parsedService.String()
}

// isLoopbackHost reports whether the given host is a loopback/localhost address.
func isLoopbackHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" || host == "::1"
}

// Initialize discovers and initializes service endpoints.
func (c *Client) Initialize(ctx context.Context) error {
	// Get device information and capabilities
	capabilities, err := c.GetCapabilities(ctx)
	if err != nil {
		return fmt.Errorf("failed to get capabilities: %w", err)
	}

	// Extract service endpoints and fix incorrect addresses (localhost or stale
	// advertised IPs after the camera roamed to a new address).
	if capabilities.Media != nil && capabilities.Media.XAddr != "" {
		c.mediaEndpoint = c.fixServiceURL(capabilities.Media.XAddr)
	}
	if capabilities.PTZ != nil && capabilities.PTZ.XAddr != "" {
		c.ptzEndpoint = c.fixServiceURL(capabilities.PTZ.XAddr)
	}
	if capabilities.Imaging != nil && capabilities.Imaging.XAddr != "" {
		c.imagingEndpoint = c.fixServiceURL(capabilities.Imaging.XAddr)
	}
	if capabilities.Events != nil && capabilities.Events.XAddr != "" {
		c.eventEndpoint = c.fixServiceURL(capabilities.Events.XAddr)
	}

	return nil
}

// Endpoint returns the device endpoint.
func (c *Client) Endpoint() string {
	return c.endpoint
}

// SetCredentials updates the authentication credentials.
func (c *Client) SetCredentials(username, password string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.username = username
	c.password = password
}

// GetCredentials returns the current credentials.
func (c *Client) GetCredentials() (username, password string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.username, c.password
}

// SetClockSkew sets the clock offset (deviceTime - localTime) used for WS-Security
// digest auth timestamps. Call this after measuring the device's clock via
// GetSystemDateAndTime to fix auth failures from clock divergence (Hikvision).
func (c *Client) SetClockSkew(skew time.Duration) {
	c.mu.Lock()
	c.clockSkew = skew
	c.mu.Unlock()
}

// newSoapClient creates a soap.Client with the current credentials and clock
// skew applied. All device/media/ptz/imaging methods should use this instead of
// soap.NewClient directly so the skew propagates to every authenticated call.
func (c *Client) newSoapClient(username, password string) *soap.Client {
	sc := soap.NewClient(c.httpClient, username, password)
	c.mu.RLock()
	sc.SetClockSkew(c.clockSkew)
	c.mu.RUnlock()
	return sc
}
