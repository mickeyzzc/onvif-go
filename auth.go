package onvif

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/device"

	"github.com/mickeyzzc/onvif-go/v2/internal/soap"
)

// authSkewThreshold is the clock divergence beyond which digest auth is
// considered at risk: ONVIF replay windows are commonly ±5 minutes, so 2
// minutes of measured skew leaves little margin.
const authSkewThreshold = 2 * time.Minute

// WithAutoClockSkew makes Initialize measure the device clock (via an
// unauthenticated GetSystemDateAndTime) and apply the measured skew to every
// WS-Security digest timestamp before the first authenticated call. Off by
// default; measurement failure is silently skipped — a client that cannot
// read the device clock still initializes with the local clock (historical
// behavior).
func WithAutoClockSkew() ClientOption {
	return func(c *Client) {
		c.autoClockSkew = true
	}
}

// AuthDiagnosisStatus classifies the outcome of DiagnoseAuth.
type AuthDiagnosisStatus string

const (
	// AuthStatusOK: digest authentication works as-is.
	AuthStatusOK AuthDiagnosisStatus = "ok"
	// AuthStatusClockSkew: digest failed with the local clock but succeeded
	// using the device's time — the root cause is clock divergence (fix the
	// camera's NTP configuration; meanwhile WithAutoClockSkew or SetClockSkew
	// with the reported skew keeps the client working).
	AuthStatusClockSkew AuthDiagnosisStatus = "clock-skew"
	// AuthStatusBadCredentials: authentication fails regardless of clock
	// skew — wrong username/password (or the device rejects ONVIF auth
	// entirely; try WithAuthFallback).
	AuthStatusBadCredentials AuthDiagnosisStatus = "bad-credentials"
)

// AuthDiagnosis is the outcome of an authentication triage.
type AuthDiagnosis struct {
	// Status is the classification (see AuthDiagnosisStatus values).
	Status AuthDiagnosisStatus
	// ClockSkew is the measured deviceTime - localTime offset (valid when
	// the clock could be read, 0 otherwise).
	ClockSkew time.Duration
	// Detail is a human-readable explanation with remediation hints.
	Detail string
}

// clockRequest/clockResponse are compact GetSystemDateAndTime shapes; only
// the UTC time is needed for skew measurement.
type clockRequest struct {
	XMLName xml.Name `xml:"tds:GetSystemDateAndTime"`
	Xmlns   string   `xml:"xmlns:tds,attr"`
}

type clockResponse struct {
	XMLName           xml.Name `xml:"GetSystemDateAndTimeResponse"`
	SystemDateAndTime struct {
		UTCDateTime struct {
			Time struct {
				Hour   int `xml:"Hour"`
				Minute int `xml:"Minute"`
				Second int `xml:"Second"`
			} `xml:"Time"`
			Date struct {
				Year  int `xml:"Year"`
				Month int `xml:"Month"`
				Day   int `xml:"Day"`
			} `xml:"Date"`
		} `xml:"UTCDateTime"`
	} `xml:"SystemDateAndTime"`
}

// deviceUTCTime converts the parsed response fields to a time.Time in UTC.
func (r *clockResponse) deviceUTCTime() (time.Time, error) {
	dt := r.SystemDateAndTime.UTCDateTime
	t := time.Date(
		dt.Date.Year, time.Month(dt.Date.Month), dt.Date.Day,
		dt.Time.Hour, dt.Time.Minute, dt.Time.Second, 0, time.UTC)

	// A zero time or implausible year means the device reported nothing
	// usable (or the response shape differed).
	if t.IsZero() || dt.Date.Year < 2000 || dt.Date.Year > 2100 {
		return time.Time{}, errors.New("device reported no usable UTCDateTime")
	}

	return t, nil
}

// MeasureClockSkew measures deviceTime - localTime with RTT compensation:
// the request is sent unauthenticated (measurement must work before auth
// does), and the local reference is the midpoint of the request round trip,
// so network latency does not pollute the measurement.
func (c *Client) MeasureClockSkew(ctx context.Context) (time.Duration, error) {
	req := clockRequest{Xmlns: device.Namespace}
	var resp clockResponse

	localStart := time.Now()
	err := c.callWithMode(ctx, c.endpoint, "", req, &resp, AuthNone)
	localEnd := time.Now()
	if err != nil {
		return 0, fmt.Errorf("MeasureClockSkew: GetSystemDateAndTime failed: %w", err)
	}

	deviceTime, err := resp.deviceUTCTime()
	if err != nil {
		return 0, fmt.Errorf("MeasureClockSkew: %w", err)
	}

	localMid := localStart.Add(localEnd.Sub(localStart) / 2)

	return deviceTime.Sub(localMid), nil
}

