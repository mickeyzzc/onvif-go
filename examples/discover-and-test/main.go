package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mickeyzzc/onvif-go"
	"github.com/mickeyzzc/onvif-go/discovery"
)

func main() {
	fmt.Println("🔍 Discovering ONVIF cameras on the network...")
	fmt.Println("This may take a few seconds...")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	devices, err := discovery.Discover(ctx, 10*time.Second)
	if err != nil {
		log.Fatalf("❌ Discovery failed: %v", err)
	}

	if len(devices) == 0 {
		fmt.Println("❌ No ONVIF cameras found on the network")
		fmt.Println("💡 Make sure:")
		fmt.Println("   - Camera is powered on and connected to the network")
		fmt.Println("   - ONVIF is enabled on the camera")
		fmt.Println("   - You're on the same network segment as the camera")
		fmt.Println("   - Camera IP 192.168.1.201 is reachable (try: ping 192.168.1.201)")
		return
	}

	fmt.Printf("✅ Found %d camera(s):\n\n", len(devices))

	var targetDevice *discovery.Device
	for i, device := range devices {
		fmt.Printf("📹 Camera #%d:\n", i+1)
		fmt.Printf("   Endpoint:  %s\n", device.GetDeviceEndpoint())
		fmt.Printf("   Name:      %s\n", device.GetName())
		fmt.Printf("   Location:  %s\n", device.GetLocation())
		fmt.Printf("   Types:     %v\n", device.Types)
		fmt.Printf("   XAddrs:    %v\n", device.XAddrs)
		fmt.Println()

		// Check if this is our target camera (192.168.1.201)
		endpoint := device.GetDeviceEndpoint()
		if len(endpoint) > 7 {
			// Simple check if endpoint contains the IP
			if len(endpoint) > 20 && (endpoint[7:20] == "192.168.1.201" || endpoint[7:21] == "192.168.1.201:") {
				targetDevice = device
			}
		}
	}

	if targetDevice == nil {
		fmt.Println("⚠️  Camera at 192.168.1.201 was not discovered")
		fmt.Println("💡 You can still try to connect manually with the correct endpoint")
		return
	}

	// Now try to connect to the discovered camera
	fmt.Printf("\n🎯 Found target camera at 192.168.1.201\n")
	fmt.Printf("Endpoint: %s\n", targetDevice.GetDeviceEndpoint())
	fmt.Println()

	// Test connection with credentials
	username := "service"
	password := "Service.1234"

	fmt.Println("📡 Connecting with credentials...")
	client, err := onvif.NewClient(
		targetDevice.GetDeviceEndpoint(),
		onvif.WithCredentials(username, password),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	ctx2 := context.Background()

	// Get device information
	fmt.Println("🔍 Retrieving device information...")
	info, err := client.GetDeviceInformation(ctx2)
	if err != nil {
		log.Fatalf("❌ Failed to get device information: %v\n\n💡 Possible issues:\n  - Wrong username or password\n  - Camera requires different authentication\n  - Try username/password combinations like: admin/admin, admin/12345, etc.\n", err)
	}

	fmt.Printf("\n✅ Device Information:\n")
	fmt.Printf("  Manufacturer:    %s\n", info.Manufacturer)
	fmt.Printf("  Model:           %s\n", info.Model)
	fmt.Printf("  Firmware:        %s\n", info.FirmwareVersion)
	fmt.Printf("  Serial Number:   %s\n", info.SerialNumber)
	fmt.Printf("  Hardware ID:     %s\n", info.HardwareID)

	// Initialize client (discover service endpoints)
	fmt.Println("\n🔧 Initializing client and discovering services...")
	if err := client.Initialize(ctx2); err != nil {
		log.Fatalf("❌ Failed to initialize client: %v", err)
	}
	fmt.Println("✅ Services discovered successfully")

	// Get capabilities
	fmt.Println("\n🎯 Getting device capabilities...")
	caps, err := client.GetCapabilities(ctx2)
	if err != nil {
		log.Printf("⚠️  Failed to get capabilities: %v", err)
	} else {
		fmt.Println("✅ Supported Services:")
		if caps.Device != nil {
			fmt.Println("  ✓ Device Service")
		}
		if caps.Media != nil {
			fmt.Println("  ✓ Media Service (Streaming)")
		}
		if caps.PTZ != nil {
			fmt.Println("  ✓ PTZ Service (Pan/Tilt/Zoom)")
		}
		if caps.Imaging != nil {
			fmt.Println("  ✓ Imaging Service")
		}
		if caps.Events != nil {
			fmt.Println("  ✓ Event Service")
		}
		if caps.Analytics != nil {
			fmt.Println("  ✓ Analytics Service")
		}
	}

	// Get media profiles
	fmt.Println("\n📹 Retrieving media profiles...")
	profiles, err := client.GetProfiles(ctx2)
	if err != nil {
		log.Fatalf("❌ Failed to get profiles: %v", err)
	}

	fmt.Printf("\n✅ Found %d profile(s):\n", len(profiles))
	for i, profile := range profiles {
		fmt.Printf("\n📺 Profile #%d:\n", i+1)
		fmt.Printf("  Token:     %s\n", profile.Token)
		fmt.Printf("  Name:      %s\n", profile.Name)

		if profile.VideoEncoderConfiguration != nil {
			fmt.Printf("  Encoding:  %s\n", profile.VideoEncoderConfiguration.Encoding)
			if profile.VideoEncoderConfiguration.Resolution != nil {
				fmt.Printf("  Resolution: %dx%d\n",
					profile.VideoEncoderConfiguration.Resolution.Width,
					profile.VideoEncoderConfiguration.Resolution.Height)
			}
			fmt.Printf("  Quality:   %.1f\n", profile.VideoEncoderConfiguration.Quality)
			if profile.VideoEncoderConfiguration.RateControl != nil {
				fmt.Printf("  Frame Rate: %d fps\n", profile.VideoEncoderConfiguration.RateControl.FrameRateLimit)
				fmt.Printf("  Bitrate:   %d kbps\n", profile.VideoEncoderConfiguration.RateControl.BitrateLimit)
			}
		}

		if profile.PTZConfiguration != nil {
			fmt.Printf("  PTZ:       Enabled\n")
		}

		// Get stream URI
		streamURI, err := client.GetStreamURI(ctx2, profile.Token)
		if err != nil {
			fmt.Printf("  Stream URI: ❌ Error - %v\n", err)
		} else {
			fmt.Printf("  Stream URI: %s\n", streamURI.URI)
			fmt.Printf("  📱 Use this URL in VLC or other RTSP player\n")
		}

		// Get snapshot URI
		snapshotURI, err := client.GetSnapshotURI(ctx2, profile.Token)
		if err != nil {
			fmt.Printf("  Snapshot URI: ❌ Error - %v\n", err)
		} else {
			fmt.Printf("  Snapshot URI: %s\n", snapshotURI.URI)
			fmt.Printf("  🌐 You can open this URL in a browser\n")
		}
	}

	// Test PTZ if available
	if len(profiles) > 0 {
		fmt.Println("\n🎮 Testing PTZ capabilities...")
		profileToken := profiles[0].Token

		status, err := client.GetStatus(ctx2, profileToken)
		if err != nil {
			fmt.Printf("⚠️  PTZ not supported or error: %v\n", err)
		} else {
			fmt.Println("✅ PTZ is supported!")
			if status.Position != nil && status.Position.PanTilt != nil {
				fmt.Printf("  Current Position: Pan=%.3f, Tilt=%.3f\n",
					status.Position.PanTilt.X,
					status.Position.PanTilt.Y)
			}
			if status.Position != nil && status.Position.Zoom != nil {
				fmt.Printf("  Current Zoom: %.3f\n", status.Position.Zoom.X)
			}

			// Get presets
			presets, err := client.GetPresets(ctx2, profileToken)
			if err != nil {
				fmt.Printf("  Presets: ❌ Error - %v\n", err)
			} else {
				fmt.Printf("  Available Presets: %d\n", len(presets))
				for _, preset := range presets {
					fmt.Printf("    - %s (Token: %s)\n", preset.Name, preset.Token)
				}
			}
		}
	}

	// Test Imaging if available
	if len(profiles) > 0 && profiles[0].VideoSourceConfiguration != nil {
		fmt.Println("\n🎨 Testing Imaging capabilities...")
		videoSourceToken := profiles[0].VideoSourceConfiguration.SourceToken

		settings, err := client.GetImagingSettings(ctx2, videoSourceToken)
		if err != nil {
			fmt.Printf("⚠️  Imaging settings not available: %v\n", err)
		} else {
			fmt.Println("✅ Current Imaging Settings:")
			if settings.Brightness != nil {
				fmt.Printf("  Brightness:     %.1f\n", *settings.Brightness)
			}
			if settings.Contrast != nil {
				fmt.Printf("  Contrast:       %.1f\n", *settings.Contrast)
			}
			if settings.ColorSaturation != nil {
				fmt.Printf("  Saturation:     %.1f\n", *settings.ColorSaturation)
			}
			if settings.Sharpness != nil {
				fmt.Printf("  Sharpness:      %.1f\n", *settings.Sharpness)
			}
			if settings.Exposure != nil {
				fmt.Printf("  Exposure Mode:  %s\n", settings.Exposure.Mode)
			}
			if settings.Focus != nil {
				fmt.Printf("  Focus Mode:     %s\n", settings.Focus.AutoFocusMode)
			}
			if settings.WhiteBalance != nil {
				fmt.Printf("  White Balance:  %s\n", settings.WhiteBalance.Mode)
			}
		}
	}

	fmt.Println("\n✅ All tests completed successfully!")
	fmt.Println("\n💡 Next steps:")
	fmt.Println("  - Use the stream URI in VLC to view the live feed")
	fmt.Println("  - Open the snapshot URI in a browser to see still images")
	fmt.Println("  - Use the PTZ controls to move the camera (if supported)")
	fmt.Println("  - Adjust imaging settings for better image quality")
}
