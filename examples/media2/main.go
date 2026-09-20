// Example: Media2 — the codec-agnostic video encoder configuration model.
//
// Media2 (ver20/media/wsdl) is where H.265 and AV1 live: encoder options
// come back one entry per supported encoding, with Encoding as a free
// media subtype name, and SetVideoEncoderConfiguration passes the name
// through verbatim. This example lists the configurations and, for each,
// the per-codec options (governance length and frame-rate ranges,
// available resolutions).
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/onvif"
)

func main() {
	endpoint := "http://192.0.2.100/onvif/device_service"

	client, err := onvif.NewClient(
		endpoint,
		onvif.WithCredentials("admin", "password"),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Discover service endpoints; Media2 rides the media service
	// endpoint unless pinned via SetServiceEndpoint.
	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("Service discovery failed: %v", err)
	}

	configs, err := client.Media2().GetVideoEncoderConfigurations(ctx)
	if err != nil {
		log.Fatalf("GetVideoEncoderConfigurations failed: %v", err)
	}

	fmt.Printf("Media2: %d video encoder configuration(s)\n\n", len(configs))

	for _, cfg := range configs {
		fmt.Printf("Config %q (%s %dx%d)\n",
			cfg.Token, cfg.Encoding,
			cfg.Resolution.Width, cfg.Resolution.Height)

		// One options entry per encoding the camera supports — H264,
		// H265, AV1, … each with its own ranges.
		options, err := client.Media2().GetVideoEncoderConfigurationOptions(ctx, cfg.Token, "")
		if err != nil {
			log.Printf("  options failed: %v", err)
			continue
		}
		for _, opt := range options {
			fmt.Printf("  encoding %-6s gov-length %v, frame-rate %v, %d resolution(s)\n",
				opt.Encoding, opt.GovLengthRange, opt.FrameRateRange,
				len(opt.Resolutions))
		}

		// To switch the camera to another encoding, take the matching
		// options entry as your bounds, adjust cfg (e.g. cfg.Encoding =
		// "H265"), and write it back — the encoding name is passed
		// through verbatim:
		//
		//   cfg.Encoding = "H265"
		//   cfg.GovLength = ...
		//   err := client.Media2().SetVideoEncoderConfiguration(ctx, cfg)
		fmt.Println()
	}
}
