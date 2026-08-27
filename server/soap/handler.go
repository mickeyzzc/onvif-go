// Package soap provides SOAP request handling for the ONVIF server.
//
// The Handler is an http.Handler: it parses the SOAP envelope, applies the
// per-action authentication policy, and dispatches the request body to the
// registered ContextHandler. Handlers receive the raw request element bytes
// (decode with ParseRequest or encoding/xml) plus a RequestContext carrying
// the action name, the client IP, and the underlying *http.Request.
package soap

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net"
	"net/http"
	"sync"

	originsoap "github.com/mickeyzzc/onvif-go/v2/internal/soap"
)

// soapEnvelopeNS is the SOAP 1.2 envelope namespace.
const soapEnvelopeNS = "http://www.w3.org/2003/05/soap-envelope"

// requestEnvelope is the decode target for incoming requests. The body is
// captured as raw inner XML — encoding/xml cannot populate interface{}
// fields, so the request element is handed to handlers as bytes.
type requestEnvelope struct {
	XMLName xml.Name           `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Header  *originsoap.Header `xml:"Header"`
	Body    struct {
		Raw []byte `xml:",innerxml"`
	} `xml:"Body"`
}

// Handler handles incoming SOAP requests. Registration is safe while
// the handler is serving traffic (embedding hosts register lazily).
type Handler struct {
	username string
	password string
	auth     *AuthPolicy
	// explicitPrefixes emits responses with s:/tds:/trt:/... namespace
	// prefixes instead of default xmlns declarations.
	explicitPrefixes bool

	mu       sync.RWMutex
	handlers map[string]ContextHandler
}

// RequestContext carries per-request state to message handlers.
type RequestContext struct {
	// Action is the canonical local name of the request element, e.g.
	// "GetStreamUri".
	Action string

	// RemoteIP is the client's IP address (host part of RemoteAddr).
	// Real cameras echo it as the host of advertised URLs so each peer
	// receives addresses reachable from its own network.
	RemoteIP string

	// Request is the underlying HTTP request.
	Request *http.Request
}

// Context returns the request-scoped context for cancellation, deadlines,
// and values. It never returns nil.
func (c *RequestContext) Context() context.Context {
	if c == nil || c.Request == nil {
		return context.Background()
	}

	return c.Request.Context()
}

// ContextHandler handles a SOAP message with access to request state.
type ContextHandler func(ctx *RequestContext, body []byte) (interface{}, error)

// MessageHandler is the legacy handler signature: no request context, body
// as raw bytes. RegisterHandler wraps it into a ContextHandler.
type MessageHandler func(body []byte) (interface{}, error)

// HandlerOptions configures a Handler.
type HandlerOptions struct {
	// Username and Password enable WS-Security UsernameToken validation.
	// When either is empty, every action is served without authentication.
	Username string
	Password string

	// Auth is the per-action authentication policy applied when credentials
	// are configured. nil → DefaultAuthPolicy (write-style actions require
	// authentication, reads stay open, PasswordText accepted).
	Auth *AuthPolicy

	// ExplicitPrefixes emits response envelopes with explicit namespace
	// prefixes (s:Envelope, trt:GetStreamUriResponse, ...) instead of
	// default xmlns declarations. RawXML responses are never rewritten.
	ExplicitPrefixes bool
}

// NewHandler creates a new SOAP handler with the default per-action
// authentication policy.
func NewHandler(username, password string) *Handler {
	return NewHandlerWithOptions(HandlerOptions{
		Username: username,
		Password: password,
	})
}

// NewHandlerWithOptions creates a new SOAP handler.
func NewHandlerWithOptions(opts HandlerOptions) *Handler {
	return &Handler{
		username:         opts.Username,
		password:         opts.Password,
		auth:             opts.Auth,
		explicitPrefixes: opts.ExplicitPrefixes,
		handlers:         make(map[string]ContextHandler),
	}
}

// RegisterHandler registers a handler for a specific action/message type.
// The action name should use the ONVIF WSDL spelling (e.g. "GetStreamUri").
func (h *Handler) RegisterHandler(action string, handler MessageHandler) {
	h.RegisterContextHandler(action, func(_ *RequestContext, body []byte) (interface{}, error) {
		return handler(body)
	})
}

// RegisterContextHandler registers a context-aware handler for a specific
// action/message type.
func (h *Handler) RegisterContextHandler(action string, handler ContextHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.handlers[action] = handler
}

// ServeHTTP implements http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.sendFault(w, "Receiver", "Failed to read request body", err.Error())

		return
	}
	_ = r.Body.Close()

	// Extract action from raw XML first (before parsing), canonicalized
	// to the ONVIF WSDL spelling so legacy client spellings still
	// dispatch to the registered handler.
	action := canonicalAction(h.extractAction(body))
	if action == "" {
		h.sendFault(w, "Sender", "Unknown action", "Could not determine request action")

		return
	}

	// Parse SOAP envelope
	var envelope requestEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		h.sendFault(w, "Sender", "Invalid SOAP envelope", err.Error())

		return
	}

	// Authenticate actions the policy protects
	if h.requiresAuth(action) && !h.authenticate(envelope.Header) {
		h.sendFault(w, "Sender", "Sender not authorized", "Invalid username or password")

		return
	}

	// Find and execute handler
	h.mu.RLock()
	handler, ok := h.handlers[action]
	h.mu.RUnlock()

	if !ok {
		h.sendFault(w, "Receiver", "Action not supported", "No handler for action: "+action)

		return
	}

	reqCtx := &RequestContext{
		Action:   action,
		RemoteIP: remoteIP(r),
		Request:  r,
	}

	// Execute handler
	response, err := handler(reqCtx, envelope.Body.Raw)
	if err != nil {
		h.sendFault(w, "Receiver", "Handler error", err.Error())

		return
	}

	// Send response
	h.sendResponse(w, response)
}

// extractAction extracts the action/message type from the SOAP body.
func (h *Handler) extractAction(bodyXML []byte) string {
	// Parse XML to find the first element inside the Body element
	decoder := xml.NewDecoder(bytes.NewReader(bodyXML))
	inBody := false
	depth := 0

	for {
		token, err := decoder.Token()
		if err != nil {
			return ""
		}

		switch t := token.(type) {
		case xml.StartElement:
			depth++
			// Check if we're entering the Body element
			if t.Name.Local == "Body" {
				inBody = true
			} else if inBody && depth > 2 {
				// Found the first element inside Body
				return t.Name.Local
			}
		case xml.EndElement:
			depth--
			if t.Name.Local == "Body" {
				inBody = false
			}
		}
	}
}

// canonicalActions maps legacy action spellings to the ONVIF WSDL names.
var canonicalActions = map[string]string{
	"GetStreamURI":   "GetStreamUri",
	"GetSnapshotURI": "GetSnapshotUri",
}

// canonicalAction returns the canonical WSDL spelling for an action name,
// or the name itself when no mapping exists.
func canonicalAction(action string) string {
	if canonical, ok := canonicalActions[action]; ok {
		return canonical
	}

	return action
}

// remoteIP extracts the client IP from an HTTP request.
func remoteIP(r *http.Request) string {
	if r == nil || r.RemoteAddr == "" {
		return ""
	}

	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}
