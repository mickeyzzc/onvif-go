// Package deviceio hosts the device-IO-service (tmd) domain types.
package deviceio

import (
	"time"
)

// RelayOutput represents relay output.
type RelayOutput struct {
	Token      string
	Properties RelayOutputSettings
}

// RelayOutputSettings represents relay output settings.
type RelayOutputSettings struct {
	Mode      RelayMode
	DelayTime time.Duration
	IdleState RelayIdleState
}

// RelayMode represents relay mode.
type RelayMode string

// RelayIdleState represents relay idle state.
type RelayIdleState string

// RelayLogicalState represents relay logical state.
type RelayLogicalState string

const (
	RelayModeMonostable RelayMode = "Monostable"
	RelayModeBistable   RelayMode = "Bistable"
)

const (
	RelayIdleStateClosed RelayIdleState = "closed"
	RelayIdleStateOpen   RelayIdleState = "open"
)

const (
	RelayLogicalStateActive   RelayLogicalState = "active"
	RelayLogicalStateInactive RelayLogicalState = "inactive"
)
