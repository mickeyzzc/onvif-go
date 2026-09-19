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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	originsoap "github.com/mickeyzzc/onvif-go/v2/internal/soap"
	"github.com/mickeyzzc/onvif-go/v2/metrics"
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
	// anonymousAllowed records the explicit AllowAnonymous opt-in (see
	// HandlerOptions).
	anonymousAllowed bool
	// maxBodyBytes bounds request bodies (0 = unlimited).
	maxBodyBytes int64
	// authFailures is the per-source brute-force backstop (issue #60).
	authFailures *authFailureTracker
	// metrics receives observability events (issue #66); never nil.
	metrics metrics.Hooks

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

	// MaxBodyBytes bounds the accepted request body; larger bodies are
	// rejected with 413. 0 = default 1 MiB, negative = unlimited (tests
	// only). Issue #59.
	MaxBodyBytes int

	// AuthFailureLimit is how many authentication failures from one source
	// trigger a lockout; 0 = default 5, negative disables. Issue #60.
	AuthFailureLimit int

	// AuthLockout is how long a locked-out source is refused.
	// 0 = default 60s.
	AuthLockout time.Duration

	// Metrics receives observability events: dispatched requests, handler
	// faults, auth failures, lockout refusals (issue #66). nil = no-ops.
	Metrics metrics.Hooks

	// AllowAnonymous documents the empty-credentials open mode explicitly.
	// Today empty Username/Password serves every action without
	// authentication regardless of this flag (legacy behavior, with a
	// startup warning); a future major version will require this opt-in.
	AllowAnonymous bool
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
	limit := opts.MaxBodyBytes
	if limit == 0 {
		limit = 1 << 20
	} else if limit < 0 {
		limit = 0 // unlimited
	}
	failureLimit := opts.AuthFailureLimit
	if failureLimit == 0 {
		failureLimit = 5
	} else if failureLimit < 0 {
		failureLimit = 0
	}
	lockout := opts.AuthLockout
	if lockout == 0 {
		lockout = 60 * time.Second
	}
	if opts.Username == "" || opts.Password == "" {
		slog.Warn("onvif: SOAP handler serving WITHOUT authentication (empty credentials) — set credentials or see AllowAnonymous")
	}
	return &Handler{
		username:         opts.Username,
		password:         opts.Password,
		auth:             opts.Auth,
		explicitPrefixes: opts.ExplicitPrefixes,
		handlers:         make(map[string]ContextHandler),
		maxBodyBytes:     int64(limit),
		anonymousAllowed: opts.AllowAnonymous,
		authFailures:     newAuthFailureTracker(failureLimit, lockout),
		metrics:          metrics.OrNoop(opts.Metrics),
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

	// Read request body, bounded (issue #59).
	var body []byte
	var err error
	if h.maxBodyBytes > 0 {
		body, err = io.ReadAll(http.MaxBytesReader(w, r.Body, h.maxBodyBytes))
	} else {
		body, err = io.ReadAll(r.Body)
	}
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
		} else {
			h.sendFault(w, "Receiver", "Failed to read request body", err.Error())
		}

		return
	}
	_ = r.Body.Close()

	// Brute-force backstop: refuse locked-out sources before any auth
	// work (issue #60).
	source := remoteIP(r)
	if h.authFailures.locked(source) {
		h.metrics.AuthLockout()
		h.sendFault(w, "Sender", "Too many authentication failures", "source temporarily locked out")

		return
	}

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

	// Extract the request element with its namespace context intact:
	// bindings declared on Envelope/Header ancestors are materialized on
	// the extracted root, so prefix-style requests decode even when the
	// client declared every prefix at the envelope level.
	bodyContent, err := extractBodyElement(body)
	if err != nil {
		h.sendFault(w, "Sender", "Invalid SOAP body", err.Error())

		return
	}

	// Authenticate actions the policy protects
	if h.requiresAuth(action) {
		if !h.authenticate(envelope.Header) {
			h.authFailures.recordFailure(source)
			h.metrics.AuthFail()
			h.sendFault(w, "Sender", "Sender not authorized", "Invalid username or password")

			return
		}
		h.authFailures.recordSuccess(source)
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
	h.metrics.SoapRequest(action)
	response, err := handler(reqCtx, bodyContent)
	if err != nil {
		h.metrics.SoapFault(action)

		var sender *SenderFaultError
		if errors.As(err, &sender) {
			h.sendFault(w, "Sender", sender.Reason, sender.Detail)

			return
		}

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

// extractBodyElement re-encodes the first child of the SOAP Body as a
// self-contained fragment: element names carry no prefix and every
// namespace is declared as a default xmlns on the element that needs it.
// The decoder resolves prefixes to URIs but EncodeToken cannot bind them
// back (it treats Attr.Name.Space as a literal prefix), so canonical
// default-xmlns form is the only token-level encoding that survives the
// round trip. Handlers can decode the fragment with namespace-strict tags
// regardless of where the client declared its prefixes — raw innerxml
// extraction instead drops ancestor declarations and leaves unbound
// prefixes for envelope-level declarers.
func extractBodyElement(bodyXML []byte) ([]byte, error) {
	decoder := xml.NewDecoder(bytes.NewReader(bodyXML))

	var (
		stack   []xml.Name // open elements; the top is the Body parent
		capture []xml.Token
		depth   int      // capture subtree depth (0 = not capturing)
		nsStack []string // effective default namespace per captured level
		ns      string   // effective default namespace while capturing
	)

	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("parse body: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			isBodyChild := depth == 0 && isBodyChildElement(stack)

			stack = append(stack, t.Name)

			if isBodyChild {
				capture = append(capture, canonicalizeElement(t, ns))
				ns = t.Name.Space
				nsStack = append(nsStack, ns)
				depth = 1

				continue
			}

			if depth > 0 {
				capture = append(capture, canonicalizeElement(t, ns))
				ns = t.Name.Space
				nsStack = append(nsStack, ns)
				depth++
			}
		case xml.EndElement:
			if depth > 0 {
				t.Name = xml.Name{Local: t.Name.Local}
				capture = append(capture, t)
				depth--

				nsStack = nsStack[:len(nsStack)-1]
				if len(nsStack) > 0 {
					ns = nsStack[len(nsStack)-1]
				}

				if depth == 0 {
					return flushCaptured(capture)
				}
			}

			stack = stack[:len(stack)-1]
		default:
			if depth > 0 {
				capture = append(capture, xml.CopyToken(token))
			}
		}
	}

	if len(capture) > 0 {
		return nil, errors.New("unbalanced SOAP body element")
	}

	return nil, nil
}

// isBodyChildElement reports whether the element about to be pushed is
// the first child of the SOAP Body: the current top of the stack is Body
// in the SOAP envelope namespace (the decoder resolves prefixes to URIs,
// so this is prefix-independent).
func isBodyChildElement(stack []xml.Name) bool {
	if len(stack) == 0 {
		return false
	}

	top := stack[len(stack)-1]

	return top.Local == "Body" && top.Space == soapEnvelopeNS
}

// canonicalizeElement rewrites one captured element into default-xmlns
// form: prefix-free name, non-xmlns attributes spelled as local names,
// and an xmlns declaration whenever the element's namespace differs from
// the in-scope default. Binding declarations from the source document are
// dropped and re-derived — they are what this canonicalization replaces.
func canonicalizeElement(t xml.StartElement, defaultNS string) xml.StartElement {
	attrs := make([]xml.Attr, 0, len(t.Attr)+1)

	for _, attr := range t.Attr {
		if attr.Name.Space == "xmlns" || (attr.Name.Space == "" && attr.Name.Local == "xmlns") {
			continue
		}

		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: attr.Name.Local}, Value: attr.Value})
	}

	if t.Name.Space != defaultNS {
		attrs = append(attrs, xmlnsAttr("", t.Name.Space))
	}

	return xml.StartElement{Name: xml.Name{Local: t.Name.Local}, Attr: attrs}
}

// xmlnsAttr builds a default xmlns declaration attribute (empty URI means
// no namespace).
func xmlnsAttr(_, uri string) xml.Attr {
	return xml.Attr{Name: xml.Name{Local: "xmlns"}, Value: uri}
}

// flushCaptured encodes the captured token slice back to bytes.
func flushCaptured(capture []xml.Token) ([]byte, error) {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)

	for _, tok := range capture {
		if err := enc.EncodeToken(tok); err != nil {
			return nil, fmt.Errorf("re-encode body element: %w", err)
		}
	}

	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("flush body element: %w", err)
	}

	return buf.Bytes(), nil
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
