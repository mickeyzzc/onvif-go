package onvif

import (
	"context" //nolint:gosec // MD5 used for ONVIF digest authentication
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/mickeyzzc/onvif-go/internal/httpdigest"
)

// downloadStatusError marks a download failure caused by a specific HTTP
// status, so retry decisions are made on the typed status instead of string
// matching — an error message embeds the request URL, and a random port
// containing "401" once sent the logic down the digest path on a mere
// timeout (a real CI flake).
type downloadStatusError struct {
	status int
	err    error
}

func (e *downloadStatusError) Error() string { return e.err.Error() }
func (e *downloadStatusError) Unwrap() error { return e.err }

// DownloadFile downloads a file from the given URL with authentication.
// Supports both Basic and Digest authentication (tries basic first, falls back to digest).
// This is not an ONVIF SOAP operation — it fetches snapshot/media files from
// the URIs the device hands out, which typically require the same credentials.
func (c *Client) DownloadFile(ctx context.Context, downloadURL string) ([]byte, error) {
	// Try basic auth first
	data, err := c.downloadWithBasicAuth(ctx, downloadURL)
	if err == nil {
		return data, nil
	}

	// If basic auth fails with 401, try digest auth
	var basicStatus *downloadStatusError
	if errors.As(err, &basicStatus) && basicStatus.status == http.StatusUnauthorized {
		digestData, digestErr := c.downloadWithDigestAuth(ctx, downloadURL)
		if digestErr == nil {
			return digestData, nil
		}
		// If digest auth also fails with 401, return the original error
		var digestStatus *downloadStatusError
		if errors.As(digestErr, &digestStatus) && digestStatus.status == http.StatusUnauthorized {
			return nil, err // Return original error (both auth methods failed)
		}

		return nil, digestErr
	}

	return nil, err
}

// downloadWithBasicAuth performs an HTTP download with Basic authentication.
func (c *Client) downloadWithBasicAuth(ctx context.Context, downloadURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	username, password := c.GetCredentials()
	if username != "" {
		req.SetBasicAuth(username, password)
	}

	req.Header.Set("User-Agent", "onvif-go-client")
	req.Header.Set("Connection", "close")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyPreview, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyPreview)
		const maxBodyPreview = 200
		if len(bodyStr) > maxBodyPreview {
			bodyStr = bodyStr[:maxBodyPreview] + "..."
		}

		// Base error message for programmatic use
		errorMsg := fmt.Sprintf("download failed with status code %d", resp.StatusCode)

		// Add structured error details
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			errorMsg += ": authentication failed (401 Unauthorized); basic auth failed, trying digest auth"
		case http.StatusForbidden:
			errorMsg += ": access denied (403 Forbidden); user may not have permission to download snapshots"
		case http.StatusNotFound:
			errorMsg += ": snapshot URI not found (404); camera may have revoked the URI, try getting a fresh snapshot URI"
		}

		if bodyStr != "" {
			errorMsg += "; response: " + bodyStr
		}

		return nil, &downloadStatusError{
			status: resp.StatusCode,
			err:    fmt.Errorf("%w: %s", ErrDownloadFailed, errorMsg),
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

// downloadWithDigestAuth performs an HTTP download with Digest authentication.
func (c *Client) downloadWithDigestAuth(ctx context.Context, downloadURL string) ([]byte, error) {
	username, password := c.GetCredentials()
	if username == "" {
		return nil, fmt.Errorf("%w", ErrDigestAuthRequiresCredentials)
	}

	// Create a custom transport with digest auth
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   DefaultTimeout,
			KeepAlive: DefaultTimeout,
		}).Dial,
		MaxIdleConns:        DefaultMaxIdleConns,
		MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
		IdleConnTimeout:     DefaultIdleConnTimeout,
	}

	// Create a custom HTTP client for digest auth
	digestClient := &http.Client{
		Transport: &httpdigest.Transport{
			Transport: tr,
			Username:  username,
			Password:  password,
		},
		Timeout: DefaultTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "onvif-go-client")
	req.Header.Set("Connection", "close")

	resp, err := digestClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("digest auth request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyPreview, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyPreview)
		const maxBodyPreview = 200
		if len(bodyStr) > maxBodyPreview {
			bodyStr = bodyStr[:maxBodyPreview] + "..."
		}

		errorMsg := fmt.Sprintf("download failed with status code %d", resp.StatusCode)

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			errorMsg += ": digest authentication failed (401 Unauthorized); check camera credentials (username/password)"
		case http.StatusForbidden:
			errorMsg += ": access denied (403 Forbidden); user may not have permission to download snapshots"
		case http.StatusNotFound:
			errorMsg += ": snapshot URI not found (404); try getting a fresh snapshot URI"
		}

		if bodyStr != "" {
			errorMsg += "; response: " + bodyStr
		}

		return nil, &downloadStatusError{
			status: resp.StatusCode,
			err:    fmt.Errorf("%w: %s", ErrDownloadFailed, errorMsg),
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}
