package device

import (
	"context"
	"fmt"
)

// InvalidateCapsCache drops the cached capabilities so the next
// GetCapabilitiesCached call re-fetches them. Call after a firmware upgrade
// or any change that could alter the device's advertised services.
func (s *Service) InvalidateCapsCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.capsCache = nil
	s.capsCached = false
}

// GetCapabilitiesCached returns the device capabilities, fetching them at
// most once: the response is essentially constant for a device's runtime.
// This matters for weak devices (ESP32-class) where every extra handshake
// counts. The returned pointer is shared — treat it as read-only.
//
// With the minimal-caps fallback enabled (the root package's
// WithMinimalCapsFallback), a device that faults on GetCapabilities yields a
// minimal (all-off) capability set, cached like a successful fetch; nil
// sub-capabilities mean "not available".
//
// Concurrent first callers are single-flighted: only one SOAP request goes
// out and the rest wait for its result.
func (s *Service) GetCapabilitiesCached(ctx context.Context) (*Capabilities, error) {
	for {
		s.mu.RLock()
		if s.capsCached {
			caps := s.capsCache
			s.mu.RUnlock()

			return caps, nil
		}

		fetching := s.capsFetching
		var ready chan struct{}
		if fetching {
			ready = s.capsReady
		}
		s.mu.RUnlock()

		if fetching {
			// Another goroutine is fetching — wait for it instead of
			// piling onto a weak device.
			select {
			case <-ready:
			case <-ctx.Done():
				return nil, fmt.Errorf("GetCapabilitiesCached: %w", ctx.Err())
			}

			continue
		}

		s.mu.Lock()
		// Double-check under the write lock.
		if s.capsCached {
			caps := s.capsCache
			s.mu.Unlock()

			return caps, nil
		}

		if s.capsFetching {
			s.mu.Unlock()

			continue
		}

		s.capsFetching = true
		s.capsReady = make(chan struct{})
		s.mu.Unlock()

		fetchedCaps, err := s.GetCapabilities(ctx)

		s.mu.Lock()
		done := s.capsReady
		if err == nil {
			s.capsCache, s.capsCached = fetchedCaps, true
		} else if s.minimalCapsFallback {
			// Degrade to a minimal capability set and cache it: the caller
			// can gate advanced features on it without re-hammering the
			// device on every call.
			s.capsCache, s.capsCached = &Capabilities{}, true
		}
		s.capsFetching = false
		s.capsReady = nil
		result, cached := s.capsCache, s.capsCached
		s.mu.Unlock()

		// Wake the waiters only after the cache state is settled.
		close(done)

		if err != nil && !cached {
			return nil, fmt.Errorf("GetCapabilitiesCached: %w", err)
		}

		return result, nil
	}
}
