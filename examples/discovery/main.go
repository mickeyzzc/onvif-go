package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/discovery"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Discovery failed: %v", err)
	}
}

func run() error {
	fmt.Println("Discovering ONVIF devices on the network...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	devices, err := discovery.Discover(ctx, 5*time.Second)
	if err != nil {
		return fmt.Errorf("discover: %w", err)
	}

	if len(devices) == 0 {
		fmt.Println("No ONVIF devices found on the network")
		return nil
	}

	fmt.Printf("\nFound %d device(s):\n\n", len(devices))

	for i, device := range devices {
		fmt.Printf("Device #%d:\n", i+1)
		fmt.Printf("  Endpoint: %s\n", device.GetDeviceEndpoint())
		fmt.Printf("  Name: %s\n", device.GetName())
		fmt.Printf("  Location: %s\n", device.GetLocation())
		fmt.Printf("  Types: %v\n", device.Types)
		fmt.Printf("  Scopes: %v\n", device.Scopes)
		fmt.Printf("  XAddrs: %v\n", device.XAddrs)
		fmt.Println()
	}

	return nil
}
