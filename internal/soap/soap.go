// Package soap provides SOAP client functionality for ONVIF communication.
package soap

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // SHA1 used for ONVIF digest authentication
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AuthMode selects how requests are authenticated.
type AuthMode int

const (
	// AuthModeDigest is the ONVIF default: WS-Security UsernameToken with
	// PasswordDigest (Base64(SHA1(nonce + created + password))).
	AuthModeDigest AuthMode = iota
	// AuthModePasswordText sends the password in cleartext inside the
	// WS-Security UsernameToken (PasswordText profile). Some firmwares
	// reject digest on specific services while accepting this.
	AuthModePasswordText
	// AuthModeHTTPBasic sends no WS-Security header and authenticates with
	// an HTTP Basic Authorization header instead.
	AuthModeHTTPBasic
	// AuthModeNone sends no authentication at all. Minimal embedded devices
	// (e.g. ESP32 firmwares) may reject every auth-bearing request.
	AuthModeNone
)

// Password type URIs from the WS-Security UsernameToken profile.
const (
	passwordDigestType = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordDigest" //nolint:lll // Long XML namespace
	passwordTextType   = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordText"   //nolint:lll // Long XML namespace
	nonceEncodingType  = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary"    //nolint:lll // Long XML namespace
)

// Exported aliases for consumers inside this module (the server-side
// UsernameToken validator matches on these URIs).
const (
	// PasswordDigestType marks a WS-Security Password element carrying
	// Base64(SHA1(nonce + created + password)).
	PasswordDigestType = passwordDigestType
	// PasswordTextType marks a WS-Security Password element carrying the
	// cleartext password.
	PasswordTextType = passwordTextType
)

// Envelope represents a SOAP envelope.
type Envelope struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Header  *Header  `xml:"http://www.w3.org/2003/05/soap-envelope Header,omitempty"`
	Body    Body     `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

// Header represents a SOAP header.
type Header struct {
	Security *Security `xml:"Security,omitempty"`
}

// Body represents a SOAP body.
type Body struct {
	Content interface{} `xml:",omitempty"`
	Fault   *Fault      `xml:"Fault,omitempty"`
}

// Fault represents a SOAP fault.
type Fault struct {
	XMLName xml.Name `xml:"http://www.w3.org/2003/05/soap-envelope Fault"`
	Code    string   `xml:"Code>Value"`
	Reason  string   `xml:"Reason>Text"`
	Detail  string   `xml:"Detail,omitempty"`
}

// Security represents WS-Security header.
type Security struct {
	XMLName        xml.Name       `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd Security"` //nolint:lll // Long XML namespace
	MustUnderstand string         `xml:"http://www.w3.org/2003/05/soap-envelope mustUnderstand,attr,omitempty"`
	UsernameToken  *UsernameToken `xml:"UsernameToken,omitempty"`
}

// UsernameToken represents a WS-Security username token.
type UsernameToken struct {
	XMLName  xml.Name `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd UsernameToken"` //nolint:lll // Long XML namespace
	Username string   `xml:"Username"`
	Password Password `xml:"Password"`
	Nonce    Nonce    `xml:"Nonce"`
	Created  string   `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-utility-1.0.xsd Created"`

	// CreatedVariant captures wsu:Created sent under the common
	// misspelled utility namespace (oasis-200401-wss-wssecurity-utility-…)
	// used by many community clients; encoding/xml namespace-qualifies
	// field matching, so the variant needs its own field. Emission always
	// uses the canonical Created above (omitempty keeps this one off the
	// wire when empty).
	CreatedVariant string `xml:"http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd Created,omitempty"` //nolint:lll // Long XML namespace
}

// CreatedValue returns the token's creation timestamp, accepting either
// the canonical or the misspelled utility namespace on decode.
func (t *UsernameToken) CreatedValue() string {
	if t.Created != "" {
		return t.Created
	}

	return t.CreatedVariant
}

// Password represents a WS-Security password.
type Password struct {
	Type     string `xml:"Type,attr"`
	Password string `xml:",chardata"`
}

// Nonce represents a WS-Security nonce.
type Nonce struct {
	Type  string `xml:"EncodingType,attr"`
	Nonce string `xml:",chardata"`
}

