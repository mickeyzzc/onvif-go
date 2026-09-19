package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestStartTLS serves the device service over TLS: Profile T's transport
// baseline. A client dialing https (trusting the self-signed cert) reads
// the device information; plain http is refused.
func TestStartTLS(t *testing.T) {
	certFile, keyFile := generateSelfSignedCert(t)

	config := createTestConfig()
	config.Host = "127.0.0.1"
	config.Port = 0 // kernel-assigned; surfaced through ListenAddr
	config.TLSCertFile = certFile
	config.TLSKeyFile = keyFile

	srv, err := New(config)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start(ctx) }()

	deadline := time.Now().Add(5 * time.Second)
	for srv.ListenAddr() == "" {
		if time.Now().After(deadline) {
			t.Fatal("server never started listening")
		}

		time.Sleep(10 * time.Millisecond)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test trust of a fresh self-signed cert
		},
	}

	resp, err := client.Post("https://"+srv.ListenAddr()+"/onvif/device_service",
		"application/soap+xml", strings.NewReader(
			`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<GetDeviceInformation xmlns="http://www.onvif.org/ver10/device/wsdl"/>
</s:Body></s:Envelope>`))
	if err != nil {
		t.Fatalf("HTTPS request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body := make([]byte, 4096)
	n, _ := resp.Body.Read(body)

	if !strings.Contains(string(body[:n]), "Test") {
		t.Errorf("device information not served over TLS:\n%s", body[:n])
	}

	plain := &http.Client{Timeout: 2 * time.Second}
	plainResp, plainErr := plain.Post("http://"+srv.ListenAddr()+"/onvif/device_service",
		"application/soap+xml", strings.NewReader(`<x/>`))
	if plainErr == nil {
		defer func() { _ = plainResp.Body.Close() }()

		// Go's TLS stack answers cleartext requests with 400 rather than
		// a transport error — either way the SOAP service was not served.
		if plainResp.StatusCode == http.StatusOK {
			t.Error("plain HTTP unexpectedly served on the TLS listener")
		}
	}
}

// generateSelfSignedCert writes a throwaway certificate/key pair into the
// test's temp dir.
func generateSelfSignedCert(t *testing.T) (string, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048) //nolint:gosec // test-only certificate
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "onvif-go-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	dir := t.TempDir()
	certFile := filepath.Join(dir, "cert.pem")
	keyFile := filepath.Join(dir, "key.pem")

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}

	if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	return certFile, keyFile
}
