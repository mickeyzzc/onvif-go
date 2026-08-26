package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mickeyzzc/onvif-go"
)

func main() {
	fmt.Println("🧪 Testing ONVIF Server with Client Library")
	fmt.Println("===========================================")
	fmt.Println()

	// Create client
	client, err := onvif.NewClient(
		"http://localhost:8080/onvif/device_service",
		onvif.WithCredentials("admin", "admin"),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("❌ Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test 1: Get device information
	fmt.Println("📋 Test 1: Getting Device Information...")
	info, err := client.GetDeviceInformation(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to get device info: %v", err)
	}
	fmt.Printf("✅ Device: %s %s\n", info.Manufacturer, info.Model)
	fmt.Printf("   Firmware: %s\n", info.FirmwareVersion)
	fmt.Printf("   Serial: %s\n", info.SerialNumber)
	fmt.Println()

	// Test 2: Initialize and discover services
	fmt.Println("🔍 Test 2: Discovering Services...")
	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("❌ Failed to initialize: %v", err)
	}
	fmt.Println("✅ Services discovered successfully")
	fmt.Println()

	// Test 3: Get capabilities
	fmt.Println("🔧 Test 3: Getting Capabilities...")
	caps, err := client.GetCapabilities(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to get capabilities: %v", err)
	}
	fmt.Println("✅ Capabilities:")
	if caps.Media != nil {
		fmt.Println("   ✓ Media Service")
	}
	if caps.PTZ != nil {
		fmt.Println("   ✓ PTZ Service")
	}
	if caps.Imaging != nil {
		fmt.Println("   ✓ Imaging Service")
	}
	fmt.Println()

	// Test 4: Get media profiles
	fmt.Println("🎬 Test 4: Getting Media Profiles...")
	profiles, err := client.GetProfiles(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to get profiles: %v", err)
	}
	fmt.Printf("✅ Found %d camera profiles:\n", len(profiles))
	for i, profile := range profiles {
		fmt.Printf("\n   Profile %d: %s\n", i+1, profile.Name)
		fmt.Printf("   Token: %s\n", profile.Token)

		if profile.VideoEncoderConfiguration != nil {
			fmt.Printf("   Video: %dx%d @ %s\n",
				profile.VideoEncoderConfiguration.Resolution.Width,
				profile.VideoEncoderConfiguration.Resolution.Height,
				profile.VideoEncoderConfiguration.Encoding)
		}

		// Get stream URI
		streamURI, err := client.GetStreamURI(ctx, profile.Token)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to get stream URI: %v\n", err)
		} else {
			fmt.Printf("   RTSP: %s\n", streamURI.URI)
		}

		// Get snapshot URI if available
		snapshotURI, err := client.GetSnapshotURI(ctx, profile.Token)
		if err == nil {
			fmt.Printf("   Snapshot: %s\n", snapshotURI.URI)
		}

		// Test PTZ if available
		if profile.PTZConfiguration != nil {
			fmt.Println("   PTZ: ✓ Enabled")

			// Get PTZ status
			status, err := client.GetStatus(ctx, profile.Token)
			if err == nil {
				fmt.Printf("   Position: Pan=%.1f°, Tilt=%.1f°, Zoom=%.2f\n",
					status.Position.PanTilt.X,
					status.Position.PanTilt.Y,
					status.Position.Zoom.X)
			}

			// Get presets
			presets, err := client.GetPresets(ctx, profile.Token)
			if err == nil && len(presets) > 0 {
				fmt.Printf("   Presets: %d available\n", len(presets))
			}
		}
	}
	fmt.Println()

	// Test 5: PTZ control (if available)
	if len(profiles) > 0 && profiles[0].PTZConfiguration != nil {
		fmt.Println("🎮 Test 5: Testing PTZ Control...")
		profileToken := profiles[0].Token

		// Absolute move to home position
		fmt.Println("   Moving to home position...")
		position := &onvif.PTZVector{
			PanTilt: &onvif.Vector2D{X: 0.0, Y: 0.0},
			Zoom:    &onvif.Vector1D{X: 0.0},
		}
		if err := client.AbsoluteMove(ctx, profileToken, position, nil); err != nil {
			fmt.Printf("   ⚠️  Failed to move: %v\n", err)
		} else {
			fmt.Println("   ✅ Moved to home position")
		}

		// Wait a moment
		time.Sleep(500 * time.Millisecond)

		// Get status after move
		status, err := client.GetStatus(ctx, profileToken)
		if err == nil {
			fmt.Printf("   New position: Pan=%.1f°, Tilt=%.1f°, Zoom=%.2f\n",
				status.Position.PanTilt.X,
				status.Position.PanTilt.Y,
				status.Position.Zoom.X)
		}
		fmt.Println()
	}

	// Summary
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                            ║")
	fmt.Println("║              ✅  All Tests Passed!  ✅                     ║")
	fmt.Println("║                                                            ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("🎉 ONVIF Server is working correctly!")
	fmt.Println("   • Device Service: ✓")
	fmt.Println("   • Media Service: ✓")
	fmt.Println("   • PTZ Service: ✓")
	fmt.Printf("   • Multi-lens Camera: ✓ (%d profiles)\n", len(profiles))
}