// DiagnoseAuth separates the three root causes that all look like "auth
// failed": clock skew (digest timestamps rejected — fix the camera's NTP),
// wrong credentials, and devices that reject ONVIF auth entirely. It works
// even when normal calls fail, which is exactly when it is useful.
//
// The procedure:
//  1. Try digest-authenticated GetDeviceInformation — success means OK.
//  2. On auth-class failure, measure the clock skew (unauthenticated).
//  3. With significant skew (> 2 min), retry GetDeviceInformation using the
//     device's time for the digest: success confirms clock skew as the root
//     cause; failure points at the credentials.
func (c *Client) DiagnoseAuth(ctx context.Context) (*AuthDiagnosis, error) {
	probe := func(skew time.Duration) error {
		req := struct {
			XMLName xml.Name `xml:"tds:GetDeviceInformation"`
			Xmlns   string   `xml:"xmlns:tds,attr"`
		}{Xmlns: device.Namespace}

		var resp struct {
			XMLName      xml.Name `xml:"GetDeviceInformationResponse"`
			Manufacturer string   `xml:"Manufacturer"`
		}

		return c.callDigestWithSkew(ctx, c.endpoint, "", req, &resp, skew)
	}

	if err := probe(c.currentClockSkew()); err == nil {
		return &AuthDiagnosis{
			Status: AuthStatusOK,
			Detail: "digest authentication works",
		}, nil
	} else if !soap.IsAuthFailure(err) {
		return nil, fmt.Errorf("DiagnoseAuth: probe failed with a non-auth error: %w", err)
	}

	skew, skewErr := c.MeasureClockSkew(ctx)
	if skewErr != nil {
		// Clock unreadable: we can neither confirm nor exclude skew; report
		// the credential branch with the caveat spelled out.
		return &AuthDiagnosis{
			Status:    AuthStatusBadCredentials,
			ClockSkew: 0,
			Detail: fmt.Sprintf("digest rejected and device clock unreadable (%v); "+
				"verify credentials, and consider WithAuthFallback for devices that "+
				"reject ONVIF auth entirely", skewErr),
		}, nil
	}

	if skew < 0 {
		skew = -skew
	}

	if skew <= authSkewThreshold {
		return &AuthDiagnosis{
			Status:    AuthStatusBadCredentials,
			ClockSkew: skew,
			Detail: fmt.Sprintf("digest rejected but clock skew is only %v (<= %v); "+
				"credentials are most likely wrong", skew, authSkewThreshold),
		}, nil
	}

	if err := probe(c.currentClockSkew() + skew); err == nil {
		return &AuthDiagnosis{
			Status:    AuthStatusClockSkew,
			ClockSkew: skew,
			Detail: fmt.Sprintf("digest succeeds using the device clock (skew %v); "+
				"configure NTP on the camera, or keep using WithAutoClockSkew", skew),
		}, nil
	}

	return &AuthDiagnosis{
		Status:    AuthStatusBadCredentials,
		ClockSkew: skew,
		Detail: fmt.Sprintf("clock skew is %v but digest still fails with the device "+
			"time; credentials are most likely wrong", skew),
	}, nil
}

// currentClockSkew snapshots the configured skew.
func (c *Client) currentClockSkew() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.clockSkew
}

// callDigestWithSkew performs a one-off digest call with an explicit clock
// skew, without mutating the client's configured skew.
func (c *Client) callDigestWithSkew(
	ctx context.Context, endpoint, action string,
	request, response interface{}, skew time.Duration,
) error {
	c.mu.RLock()
	username, password := c.username, c.password
	httpClient := c.httpClient
	c.mu.RUnlock()

	sc := soap.NewClient(httpClient, username, password)
	sc.SetClockSkew(skew)
	sc.SetAuthMode(soap.AuthModeDigest)

	return sc.Call(ctx, endpoint, action, request, response)
}
