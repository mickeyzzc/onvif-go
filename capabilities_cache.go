package onvif

import (
	"context"
	"fmt"
)

// WithMinimalCapsFallback makes the cached capabilities accessor degrade
// gracefully: when GetCapabilities fails (minimal embedded devices can fault
// on it), a minimal all-off capability set is returned AND cached instead of
// an error. Callers gate advanced calls on it (no PTZ capability → no PTZ
// calls), and a weak device is never hammered with retries. Without this
// option the conservative default applies: failures surface as errors and
// are not cached.
func WithMinimalCapsFallback() ClientOption {
	return func(c *Client) {
		c.minimalCapsFallback = true
	}
}

// InvalidateCapabilitiesCache drops the cached capabilities so the next
// GetCapabilitiesCached call re-fetches them. Call after a firmware upgrade
// or any change that could alter the device's advertised services.
func (c *Client) InvalidateCapabilitiesCache() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.capsCache = nil
	c.capsCached = false
}

// GetCapabilitiesCached returns the device capabilities, fetching them at
// most once: the response is essentially constant for a device's runtime.
// This matters for weak devices (ESP32-class) where every extra handshake
// counts. The returned pointer is shared — treat it as read-only.
//
// With WithMinimalCapsFallback configured, a device that faults on
// GetCapabilities yields a minimal (all-off) capability set, cached like a
// successful fetch; nil sub-capabilities mean "not available".
//
// Concurrent first callers are single-flighted: only one SOAP request goes
// out and the rest wait for its result.
func (s *DeviceService) GetCapabilitiesCached(ctx context.Context) (*Capabilities, error) {
	c := s.client

	for {
		c.mu.RLock()
		if c.capsCached {
			caps := c.capsCache
			c.mu.RUnlock()

			return caps, nil
		}

		fetching := c.capsFetching
		var ready chan struct{}
		if fetching {
			ready = c.capsReady
		}
		c.mu.RUnlock()

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

		c.mu.Lock()
		// Double-check under the write lock.
		if c.capsCached {
			caps := c.capsCache
			c.mu.Unlock()

			return caps, nil
		}

		if c.capsFetching {
			c.mu.Unlock()

			continue
		}

		c.capsFetching = true
		c.capsReady = make(chan struct{})
		c.mu.Unlock()

		fetchedCaps, err := s.GetCapabilities(ctx)

		c.mu.Lock()
		done := c.capsReady
		if err == nil {
			c.capsCache, c.capsCached = fetchedCaps, true
		} else if c.minimalCapsFallback {
			// Degrade to a minimal capability set and cache it: the caller
			// can gate advanced features on it without re-hammering the
			// device on every call.
			c.capsCache, c.capsCached = &Capabilities{}, true
		}
		c.capsFetching = false
		c.capsReady = nil
		result, cached := c.capsCache, c.capsCached
		c.mu.Unlock()

		// Wake the waiters only after the cache state is settled.
		close(done)

		if err != nil && !cached {
			return nil, fmt.Errorf("GetCapabilitiesCached: %w", err)
		}

		return result, nil
	}
}
