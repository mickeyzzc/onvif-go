// Package testdata provides synthetic camera configuration data for testing.
//
// All entries are FABRICATED: RFC 5737 documentation IPs (192.0.2.x /
// 198.51.100.x), synthetic UUIDs, and generic vendor/model names. No real
// network data is committed to this repository (see testdata/README.md).
package testdata

// DiscoveredCamera represents a camera found on the network
type DiscoveredCamera struct {
	ID            int
	Endpoint      string
	XAddrs        []string
	Manufacturer  string
	Model         string
	IP            string
	Port          int
	Profiles      []string
	SupportsHTTPS bool
}

// TestCameras contains synthetic cameras for testing
var TestCameras = []DiscoveredCamera{
	{
		ID:           1,
		Endpoint:     "urn:uuid:a0000000-0000-4000-8000-000000000001",
		XAddrs:       []string{"http://192.0.2.11:8000/onvif/device_service"},
		Manufacturer: "VendorA",
		Model:        "PTZ-Cam-A1",
		IP:           "192.0.2.11",
		Port:         8000,
		Profiles:     []string{"Streaming", "T"},
	},
	{
		ID:            2,
		Endpoint:      "urn:uuid:a0000000-0000-4000-8000-000000000002",
		XAddrs:        []string{"http://192.0.2.12/onvif/device_service", "https://192.0.2.12/onvif/device_service"},
		Manufacturer:  "VendorB",
		Model:         "Dome-B2",
		IP:            "192.0.2.12",
		Port:          80,
		Profiles:      []string{"Streaming", "G", "T"},
		SupportsHTTPS: true,
	},
	{
		ID:           3,
		Endpoint:     "urn:uuid:a0000000-0000-4000-8000-000000000003",
		XAddrs:       []string{"http://198.51.100.13/onvif/device_service"},
		Manufacturer: "VendorC",
		Model:        "Corner-C3",
		IP:           "198.51.100.13",
		Port:         80,
		Profiles:     []string{"Streaming", "G", "M", "T"},
	},
	{
		ID:            4,
		Endpoint:      "urn:uuid:a0000000-0000-4000-8000-000000000004",
		XAddrs:        []string{"http://198.51.100.14:8000/onvif/device_service", "https://198.51.100.14:8000/onvif/device_service"},
		Manufacturer:  "VendorA",
		Model:         "Fisheye-A4",
		IP:            "198.51.100.14",
		Port:          8000,
		Profiles:      []string{"Streaming", "G", "T", "M"},
		SupportsHTTPS: true,
	},
}

// GetCameraByManufacturer returns cameras filtered by manufacturer
func GetCameraByManufacturer(manufacturer string) []DiscoveredCamera {
	var result []DiscoveredCamera
	for _, cam := range TestCameras {
		if cam.Manufacturer == manufacturer {
			result = append(result, cam)
		}
	}
	return result
}

// GetCameraByProfile returns cameras that support a specific profile
func GetCameraByProfile(profile string) []DiscoveredCamera {
	var result []DiscoveredCamera
	for _, cam := range TestCameras {
		for _, p := range cam.Profiles {
			if p == profile {
				result = append(result, cam)
				break
			}
		}
	}
	return result
}

// GetHTTPSCameras returns cameras that support HTTPS
func GetHTTPSCameras() []DiscoveredCamera {
	var result []DiscoveredCamera
	for _, cam := range TestCameras {
		if cam.SupportsHTTPS {
			result = append(result, cam)
		}
	}
	return result
}
