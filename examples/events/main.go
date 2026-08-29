// Event subscription example: create a PullPoint subscription through
// the managed event stream, print incoming notifications (motion, digital
// input, analytics…), and shut down cleanly after 30 seconds.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/onvif"
)

func main() {
	endpoint := "http://192.168.1.100/onvif/device_service"
	username := "admin"
	password := "password"

	fmt.Println("Connecting to ONVIF camera...")

	client, err := onvif.NewClient(
		endpoint,
		onvif.WithCredentials(username, password),
		onvif.WithTimeout(30*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := client.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	fmt.Println("Subscribing to events (PullPoint)...")

	// The managed stream handles the whole lifecycle: pull-point creation,
	// long-poll pulling, renewal before expiry, and best-effort unsubscribe
	// on shutdown. Panics in the handler are isolated.
	stream, err := client.Events().SubscribeEvents(ctx, func(msg onvif.NotificationMessage) {
		fmt.Printf("[%s] topic=%s op=%s\n", msg.Message.UtcTime.Format(time.RFC3339), msg.Topic, msg.Message.PropertyOperation)

		for _, item := range msg.Message.Data {
			fmt.Printf("    %s = %s\n", item.Name, item.Value)
		}
	}, &onvif.SubscribeEventsOptions{
		// Empty Filter subscribes to every topic the device offers.
		// Example topic filter: "tns1:VideoSource/MotionAlarm".
		SubscriptionDuration: time.Hour,
	})
	if err != nil {
		if errors.Is(err, onvif.ErrEventsNotSupported) {
			log.Fatalf("Camera does not implement the ONVIF events service")
		}

		log.Fatalf("SubscribeEvents: %v", err)
	}

	fmt.Println("Listening for events for 30s (Ctrl+C to stop early)...")

	select {
	case <-time.After(30 * time.Second):
	case <-ctx.Done():
	case <-stream.Done():
		fmt.Println("Event stream ended on its own (renewal failure or device restart)")
	}

	fmt.Println("Unsubscribing...")

	if err := stream.Unsubscribe(context.Background()); err != nil {
		log.Printf("Unsubscribe: %v", err)
	}

	// Deterministic cleanup: wait until the polling loop has fully exited.
	<-stream.Done()
	fmt.Println("Done.")
}
