// Example: the ONVIF server over TLS (Profile T baseline).
//
// Set Config.TLSCertFile and Config.TLSKeyFile (both, together) and the
// same Start call serves HTTPS instead of HTTP. Quick self-signed pair
// for trying this out:
//
//	openssl req -x509 -newkey rsa:2048 -nodes -days 30 \
//	  -keyout key.pem -out cert.pem -subj "/CN=onvif.local"
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/mickeyzzc/onvif-go/v2/server"
)

func main() {
	cert := flag.String("cert", "cert.pem", "TLS certificate file (PEM)")
	key := flag.String("key", "key.pem", "TLS private key file (PEM)")
	port := flag.Int("port", 8443, "HTTPS port (0 = kernel-assigned; then read ListenAddr)")
	flag.Parse()

	config := server.DefaultConfig()
	config.Port = *port
	config.TLSCertFile = *cert
	config.TLSKeyFile = *key

	// DefaultConfig ships empty credentials; set them explicitly
	// (or run in the documented everything-open mode on purpose).
	config.Username = "admin"
	config.Password = "change-me"

	srv, err := server.New(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("ONVIF simulator over TLS: https://0.0.0.0:%d/onvif/device_service\n", *port)
	fmt.Println("Point an ONVIF client at that endpoint (e.g. onvif.NewClient with an https:// URL).")
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println()

	// Start binds the listener explicitly, so with Port: 0 the
	// kernel-assigned address is observable via srv.ListenAddr() right
	// after Start returns — run Start in a goroutine to use that.
	// Start blocks serving until the context is cancelled.
	ctx := context.Background()
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
