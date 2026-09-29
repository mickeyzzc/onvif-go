package simulator

import (
	"errors"
	"sync"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
)

// ioSimulator is the simulator's alarm I/O surface: two bistable relays
// and one digital input — enough for NVR alarm-linkage integration
// (arm relay → verify → release, and wire alarm inputs).
type ioSimulator struct {
	mu     sync.Mutex
	relays []provider.RelayOutput
	inputs []provider.DigitalInput
}

func newIOSimulator() *ioSimulator {
	return &ioSimulator{
		relays: []provider.RelayOutput{
			{Token: "relay_1", Mode: "Bistable", DelayTime: "PT0S", IdleState: "open", LogicalState: "inactive"},
			{Token: "relay_2", Mode: "Bistable", DelayTime: "PT0S", IdleState: "open", LogicalState: "inactive"},
		},
		inputs: []provider.DigitalInput{
			{Token: "di_1", IdleState: "open"},
		},
	}
}

// RelayOutputs returns a copy of the relay list with live states.
func (s *ioSimulator) RelayOutputs() []provider.RelayOutput {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]provider.RelayOutput(nil), s.relays...)
}

// SetRelayState flips a relay's logical state (active/inactive).
func (s *ioSimulator) SetRelayState(token, logicalState string) error {
	if logicalState != "active" && logicalState != "inactive" {
		return errors.New("logical state must be active or inactive") //nolint:goerr113 // simulator input validation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.relays {
		if s.relays[i].Token == token {
			s.relays[i].LogicalState = logicalState

			return nil
		}
	}

	return provider.ErrNotFound
}

// DigitalInputs returns the input lines.
func (s *ioSimulator) DigitalInputs() []provider.DigitalInput {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]provider.DigitalInput(nil), s.inputs...)
}
