// Package discovery provides ONVIF device discovery functionality using WS-Discovery protocol.
package discovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/wsdiscovery"
)

const (
	// WS-Discovery multicast address.
	multicastAddr = "239.255.255.250:3702"
	// UUID generation constants.
	uuidMod1000  = 1000
	uuidMod10000 = 10000
)

// Device represents a discovered ONVIF device.
type Device struct {
	// EndpointRef is the WS-Discovery endpoint address, conventionally a
	// "urn:uuid:..." form. It is a stable transport-level identifier for the
	// discovery protocol — it is NOT the device serial number. Correlating a
	// camera across protocols (ONVIF vs GB28181) must use
	// Info.SerialNumber; comparing EndpointRef against a serial silently
	// never matches.
	EndpointRef string

	// Name, Hardware and Location are the structured forms of the well-known
	// ONVIF scopes (see ParseScopes), filled wherever scopes are parsed.
	// Empty when the device does not advertise them.
	Name     string
	Hardware string
	Location string

	// XAddrs contains the device service addresses
	XAddrs []string

	// Types contains the device types
	Types []string

	// Scopes contains the device scopes (name, location, etc.)
	Scopes []string

	// Metadata version
	MetadataVersion int

	// Info carries identity fields (manufacturer, model, serial) when a
	// probe or enrichment round fetched them via GetDeviceInformation.
	// nil when unavailable.
	Info *DeviceInfo
}

// ProbeMatch represents a WS-Discovery probe match (the shared
// wsdiscovery codec type; the client and the device-side responder
// parse each other's messages with the same definitions).
type ProbeMatch = wsdiscovery.Match

// DiscoverOptions contains options for device discovery.
type DiscoverOptions struct {
	// NetworkInterface specifies the network interface to use for multicast.
	// If empty, the system will choose the default interface.
	// Examples: "eth0", "wlan0", "192.168.1.100"
	NetworkInterface string

	// Context and timeout are handled by the caller
}

// Discover performs ONVIF device discovery using WS-Discovery protocol.
// For advanced options like specifying a network interface, use DiscoverWithOptions.
func Discover(ctx context.Context, timeout time.Duration) ([]*Device, error) {
	return DiscoverWithOptions(ctx, timeout, &DiscoverOptions{})
}

// DiscoverWithOptions discovers ONVIF devices with custom options.
//
//nolint:gocyclo // Discovery function has high complexity due to multiple network operations
func DiscoverWithOptions(ctx context.Context, timeout time.Duration, opts *DiscoverOptions) ([]*Device, error) {
	if opts == nil {
		opts = &DiscoverOptions{}
	}

	// Create UDP connection for multicast
	addr, err := net.ResolveUDPAddr("udp", multicastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve multicast address: %w", err)
	}

	// Get the network interface to use
	var iface *net.Interface
	if opts.NetworkInterface != "" {
		iface, err = resolveNetworkInterface(opts.NetworkInterface)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve network interface: %w", err)
		}
	}

	conn, err := net.ListenMulticastUDP("udp", iface, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on multicast address: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	// Set read deadline
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Generate message ID
	messageID := generateUUID()

	// Send probe message (shared codec with the device-side responder)
	if _, err := conn.WriteToUDP(wsdiscovery.BuildProbe(messageID), addr); err != nil {
		return nil, fmt.Errorf("failed to send probe message: %w", err)
	}

	// Collect responses
	devices := make(map[string]*Device)
	const maxUDPPacketSize = 8192
	buffer := make([]byte, maxUDPPacketSize)

	// Read responses until timeout or context cancellation
	for {
		select {
		case <-ctx.Done():
			return deviceMapToSlice(devices), ctx.Err()
		default:
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					// Timeout reached, return collected devices
					return deviceMapToSlice(devices), nil
				}

				return deviceMapToSlice(devices), fmt.Errorf("failed to read UDP response: %w", err)
			}

			// Parse response
			device, err := parseProbeResponse(buffer[:n])
			if err != nil {
				// Skip invalid responses
				continue
			}

			// Add to devices map (deduplicate by endpoint)
			if device != nil && device.EndpointRef != "" {
				devices[device.EndpointRef] = device
			}
		}
	}
}

// parseProbeResponse parses a WS-Discovery probe response through the
// shared wsdiscovery codec.
func parseProbeResponse(data []byte) (*Device, error) {
	matches, err := wsdiscovery.ParseProbeMatches(data)
	if err != nil {
		if errors.Is(err, wsdiscovery.ErrNoMatches) {
			return nil, fmt.Errorf("%w", ErrNoProbeMatches)
		}

		return nil, fmt.Errorf("failed to unmarshal probe response: %w", err)
	}

	// Take the first probe match
	match := matches[0]

	device := &Device{
		EndpointRef:     match.EndpointRef,
		XAddrs:          parseSpaceSeparated(match.XAddrs),
		Types:           parseSpaceSeparated(match.Types),
		Scopes:          parseSpaceSeparated(match.Scopes),
		MetadataVersion: match.MetadataVersion,
	}
	device.Name, device.Hardware, device.Location = scopeFields(device.Scopes)

	return device, nil
}

