package ptz

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

// Issue #111: GetConfigurationOptions / SetConfiguration /
// SendAuxiliaryCommand / GetNodes / GetPTZServiceCapabilities.

const configurationOptionsResponse = `<GetConfigurationOptionsResponse>
	<PTZConfigurationOptions>
		<Spaces>
			<AbsolutePanTiltPositionSpace>
				<URI>http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace</URI>
				<XRange><Min>-1</Min><Max>1</Max></XRange>
				<YRange><Min>-1</Min><Max>1</Max></YRange>
			</AbsolutePanTiltPositionSpace>
			<AbsoluteZoomPositionSpace>
				<URI>http://www.onvif.org/ver10/tptz/ZoomSpaces/PositionGenericSpace</URI>
				<XRange><Min>0</Min><Max>1</Max></XRange>
			</AbsoluteZoomPositionSpace>
			<ContinuousPanTiltVelocitySpace>
				<URI>http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace</URI>
				<XRange><Min>-1</Min><Max>1</Max></XRange>
				<YRange><Min>-1</Min><Max>1</Max></YRange>
			</ContinuousPanTiltVelocitySpace>
			<ContinuousZoomVelocitySpace>
				<URI>http://www.onvif.org/ver10/tptz/ZoomSpaces/VelocityGenericSpace</URI>
				<XRange><Min>-1</Min><Max>1</Max></XRange>
			</ContinuousZoomVelocitySpace>
		</Spaces>
		<PTZTimeout>PT5S</PTZTimeout>
	</PTZConfigurationOptions>
</GetConfigurationOptionsResponse>`

const nodesResponse = `<GetNodesResponse>
	<PTZNode token="ptz-node-1" FixedHomePosition="true">
		<Name>Node 1</Name>
		<SupportedPTZSpaces>
			<AbsolutePanTiltPositionSpace>
				<URI>http://www.onvif.org/ver10/tptz/PanTiltSpaces/PositionGenericSpace</URI>
				<XRange><Min>-1</Min><Max>1</Max></XRange>
				<YRange><Min>-1</Min><Max>1</Max></YRange>
			</AbsolutePanTiltPositionSpace>
		</SupportedPTZSpaces>
	</PTZNode>
</GetNodesResponse>`

const ptzCapabilitiesResponse = `<GetServiceCapabilitiesResponse>
	<Capabilities EFlip="false" Reverse="false" GetCompatibleConfigurations="true"/>
</GetServiceCapabilitiesResponse>`

const auxiliaryResponse = `<AuxiliaryCommandResponse>
	<AuxiliaryData>wiper:active</AuxiliaryData>
</AuxiliaryCommandResponse>`

func TestGetConfigurationOptionsParsesSpaces(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/ptz", func(action, _ string) (string, error) {
		if action != "tptz:GetConfigurationOptions" {
			return "", errors.New("unexpected action " + action)
		}

		return configurationOptionsResponse, nil
	}))

	opts, err := svc.GetConfigurationOptions(context.Background(), "ptz-conf-1")
	if err != nil {
		t.Fatalf("GetConfigurationOptions: %v", err)
	}

	if len(opts.Spaces.AbsolutePanTiltPositionSpace) != 1 {
		t.Fatalf("absolute pan/tilt spaces = %d, want 1", len(opts.Spaces.AbsolutePanTiltPositionSpace))
	}

	abs := opts.Spaces.AbsolutePanTiltPositionSpace[0]
	if abs.XRange == nil || abs.XRange.Min != -1 || abs.XRange.Max != 1 {
		t.Errorf("absolute pan/tilt XRange = %+v, want [-1,1]", abs.XRange)
	}

	if len(opts.Spaces.ContinuousZoomVelocitySpace) != 1 || opts.Spaces.ContinuousZoomVelocitySpace[0].URI == "" {
		t.Error("continuous zoom velocity space missing")
	}

	if opts.PTZTimeout != 5*time.Second {
		t.Errorf("PTZTimeout = %v, want 5s", opts.PTZTimeout)
	}
}

