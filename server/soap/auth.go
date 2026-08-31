package soap

import (
	"crypto/sha1" //nolint:gosec // SHA1 required by the ONVIF digest formula
	"crypto/subtle"
	"encoding/base64"
	"strings"

	originsoap "github.com/mickeyzzc/onvif-go/v2/internal/soap"
)

// DefaultProtectedPrefixes lists the action prefixes that require
// authentication under the default policy: write-style operations
// (Set*, Remove*, Create*, Go*) are protected while read operations
// stay open.
//
// It is a function (not a package var) so callers always get a fresh
// slice — appending to a shared package-level var would mutate the
// default policy process-wide.
func DefaultProtectedPrefixes() []string {
	return []string{"Set", "Remove", "Create", "Go"}
}

// AuthPolicy decides which SOAP actions require WS-Security credentials
// and which UsernameToken password types are accepted. It only applies
// when the Handler has credentials configured.
type AuthPolicy struct {
	// Prefixes lists action prefixes requiring authentication. nil →
	// DefaultProtectedPrefixes.
	Prefixes []string

	// Actions lists exact action names additionally requiring
	// authentication (e.g. "SystemReboot").
	Actions []string

	// All requires authentication for every action (strict mode, the
	// pre-v2 behavior).
	All bool

	// AllowPasswordText accepts cleartext WS-Security passwords. Default
	// true; PasswordDigest is always accepted.
	AllowPasswordText bool
}

// DefaultAuthPolicy returns the default policy: write-style actions
// authenticated, reads open, PasswordText accepted.
func DefaultAuthPolicy() *AuthPolicy {
	return &AuthPolicy{
		Prefixes:          DefaultProtectedPrefixes(),
		AllowPasswordText: true,
	}
}

// protectedPrefixes returns the effective prefix list.
func (p *AuthPolicy) protectedPrefixes() []string {
	if p.Prefixes == nil {
		return DefaultProtectedPrefixes()
	}

	return p.Prefixes
}

// require returns true when the action must be authenticated.
func (p *AuthPolicy) require(action string) bool {
	if p.All {
		return true
	}

	for _, a := range p.Actions {
		if a == action {
			return true
		}
	}

	for _, prefix := range p.protectedPrefixes() {
		if strings.HasPrefix(action, prefix) {
			return true
		}
	}

	return false
}

// policy resolves the effective policy, defaulting when none was set.
func (h *Handler) policy() *AuthPolicy {
	if h.auth != nil {
		return h.auth
	}

	return DefaultAuthPolicy()
}

// requiresAuth reports whether the action needs valid credentials. With no
// credentials configured the server stays fully open (unchanged behavior).
func (h *Handler) requiresAuth(action string) bool {
	if h.username == "" && h.password == "" {
		return false
	}

	return h.policy().require(action)
}

// authenticate validates the WS-Security UsernameToken of a request
// header. Both PasswordDigest (the ONVIF default) and PasswordText are
// accepted, per the policy.
func (h *Handler) authenticate(header *originsoap.Header) bool {
	if header == nil || header.Security == nil || header.Security.UsernameToken == nil {
		return false
	}

	token := header.Security.UsernameToken

	if token.Username != h.username {
		return false
	}

	if token.Password.Type == originsoap.PasswordTextType {
		if !h.policy().AllowPasswordText {
			return false
		}

		// Cleartext comparison; constant-time to avoid trivially leaking
		// the expected password length/early-mismatch through timing.
		return constantTimeEqual(token.Password.Password, h.password)
	}

	// Default (and absent Type): PasswordDigest =
	// Base64(SHA1(nonce + created + password)). The created timestamp is
	// read via CreatedValue so clients using the misspelled utility
	// namespace still authenticate (issue #40).
	nonce, err := base64.StdEncoding.DecodeString(token.Nonce.Nonce)
	if err != nil {
		return false
	}

	hash := sha1.New() //nolint:gosec // SHA1 required by the ONVIF digest formula
	hash.Write(nonce)
	hash.Write([]byte(token.CreatedValue()))
	hash.Write([]byte(h.password))

	got, err := base64.StdEncoding.DecodeString(token.Password.Password)
	if err != nil {
		return false
	}

	return constantTimeEqual(string(got), string(hash.Sum(nil)))
}

// constantTimeEqual compares two strings without timing side channels.
func constantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