// scopeFields is the ParseScopes adapter for direct struct filling.
func scopeFields(scopes []string) (name, hardware, location string) {
	info := ParseScopes(scopes)

	return info.Name, info.Hardware, info.Location
}

// parseSpaceSeparated parses a space-separated string into a slice.
func parseSpaceSeparated(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}

	return strings.Fields(s)
}

// deviceMapToSlice converts a map of devices to a slice.
func deviceMapToSlice(m map[string]*Device) []*Device {
	devices := make([]*Device, 0, len(m))
	for _, device := range m {
		devices = append(devices, device)
	}

	return devices
}

// generateUUID generates a simple UUID (not cryptographically secure).
func generateUUID() string {
	now := time.Now()
	nanos := now.UnixNano()
	secs := now.Unix()

	return fmt.Sprintf("%d-%d-%d-%d-%d",
		nanos,
		secs,
		nanos%uuidMod1000,
		secs%uuidMod1000,
		nanos%uuidMod10000)
}

// resolveNetworkInterface resolves a network interface by name or IP address.
//
//nolint:gocyclo,gocognit // Network interface resolution has high complexity due to multiple validation paths
func resolveNetworkInterface(ifaceSpec string) (*net.Interface, error) {
	// Try to get interface by name (e.g., "eth0", "wlan0")
	if iface, err := net.InterfaceByName(ifaceSpec); err == nil {
		return iface, nil
	}

	// Try to parse as IP address and find the interface
	if ip := net.ParseIP(ifaceSpec); ip != nil {
		interfaces, err := net.Interfaces()
		if err != nil {
			return nil, fmt.Errorf("failed to list network interfaces: %w", err)
		}

		for _, iface := range interfaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				switch v := addr.(type) {
				case *net.IPNet:
					if v.IP.Equal(ip) {
						return &iface, nil
					}
				case *net.IPAddr:
					if v.IP.Equal(ip) {
						return &iface, nil
					}
				}
			}
		}
	}

	// List available interfaces for error message
	interfaces, err := net.Interfaces()
	if err != nil {
		interfaces = nil // Continue with empty list if we can't get interfaces
	}
	availableInterfaces := make([]string, 0)
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue // Skip this interface if we can't get addresses
		}
		ifaceInfo := iface.Name
		if len(addrs) > 0 {
			var addrStrs []string
			for _, addr := range addrs {
				addrStrs = append(addrStrs, addr.String())
			}
			ifaceInfo += " [" + strings.Join(addrStrs, ", ") + "]"
		}
		availableInterfaces = append(availableInterfaces, ifaceInfo)
	}

	return nil, fmt.Errorf("%w: %q. Available interfaces: %v", ErrNetworkInterfaceNotFound, ifaceSpec, availableInterfaces)
}

// ListNetworkInterfaces returns all available network interfaces with their addresses.
func ListNetworkInterfaces() ([]NetworkInterface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to list network interfaces: %w", err)
	}

	result := make([]NetworkInterface, 0, len(interfaces))
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var ipAddrs []string
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ipAddrs = append(ipAddrs, v.IP.String())
			case *net.IPAddr:
				ipAddrs = append(ipAddrs, v.IP.String())
			}
		}

		result = append(result, NetworkInterface{
			Name:      iface.Name,
			Addresses: ipAddrs,
			Up:        iface.Flags&net.FlagUp != 0,
			Multicast: iface.Flags&net.FlagMulticast != 0,
		})
	}

	return result, nil
}

// NetworkInterface represents a network interface.
type NetworkInterface struct {
	// Name of the interface (e.g., "eth0", "wlan0")
	Name string

	// IP addresses assigned to this interface
	Addresses []string

	// Up indicates if the interface is up
	Up bool

	// Multicast indicates if the interface supports multicast
	Multicast bool
}

// GetDeviceEndpoint extracts the primary device endpoint from XAddrs.
func (d *Device) GetDeviceEndpoint() string {
	if len(d.XAddrs) == 0 {
		return ""
	}

	// Return the first XAddr
	return d.XAddrs[0]
}

// GetName extracts the device name from scopes.
func (d *Device) GetName() string {
	if d.Name != "" {
		return d.Name
	}

	return ParseScopes(d.Scopes).Name
}

// GetLocation extracts the device location from scopes.
func (d *Device) GetLocation() string {
	if d.Location != "" {
		return d.Location
	}

	return ParseScopes(d.Scopes).Location
}
