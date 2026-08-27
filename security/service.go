package security

import "github.com/mickeyzzc/onvif-go/v2/internal/api"

// Service covers user management and access control (users, access policy,
// certificates) — kept apart from device because many firmwares gate these
// behind stricter auth.
type Service struct {
	c api.Caller
}

// New creates a security service bound to a caller.
func New(c api.Caller) *Service { return &Service{c: c} }
