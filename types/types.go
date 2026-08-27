// Package types hosts the shared cross-service data primitives and sentinels.
package types

import (
	"errors"
)

// IPAddress represents an IP address.
type IPAddress struct {
	Type        string // IPv4 or IPv6
	Address     string
	IPv4Address string
	IPv6Address string
}

// IntRectangle represents a rectangle with integer coordinates.
type IntRectangle struct {
	X      int
	Y      int
	Width  int
	Height int
}

// FloatRange represents a float range.
type FloatRange struct {
	Min float64
	Max float64
}

// SimpleItem represents a simple configuration item.
type SimpleItem struct {
	Name  string
	Value string
}

// ElementItem represents an element configuration item.
type ElementItem struct {
	Name string
}

// IntRange represents integer range.
type IntRange struct {
	Min int
	Max int
}

// ErrInvalidParameter is the shared sentinel for invalid arguments,
// re-exported by the root package.
var ErrInvalidParameter = errors.New("invalid parameter")

// for 24) so consumers do not each reimplement the conversion.
type PrefixedIPv4Address struct {
	Address      string
	PrefixLength int
	Netmask      string
}

// PrefixedIPv6Address represents an IPv6 address with prefix.
type PrefixedIPv6Address struct {
	Address      string
	PrefixLength int
}

// ErrServiceNotSupported is returned when a service is not supported by the
// device, re-exported by the root package.
var ErrServiceNotSupported = errors.New("service not supported")
