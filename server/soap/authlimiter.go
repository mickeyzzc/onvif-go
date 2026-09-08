package soap

import (
	"sync"
	"time"
)

// authFailureTracker counts authentication failures per source IP and
// locks repeat offenders out for a window — the brute-force backstop of
// issue #60. Safe for concurrent use.
type authFailureTracker struct {
	mu      sync.Mutex
	entries map[string]*authFailEntry

	max     int
	lockout time.Duration
	now     func() time.Time
}

type authFailEntry struct {
	fails       int
	lockedUntil time.Time
}

func newAuthFailureTracker(max int, lockout time.Duration) *authFailureTracker {
	return &authFailureTracker{
		entries: make(map[string]*authFailEntry),
		max:     max,
		lockout: lockout,
		now:     time.Now,
	}
}

// locked reports whether the source is currently locked out; served
// lockouts clear lazily.
func (t *authFailureTracker) locked(key string) bool {
	if t.max <= 0 {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[key]
	if !ok {
		return false
	}
	now := t.now()
	if now.Before(e.lockedUntil) {
		return true
	}
	if !e.lockedUntil.IsZero() {
		delete(t.entries, key)
	}
	return false
}

// recordFailure adds one failure; at max failures the source locks out.
func (t *authFailureTracker) recordFailure(key string) {
	if t.max <= 0 {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[key]
	if !ok {
		e = &authFailEntry{}
		t.entries[key] = e
	}
	e.fails++
	if e.fails >= t.max {
		e.lockedUntil = t.now().Add(t.lockout)
	}
}

// recordSuccess clears the source's failure budget.
func (t *authFailureTracker) recordSuccess(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, key)
}
