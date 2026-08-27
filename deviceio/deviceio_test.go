package deviceio

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

func TestGetDeviceIOServiceCapabilities(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/device", func(action, _ string) (string, error) {
		if action != "tmd:GetServiceCapabilities" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetServiceCapabilitiesResponse>
	<Capabilities VideoSources="2" VideoOutputs="1" AudioSources="1" AudioOutputs="1" RelayOutputs="4"/>
</GetServiceCapabilitiesResponse>`, nil
	})

	caps, err := New(caller).GetDeviceIOServiceCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetDeviceIOServiceCapabilities: %v", err)
	}

	if caps.RelayOutputs != 4 || caps.VideoSources != 2 {
		t.Errorf("capabilities = %+v", caps)
	}
}

func TestGetDigitalInputs(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/device", func(action, _ string) (string, error) {
		if action != "tmd:GetDigitalInputs" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetDigitalInputsResponse>
	<DigitalInputs token="di-1" IdleState="closed"/>
	<DigitalInputs token="di-2" IdleState="open"/>
</GetDigitalInputsResponse>`, nil
	})

	inputs, err := New(caller).GetDigitalInputs(context.Background())
	if err != nil {
		t.Fatalf("GetDigitalInputs: %v", err)
	}

	if len(inputs) != 2 || inputs[0].Token != "di-1" || inputs[1].IdleState != "open" {
		t.Errorf("inputs = %+v", inputs)
	}
}

func TestGetSerialPorts(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/device", func(action, _ string) (string, error) {
		if action != "tmd:GetSerialPorts" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetSerialPortsResponse>
	<SerialPorts token="sp-1"><Type>RS485</Type></SerialPorts>
</GetSerialPortsResponse>`, nil
	})

	ports, err := New(caller).GetSerialPorts(context.Background())
	if err != nil {
		t.Fatalf("GetSerialPorts: %v", err)
	}

	if len(ports) != 1 || ports[0].Token != "sp-1" || ports[0].Type != "RS485" {
		t.Errorf("ports = %+v", ports)
	}
}

func TestSetDigitalInputConfigurations(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/device", func(action, reqXML string) (string, error) {
		if action != "tmd:SetDigitalInputConfigurations" {
			return "", errors.New("unexpected action " + action)
		}

		if !strings.Contains(reqXML, "di-1") {
			t.Errorf("configuration token not encoded: %s", reqXML)
		}

		return `<SetDigitalInputConfigurationsResponse/>`, nil
	})

	err := New(caller).SetDigitalInputConfigurations(context.Background(), []*DigitalInput{
		{Token: "di-1", IdleState: "closed"},
	})
	if err != nil {
		t.Fatalf("SetDigitalInputConfigurations: %v", err)
	}
}

func TestGetVideoOutputs(t *testing.T) {
	caller := testutil.NewFakeCaller("http://fake/device", func(action, _ string) (string, error) {
		if action != "tmd:GetVideoOutputs" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetVideoOutputsResponse>
	<VideoOutputs token="vo-1"><Layout><Name>HDMI</Name></Layout></VideoOutputs>
</GetVideoOutputsResponse>`, nil
	})

	outputs, err := New(caller).GetVideoOutputs(context.Background())
	if err != nil {
		t.Fatalf("GetVideoOutputs: %v", err)
	}

	if len(outputs) != 1 || outputs[0].Token != "vo-1" {
		t.Errorf("outputs = %+v", outputs)
	}
}