func TestSetConfigurationEncodesFullConfiguration(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/ptz", func(action, reqXML string) (string, error) {
		if action != "tptz:SetConfiguration" {
			return "", errors.New("unexpected action " + action)
		}

		for _, want := range []string{
			`<tptz:PTZConfiguration token="ptz-conf-1">`,
			"<tt:Name>Main PTZ</tt:Name>",
			"<tt:NodeToken>ptz-node-1</tt:NodeToken>",
			"<tt:DefaultPTZSpeed>",
			"<tptz:ForcePersistence>true</tptz:ForcePersistence>",
		} {
			if !strings.Contains(reqXML, want) {
				t.Errorf("request missing %q:\n%s", want, reqXML)
			}
		}

		return "<SetConfigurationResponse/>", nil
	}))

	err := svc.SetConfiguration(context.Background(), &PTZConfiguration{
		Token:     "ptz-conf-1",
		Name:      "Main PTZ",
		NodeToken: "ptz-node-1",
		DefaultPTZSpeed: &PTZSpeed{
			PanTilt: &Vector2D{X: 0.5, Y: 0.5},
		},
	}, true)
	if err != nil {
		t.Fatalf("SetConfiguration: %v", err)
	}
}

func TestSendAuxiliaryCommandRoundTrips(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/ptz", func(action, reqXML string) (string, error) {
		if action != "tptz:SendAuxiliaryCommand" {
			return "", errors.New("unexpected action " + action)
		}

		if !strings.Contains(reqXML, "<tt:AuxiliaryData>wiper:on</tt:AuxiliaryData>") {
			t.Errorf("request missing auxiliary data:\n%s", reqXML)
		}

		return auxiliaryResponse, nil
	}))

	data, err := svc.SendAuxiliaryCommand(context.Background(), "profile-1", "wiper:on")
	if err != nil {
		t.Fatalf("SendAuxiliaryCommand: %v", err)
	}

	if data != "wiper:active" {
		t.Errorf("auxiliary response data = %q, want %q", data, "wiper:active")
	}
}

func TestGetNodesParsesNodeList(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/ptz", func(action, _ string) (string, error) {
		if action != "tptz:GetNodes" {
			return "", errors.New("unexpected action " + action)
		}

		return nodesResponse, nil
	}))

	nodes, err := svc.GetNodes(context.Background())
	if err != nil {
		t.Fatalf("GetNodes: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(nodes))
	}

	node := nodes[0]
	if node.Token != "ptz-node-1" || node.Name != "Node 1" {
		t.Errorf("node = %+v, want token ptz-node-1 / name Node 1", node)
	}

	if !node.FixedHomePosition {
		t.Error("FixedHomePosition = false, want true")
	}

	if node.SupportedPTZSpaces == nil || len(node.SupportedPTZSpaces.AbsolutePanTiltPositionSpace) != 1 {
		t.Errorf("supported spaces not parsed: %+v", node.SupportedPTZSpaces)
	}
}

func TestGetServiceCapabilitiesParsesAttributes(t *testing.T) {
	svc := New(testutil.NewFakeCaller("http://fake/ptz", func(action, _ string) (string, error) {
		if action != "tptz:GetServiceCapabilities" {
			return "", errors.New("unexpected action " + action)
		}

		return ptzCapabilitiesResponse, nil
	}))

	caps, err := svc.GetServiceCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetServiceCapabilities: %v", err)
	}

	if !caps.GetCompatibleConfigurations {
		t.Error("GetCompatibleConfigurations = false, want true")
	}

	if caps.EFlip || caps.Reverse {
		t.Errorf("EFlip/Reverse = %v/%v, want false/false", caps.EFlip, caps.Reverse)
	}
}

func TestCompletionOpsNotSupportedWithoutEndpoint(t *testing.T) {
	svc := New(testutil.NewFakeCaller("", func(string, string) (string, error) { return "", nil }))

	if _, err := svc.GetConfigurationOptions(context.Background(), "c"); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("GetConfigurationOptions error = %v, want ErrServiceNotSupported", err)
	}

	if err := svc.SetConfiguration(context.Background(), &PTZConfiguration{}, false); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("SetConfiguration error = %v, want ErrServiceNotSupported", err)
	}

	if _, err := svc.SendAuxiliaryCommand(context.Background(), "p", "x"); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("SendAuxiliaryCommand error = %v, want ErrServiceNotSupported", err)
	}

	if _, err := svc.GetNodes(context.Background()); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("GetNodes error = %v, want ErrServiceNotSupported", err)
	}

	if _, err := svc.GetServiceCapabilities(context.Background()); !errors.Is(err, types.ErrServiceNotSupported) {
		t.Errorf("GetServiceCapabilities error = %v, want ErrServiceNotSupported", err)
	}
}