// Client represents a SOAP client.
type Client struct {
	httpClient *http.Client
	username   string
	password   string
	authMode   AuthMode
	debug      bool
	logger     func(format string, args ...interface{})
	// clockSkew is the offset (deviceTime - localTime) applied to the Created
	// timestamp in WS-Security UsernameToken digest auth. ONVIF cameras (notably
	// Hikvision) reject digests whose Created timestamp falls outside their
	// replay window (commonly ±5 min). When the NVR's clock and the camera's
	// clock diverge, every digest is rejected as "sender not authorized" — a
	// generic-looking auth failure. Setting the skew (measured via the
	// unauthenticated GetSystemDateAndTime) makes the digest use the device's
	// view of "now". Default 0 = use local clock (legacy behavior).
	clockSkew time.Duration
}

// NewClient creates a new SOAP client.
func NewClient(httpClient *http.Client, username, password string) *Client {
	return &Client{
		httpClient: httpClient,
		username:   username,
		password:   password,
		authMode:   AuthModeDigest,
		debug:      false,
		logger:     nil,
	}
}

// SetAuthMode selects the authentication mode used for subsequent calls.
// The zero value AuthModeDigest preserves the historical behavior.
func (c *Client) SetAuthMode(mode AuthMode) {
	c.authMode = mode
}

// SetDebug enables debug logging with a custom logger.
func (c *Client) SetDebug(enabled bool, logger func(format string, args ...interface{})) {
	c.debug = enabled
	c.logger = logger
}

// SetClockSkew sets the clock offset (deviceTime - localTime) used when building
// WS-Security UsernameToken digest timestamps. Call this after measuring the
// device's clock via GetSystemDateAndTime to fix auth failures caused by clock
// divergence (Hikvision time-skew issue). Pass 0 to use the local clock.
func (c *Client) SetClockSkew(skew time.Duration) {
	c.clockSkew = skew
}

// logDebugf logs debug information if debug mode is enabled.
func (c *Client) logDebugf(format string, args ...interface{}) {
	if c.debug && c.logger != nil {
		c.logger(format, args...)
	}
}

