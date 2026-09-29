package server

import (
	"errors"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/server/provider"
	"github.com/mickeyzzc/onvif-go/v2/server/soap"
)

// noHardwareRelays is the hardware-honest RelayController: a device with
// no relay outputs and no digital inputs (real hosts inject this instead
// of the simulator's fabricated alarm I/O).
type noHardwareRelays struct{}

func (noHardwareRelays) RelayOutputs() []provider.RelayOutput { return nil }

func (noHardwareRelays) DigitalInputs() []provider.DigitalInput { return nil }

func (noHardwareRelays) SetRelayState(_, _ string) error { return provider.ErrNotFound }

// TestWithRelayControllerHonestEmptySets locks the host seam: every
// DeviceIO action stays answerable while reporting exactly the
// injected (empty) hardware set.
func TestWithRelayControllerHonestEmptySets(t *testing.T) {
	config := createTestConfig()
	config.SupportDeviceIO = true

	server, err := New(config, WithRelayController(noHardwareRelays{}))
	if err != nil {
		t.Fatalf("New with relay controller: %v", err)
	}

	resp, err := server.HandleGetRelayOutputs(nil, nil)
	if err != nil {
		t.Fatalf("HandleGetRelayOutputs: %v", err)
	}
	if got := len(resp.(getRelayOutputsResponse).RelayOutputs); got != 0 {
		t.Errorf("RelayOutputs = %d entries, want the injected empty set", got)
	}

	diResp, err := server.HandleGetDigitalInputs(nil, nil)
	if err != nil {
		t.Fatalf("HandleGetDigitalInputs: %v", err)
	}
	if got := len(diResp.(getDigitalInputsResponse).DigitalInputs); got != 0 {
		t.Errorf("DigitalInputs = %d entries, want the injected empty set", got)
	}

	capsResp, err := server.HandleGetDeviceIOServiceCapabilities(nil, nil)
	if err != nil {
		t.Fatalf("HandleGetDeviceIOServiceCapabilities: %v", err)
	}
	caps := capsResp.(deviceIOCapabilitiesWire)
	if caps.Capabilities.RelayOutputs != 0 || caps.Capabilities.DigitalInputs != 0 {
		t.Errorf("capabilities counts = %+v, want all zero", caps.Capabilities)
	}

	// Mutating a relay that does not exist must fault, not pretend.
	body := []byte(`<SetRelayOutputState xmlns="http://www.onvif.org/ver10/deviceIO/wsdl">` +
		`<RelayOutputToken>relay_1</RelayOutputToken><LogicalState>active</LogicalState></SetRelayOutputState>`)
	_, err = server.HandleSetRelayOutputState(nil, body)
	var sender *soap.SenderFaultError
	if !errors.As(err, &sender) {
		t.Fatalf("HandleSetRelayOutputState on empty hardware: err = %v, want SenderFault", err)
	}
}

// TestWithRelayControllerNilKeepsSimulator locks the default: without the
// option the simulator's alarm I/O set keeps backing the family (the
// virtual-camera behavior of #119).
func TestWithRelayControllerNilKeepsSimulator(t *testing.T) {
	server, err := New(createTestConfig(), WithRelayController(nil))
	if err != nil {
		t.Fatalf("New with nil relay controller: %v", err)
	}

	resp, err := server.HandleGetRelayOutputs(nil, nil)
	if err != nil {
		t.Fatalf("HandleGetRelayOutputs: %v", err)
	}
	outputs := resp.(getRelayOutputsResponse).RelayOutputs
	if len(outputs) == 0 || outputs[0].Token != "relay_1" {
		t.Errorf("default relays = %+v, want the simulator set (relay_1 first)", outputs)
	}
}