// Call makes a SOAP call to the specified endpoint.
//
// Faults are detected regardless of the HTTP status: ONVIF devices frequently
// return SOAP Faults (including NotAuthorized) with HTTP 200, and before fault
// detection existed such responses were mistaken for successful — or worse,
// silently dropped — operation results.
func (c *Client) Call(ctx context.Context, endpoint, action string, request, response interface{}) error {
	// Build SOAP envelope
	envelope := &Envelope{
		Body: Body{
			Content: request,
		},
	}

	// Add the WS-Security header for the header-based modes when credentials
	// are present (HTTP Basic and None need no SOAP header).
	if c.username != "" && c.password != "" &&
		(c.authMode == AuthModeDigest || c.authMode == AuthModePasswordText) {
		envelope.Header = &Header{
			Security: c.createSecurityHeader(),
		}
	}

	// Marshal envelope to XML
	body, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SOAP envelope: %w", err)
	}

	// Add XML declaration
	xmlBody := append([]byte(xml.Header), body...)

	// Log request if debug is enabled
	c.logDebugf("=== SOAP Request ===\nEndpoint: %s\nAction: %s\n%s\n", endpoint, action, string(xmlBody))

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(xmlBody))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	if action != "" {
		req.Header.Set("SOAPAction", action)
	}

	if c.authMode == AuthModeHTTPBasic && c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response if debug is enabled
	c.logDebugf("=== SOAP Response ===\nStatus: %d\n%s\n", resp.StatusCode, string(respBody))

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return &HTTPStatusError{Status: resp.StatusCode, Body: string(respBody)}
	}

	// If response is empty, return immediately
	if len(respBody) == 0 {
		return fmt.Errorf("%w", ErrEmptyResponseBody)
	}

	// Extract the raw Body content first: fault detection must happen for
	// every call (including void operations whose response is nil), because a
	// 200-with-Fault must never be reported as success.
	var envelopeResp struct {
		Body struct {
			Content []byte `xml:",innerxml"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal(respBody, &envelopeResp); err != nil {
		return fmt.Errorf("failed to unmarshal SOAP envelope: %w", err)
	}

	if fault := parseFault(envelopeResp.Body.Content, resp.StatusCode); fault != nil {
		return fault
	}

	// Unmarshal response content if response is provided
	if response != nil {
		if err := xml.Unmarshal(envelopeResp.Body.Content, response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// faultXML parses a Fault element by local name so any namespace prefix
// (s:, soapenv:, SOAP-ENV:, ...) and either SOAP version matches.
type faultXML struct {
	XMLName xml.Name `xml:"Fault"`

	// SOAP 1.2 shape.
	Code struct {
		Value   string `xml:"Value"`
		Subcode struct {
			Value string `xml:"Value"`
		} `xml:"Subcode"`
	} `xml:"Code"`
	Reason struct {
		Text string `xml:"Text"`
	} `xml:"Reason"`
	Detail struct {
		Inner string `xml:",innerxml"`
	} `xml:"Detail"`

	// SOAP 1.1 shape.
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}

// parseFault detects a SOAP Fault in the raw body content and converts it to a
// FaultError. Returns nil when the content is not a fault: a failed unmarshal
// here simply means the body holds an operation response instead of a Fault
// element. Matching is by local element name, so it tolerates arbitrary
// prefixes and both SOAP 1.1 and 1.2 fault layouts.
func parseFault(bodyContent []byte, httpStatus int) *FaultError {
	var f faultXML
	if err := xml.Unmarshal(bodyContent, &f); err != nil {
		return nil //nolint:nilerr // unmarshal failure = body is not a fault
	}

	fault := &FaultError{
		Code:       f.Code.Value,
		Subcode:    f.Code.Subcode.Value,
		Reason:     f.Reason.Text,
		Detail:     f.Detail.Inner,
		HTTPStatus: httpStatus,
	}

	// Map SOAP 1.1 fields when the 1.2 shape is empty.
	if fault.Code == "" {
		fault.Code = f.FaultCode
	}

	if fault.Reason == "" {
		fault.Reason = f.FaultString
	}

	return fault
}

// createSecurityHeader creates a WS-Security header with a UsernameToken for
// the configured auth mode (digest or cleartext password).
func (c *Client) createSecurityHeader() *Security {
	if c.authMode == AuthModePasswordText {
		return &Security{
			MustUnderstand: "1",
			UsernameToken: &UsernameToken{
				Username: c.username,
				Password: Password{
					Type:     passwordTextType,
					Password: c.password,
				},
			},
		}
	}

	// Generate nonce
	const nonceSize = 16
	nonceBytes := make([]byte, nonceSize)

	_, _ = rand.Read(nonceBytes)
	nonce := base64.StdEncoding.EncodeToString(nonceBytes)

	// Get current timestamp, adjusted by the device's clock skew so the digest's
	// Created field matches the camera's view of "now" (fixes Hikvision time-skew
	// auth rejections). clockSkew is 0 when unset (legacy local-time behavior).
	created := time.Now().UTC().Add(c.clockSkew).Format(time.RFC3339)

	// Calculate password digest: Base64(SHA1(nonce + created + password))
	hash := sha1.New() //nolint:gosec // SHA1 required for ONVIF digest auth
	hash.Write(nonceBytes)
	hash.Write([]byte(created))
	hash.Write([]byte(c.password))
	digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	return &Security{
		MustUnderstand: "1",
		UsernameToken: &UsernameToken{
			Username: c.username,
			Password: Password{
				Type:     passwordDigestType,
				Password: digest,
			},
			Nonce: Nonce{
				Type:  nonceEncodingType,
				Nonce: nonce,
			},
			Created: created,
		},
	}
}

// BuildEnvelope builds a SOAP envelope with the given body content.
func BuildEnvelope(body interface{}, username, password string) (*Envelope, error) {
	envelope := &Envelope{
		Body: Body{
			Content: body,
		},
	}

	if username != "" && password != "" {
		client := &Client{username: username, password: password}
		envelope.Header = &Header{
			Security: client.createSecurityHeader(),
		}
	}

	return envelope, nil
}
