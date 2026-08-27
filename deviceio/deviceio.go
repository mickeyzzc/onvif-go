package deviceio

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/mickeyzzc/onvif-go/v2/internal/api"
	"github.com/mickeyzzc/onvif-go/v2/media"
)

// Namespace is the device IO service WSDL namespace (tmd).
const Namespace = "http://www.onvif.org/ver10/deviceIO/wsdl"

// Device IO service namespace.
const deviceIONamespace = "http://www.onvif.org/ver10/deviceIO/wsdl"

// Device IO service errors.
var (
	// ErrInvalidDigitalInputToken is returned when digital input token is invalid.
	ErrInvalidDigitalInputToken = errors.New("invalid digital input token: cannot be empty")
	// ErrInvalidVideoOutputToken is returned when video output token is invalid.
	ErrInvalidVideoOutputToken = errors.New("invalid video output token: cannot be empty")
	// ErrInvalidSerialPortToken is returned when serial port token is invalid.
	ErrInvalidSerialPortToken = errors.New("invalid serial port token: cannot be empty")
	// ErrInvalidSerialData is returned when serial data is invalid.
	ErrInvalidSerialData = errors.New("invalid serial data: cannot be empty")
	// ErrDigitalInputConfigNil is returned when digital input config is nil.
	ErrDigitalInputConfigNil = errors.New("digital input config cannot be nil")
	// ErrSerialPortConfigNil is returned when serial port config is nil.
	ErrSerialPortConfigNil = errors.New("serial port config cannot be nil")
	// ErrVideoOutputConfigNil is returned when video output config is nil.
	ErrVideoOutputConfigNil = errors.New("video output configuration cannot be nil")
	// ErrInvalidRelayOutputToken is returned when relay output token is invalid.
	ErrInvalidRelayOutputToken = errors.New("invalid relay output token: cannot be empty")
)

// DeviceIOServiceCapabilities represents the capabilities of the device IO service.
type DeviceIOServiceCapabilities struct {
	VideoSources            int
	VideoOutputs            int
	AudioSources            int
	AudioOutputs            int
	RelayOutputs            int
	SerialPorts             int
	DigitalInputs           int
	DigitalInputOptions     bool
	SerialPortConfiguration bool
}

// DigitalInput represents a digital input.
type DigitalInput struct {
	Token     string
	IdleState DigitalIdleState
}

// DigitalIdleState represents the idle state of a digital input.
type DigitalIdleState string

// Digital idle state constants.
const (
	DigitalIdleOpen   DigitalIdleState = "open"
	DigitalIdleClosed DigitalIdleState = "closed"
)

// VideoOutput represents a video output.
type VideoOutput struct {
	Token       string
	Layout      *Layout
	Resolution  *media.VideoResolution
	RefreshRate float64
	AspectRatio string
}

// Layout represents a video output layout.
type Layout struct {
	Pane      []PaneLayout
	Extension interface{}
}

// PaneLayout represents a pane layout.
type PaneLayout struct {
	Pane string
	Area FloatRectangle
}

// FloatRectangle represents a floating point rectangle.
type FloatRectangle struct {
	Bottom float64
	Top    float64
	Right  float64
	Left   float64
}

// SerialPort represents a serial port.
type SerialPort struct {
	Token string
	Type  SerialPortType
}

// SerialPortType represents the type of a serial port.
type SerialPortType string

// Serial port type constants.
const (
	SerialPortTypeRS232   SerialPortType = "RS232"
	SerialPortTypeRS422   SerialPortType = "RS422"
	SerialPortTypeRS485   SerialPortType = "RS485"
	SerialPortTypeGeneric SerialPortType = "Generic"
)

// SerialPortConfiguration represents a serial port configuration.
type SerialPortConfiguration struct {
	Token           string
	Type            SerialPortType
	BaudRate        int
	ParityBit       ParityBit
	CharacterLength int
	StopBit         float64
}

// ParityBit represents the parity bit setting.
type ParityBit string

// Parity bit constants.
const (
	ParityNone  ParityBit = "None"
	ParityOdd   ParityBit = "Odd"
	ParityEven  ParityBit = "Even"
	ParityMark  ParityBit = "Mark"
	ParitySpace ParityBit = "Space"
)

// SerialPortConfigurationOptions represents serial port configuration options.
type SerialPortConfigurationOptions struct {
	Token               string
	BaudRateList        []int
	ParityBitList       []ParityBit
	CharacterLengthList []int
	StopBitList         []float64
}

// DigitalInputConfigurationOptions represents digital input configuration options.
type DigitalInputConfigurationOptions struct {
	IdleStateOptions []DigitalIdleState
}

// VideoOutputConfiguration represents a video output configuration.
type VideoOutputConfiguration struct {
	Token            string
	Name             string
	UseCount         int
	OutputToken      string
	ForcePersistence bool
}

// VideoOutputConfigurationOptions represents video output configuration options.
type VideoOutputConfigurationOptions struct {
	Name                  StringRange
	OutputTokensAvailable []string
}

// StringRange represents a range of string values.
type StringRange struct {
	Min int
	Max int
}

// RelayOutputOptions represents relay output configuration options.
type RelayOutputOptions struct {
	Token      string
	Mode       []RelayMode
	DelayTimes []string
	Discrete   bool
}

// getDeviceIOEndpoint returns the device IO endpoint.
func (s *Service) getDeviceIOEndpoint() string {
	// Device IO typically uses the main device endpoint.
	return s.c.EndpointFor(api.ServiceDevice)
}

// GetDeviceIOServiceCapabilities retrieves the capabilities of the device IO service.
func (s *Service) GetDeviceIOServiceCapabilities(ctx context.Context) (*DeviceIOServiceCapabilities, error) {
	endpoint := s.getDeviceIOEndpoint()

	type GetServiceCapabilities struct {
		XMLName xml.Name `xml:"tmd:GetServiceCapabilities"`
		Xmlns   string   `xml:"xmlns:tmd,attr"`
	}

	type GetServiceCapabilitiesResponse struct {
		XMLName xml.Name `xml:"GetServiceCapabilitiesResponse"`

		Capabilities struct {
			VideoSources            int  `xml:"VideoSources,attr"`
			VideoOutputs            int  `xml:"VideoOutputs,attr"`
			AudioSources            int  `xml:"AudioSources,attr"`
			AudioOutputs            int  `xml:"AudioOutputs,attr"`
			RelayOutputs            int  `xml:"RelayOutputs,attr"`
			SerialPorts             int  `xml:"SerialPorts,attr"`
			DigitalInputs           int  `xml:"DigitalInputs,attr"`
			DigitalInputOptions     bool `xml:"DigitalInputOptions,attr"`
			SerialPortConfiguration bool `xml:"SerialPortConfiguration,attr"`
		} `xml:"Capabilities"`
	}

	req := GetServiceCapabilities{
		Xmlns: deviceIONamespace,
	}

	var resp GetServiceCapabilitiesResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetDeviceIOServiceCapabilities failed: %w", err)
	}

	return &DeviceIOServiceCapabilities{
		VideoSources:            resp.Capabilities.VideoSources,
		VideoOutputs:            resp.Capabilities.VideoOutputs,
		AudioSources:            resp.Capabilities.AudioSources,
		AudioOutputs:            resp.Capabilities.AudioOutputs,
		RelayOutputs:            resp.Capabilities.RelayOutputs,
		SerialPorts:             resp.Capabilities.SerialPorts,
		DigitalInputs:           resp.Capabilities.DigitalInputs,
		DigitalInputOptions:     resp.Capabilities.DigitalInputOptions,
		SerialPortConfiguration: resp.Capabilities.SerialPortConfiguration,
	}, nil
}

// GetDigitalInputs retrieves all digital inputs.
func (s *Service) GetDigitalInputs(ctx context.Context) ([]*DigitalInput, error) {
	endpoint := s.getDeviceIOEndpoint()

	type GetDigitalInputs struct {
		XMLName xml.Name `xml:"tmd:GetDigitalInputs"`
		Xmlns   string   `xml:"xmlns:tmd,attr"`
	}

	type GetDigitalInputsResponse struct {
		XMLName       xml.Name `xml:"GetDigitalInputsResponse"`
		DigitalInputs []struct {
			Token     string `xml:"token,attr"`
			IdleState string `xml:"IdleState,attr"`
		} `xml:"DigitalInputs"`
	}

	req := GetDigitalInputs{
		Xmlns: deviceIONamespace,
	}

	var resp GetDigitalInputsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetDigitalInputs failed: %w", err)
	}

	inputs := make([]*DigitalInput, len(resp.DigitalInputs))
	for i, di := range resp.DigitalInputs {
		inputs[i] = &DigitalInput{
			Token:     di.Token,
			IdleState: DigitalIdleState(di.IdleState),
		}
	}

	return inputs, nil
}

// GetDigitalInputConfigurationOptions retrieves digital input configuration options.
func (s *Service) GetDigitalInputConfigurationOptions(ctx context.Context, token string) (*DigitalInputConfigurationOptions, error) {
	if token == "" {
		return nil, ErrInvalidDigitalInputToken
	}

	endpoint := s.getDeviceIOEndpoint()

	type GetDigitalInputConfigurationOptions struct {
		XMLName xml.Name `xml:"tmd:GetDigitalInputConfigurationOptions"`
		Xmlns   string   `xml:"xmlns:tmd,attr"`
		Token   string   `xml:"tmd:Token"`
	}

	type GetDigitalInputConfigurationOptionsResponse struct {
		XMLName                          xml.Name `xml:"GetDigitalInputConfigurationOptionsResponse"`
		DigitalInputConfigurationOptions struct {
			IdleState []string `xml:"IdleState"`
		} `xml:"DigitalInputConfigurationOptions"`
	}

	req := GetDigitalInputConfigurationOptions{
		Xmlns: deviceIONamespace,
		Token: token,
	}

	var resp GetDigitalInputConfigurationOptionsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetDigitalInputConfigurationOptions failed: %w", err)
	}

	options := &DigitalInputConfigurationOptions{
		IdleStateOptions: make([]DigitalIdleState, len(resp.DigitalInputConfigurationOptions.IdleState)),
	}

	for i, state := range resp.DigitalInputConfigurationOptions.IdleState {
		options.IdleStateOptions[i] = DigitalIdleState(state)
	}

	return options, nil
}

// SetDigitalInputConfigurations sets digital input configurations.
func (s *Service) SetDigitalInputConfigurations(ctx context.Context, inputs []*DigitalInput) error {
	if len(inputs) == 0 {
		return ErrDigitalInputConfigNil
	}

	endpoint := s.getDeviceIOEndpoint()

	type DigitalInputXML struct {
		Token     string `xml:"token,attr"`
		IdleState string `xml:"IdleState,attr,omitempty"`
	}

	type SetDigitalInputConfigurations struct {
		XMLName       xml.Name          `xml:"tmd:SetDigitalInputConfigurations"`
		Xmlns         string            `xml:"xmlns:tmd,attr"`
		DigitalInputs []DigitalInputXML `xml:"tmd:DigitalInputs"`
	}

	type SetDigitalInputConfigurationsResponse struct {
		XMLName xml.Name `xml:"SetDigitalInputConfigurationsResponse"`
	}

	digitalInputsXML := make([]DigitalInputXML, len(inputs))
	for i, input := range inputs {
		if input.Token == "" {
			return ErrInvalidDigitalInputToken
		}

		digitalInputsXML[i] = DigitalInputXML{
			Token:     input.Token,
			IdleState: string(input.IdleState),
		}
	}

	req := SetDigitalInputConfigurations{
		Xmlns:         deviceIONamespace,
		DigitalInputs: digitalInputsXML,
	}

	var resp SetDigitalInputConfigurationsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return fmt.Errorf("SetDigitalInputConfigurations failed: %w", err)
	}

	return nil
}

// GetVideoOutputs retrieves all video outputs.
func (s *Service) GetVideoOutputs(ctx context.Context) ([]*VideoOutput, error) {
	endpoint := s.getDeviceIOEndpoint()

	type GetVideoOutputs struct {
		XMLName xml.Name `xml:"tmd:GetVideoOutputs"`
		Xmlns   string   `xml:"xmlns:tmd,attr"`
	}

	type GetVideoOutputsResponse struct {
		XMLName      xml.Name `xml:"GetVideoOutputsResponse"`
		VideoOutputs []struct {
			Token  string `xml:"token,attr"`
			Layout *struct {
				Pane []struct {
					Pane string `xml:"Pane,attr"`
					Area struct {
						Bottom float64 `xml:"bottom,attr"`
						Top    float64 `xml:"top,attr"`
						Right  float64 `xml:"right,attr"`
						Left   float64 `xml:"left,attr"`
					} `xml:"Area"`
				} `xml:"Pane"`
			} `xml:"Layout"`
			Resolution *struct {
				Width  int `xml:"Width"`
				Height int `xml:"Height"`
			} `xml:"Resolution"`
			RefreshRate float64 `xml:"RefreshRate"`
			AspectRatio string  `xml:"AspectRatio"`
		} `xml:"VideoOutputs"`
	}

	req := GetVideoOutputs{
		Xmlns: deviceIONamespace,
	}

	var resp GetVideoOutputsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoOutputs failed: %w", err)
	}

	outputs := make([]*VideoOutput, len(resp.VideoOutputs))
	for i, vo := range resp.VideoOutputs {
		output := &VideoOutput{
			Token:       vo.Token,
			RefreshRate: vo.RefreshRate,
			AspectRatio: vo.AspectRatio,
		}

		if vo.Resolution != nil {
			output.Resolution = &media.VideoResolution{
				Width:  vo.Resolution.Width,
				Height: vo.Resolution.Height,
			}
		}

		if vo.Layout != nil {
			output.Layout = &Layout{
				Pane: make([]PaneLayout, len(vo.Layout.Pane)),
			}

			for j, pane := range vo.Layout.Pane {
				output.Layout.Pane[j] = PaneLayout{
					Pane: pane.Pane,
					Area: FloatRectangle{
						Bottom: pane.Area.Bottom,
						Top:    pane.Area.Top,
						Right:  pane.Area.Right,
						Left:   pane.Area.Left,
					},
				}
			}
		}

		outputs[i] = output
	}

	return outputs, nil
}

// GetSerialPorts retrieves all serial ports.
func (s *Service) GetSerialPorts(ctx context.Context) ([]*SerialPort, error) {
	endpoint := s.getDeviceIOEndpoint()

	type GetSerialPorts struct {
		XMLName xml.Name `xml:"tmd:GetSerialPorts"`
		Xmlns   string   `xml:"xmlns:tmd,attr"`
	}

	type GetSerialPortsResponse struct {
		XMLName     xml.Name `xml:"GetSerialPortsResponse"`
		SerialPorts []struct {
			Token string `xml:"token,attr"`
			Type  string `xml:"Type"`
		} `xml:"SerialPorts"`
	}

	req := GetSerialPorts{
		Xmlns: deviceIONamespace,
	}

	var resp GetSerialPortsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSerialPorts failed: %w", err)
	}

	ports := make([]*SerialPort, len(resp.SerialPorts))
	for i, sp := range resp.SerialPorts {
		ports[i] = &SerialPort{
			Token: sp.Token,
			Type:  SerialPortType(sp.Type),
		}
	}

	return ports, nil
}

// GetSerialPortConfiguration retrieves a serial port configuration.
func (s *Service) GetSerialPortConfiguration(ctx context.Context, serialPortToken string) (*SerialPortConfiguration, error) {
	if serialPortToken == "" {
		return nil, ErrInvalidSerialPortToken
	}

	endpoint := s.getDeviceIOEndpoint()

	type GetSerialPortConfiguration struct {
		XMLName         xml.Name `xml:"tmd:GetSerialPortConfiguration"`
		Xmlns           string   `xml:"xmlns:tmd,attr"`
		SerialPortToken string   `xml:"tmd:SerialPortToken"`
	}

	type GetSerialPortConfigurationResponse struct {
		XMLName                 xml.Name `xml:"GetSerialPortConfigurationResponse"`
		SerialPortConfiguration struct {
			Token           string  `xml:"token,attr"`
			Type            string  `xml:"Type"`
			BaudRate        int     `xml:"BaudRate"`
			ParityBit       string  `xml:"ParityBit"`
			CharacterLength int     `xml:"CharacterLength"`
			StopBit         float64 `xml:"StopBit"`
		} `xml:"SerialPortConfiguration"`
	}

	req := GetSerialPortConfiguration{
		Xmlns:           deviceIONamespace,
		SerialPortToken: serialPortToken,
	}

	var resp GetSerialPortConfigurationResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSerialPortConfiguration failed: %w", err)
	}

	return &SerialPortConfiguration{
		Token:           resp.SerialPortConfiguration.Token,
		Type:            SerialPortType(resp.SerialPortConfiguration.Type),
		BaudRate:        resp.SerialPortConfiguration.BaudRate,
		ParityBit:       ParityBit(resp.SerialPortConfiguration.ParityBit),
		CharacterLength: resp.SerialPortConfiguration.CharacterLength,
		StopBit:         resp.SerialPortConfiguration.StopBit,
	}, nil
}

// GetSerialPortConfigurationOptions retrieves serial port configuration options.
func (s *Service) GetSerialPortConfigurationOptions(ctx context.Context, serialPortToken string) (*SerialPortConfigurationOptions, error) {
	if serialPortToken == "" {
		return nil, ErrInvalidSerialPortToken
	}

	endpoint := s.getDeviceIOEndpoint()

	type GetSerialPortConfigurationOptions struct {
		XMLName         xml.Name `xml:"tmd:GetSerialPortConfigurationOptions"`
		Xmlns           string   `xml:"xmlns:tmd,attr"`
		SerialPortToken string   `xml:"tmd:SerialPortToken"`
	}

	type GetSerialPortConfigurationOptionsResponse struct {
		XMLName                        xml.Name `xml:"GetSerialPortConfigurationOptionsResponse"`
		SerialPortConfigurationOptions struct {
			Token          string    `xml:"token,attr"`
			BaudRateList   []int     `xml:"BaudRateList>Items"`
			ParityBitList  []string  `xml:"ParityBitList>Items"`
			CharLengthList []int     `xml:"CharacterLengthList>Items"`
			StopBitList    []float64 `xml:"StopBitList>Items"`
		} `xml:"SerialPortConfigurationOptions"`
	}

	req := GetSerialPortConfigurationOptions{
		Xmlns:           deviceIONamespace,
		SerialPortToken: serialPortToken,
	}

	var resp GetSerialPortConfigurationOptionsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetSerialPortConfigurationOptions failed: %w", err)
	}

	options := &SerialPortConfigurationOptions{
		Token:               resp.SerialPortConfigurationOptions.Token,
		BaudRateList:        resp.SerialPortConfigurationOptions.BaudRateList,
		CharacterLengthList: resp.SerialPortConfigurationOptions.CharLengthList,
		StopBitList:         resp.SerialPortConfigurationOptions.StopBitList,
	}

	// Convert parity bit strings to ParityBit type.
	options.ParityBitList = make([]ParityBit, len(resp.SerialPortConfigurationOptions.ParityBitList))
	for i, pb := range resp.SerialPortConfigurationOptions.ParityBitList {
		options.ParityBitList[i] = ParityBit(pb)
	}

	return options, nil
}

// SetSerialPortConfiguration sets a serial port configuration.
func (s *Service) SetSerialPortConfiguration(ctx context.Context, config *SerialPortConfiguration) error {
	if config == nil {
		return ErrSerialPortConfigNil
	}

	if config.Token == "" {
		return ErrInvalidSerialPortToken
	}

	endpoint := s.getDeviceIOEndpoint()

	type SerialPortConfigurationXML struct {
		Token           string  `xml:"token,attr"`
		Type            string  `xml:"tmd:Type"`
		BaudRate        int     `xml:"tmd:BaudRate"`
		ParityBit       string  `xml:"tmd:ParityBit"`
		CharacterLength int     `xml:"tmd:CharacterLength"`
		StopBit         float64 `xml:"tmd:StopBit"`
	}

	type SetSerialPortConfiguration struct {
		XMLName                 xml.Name                   `xml:"tmd:SetSerialPortConfiguration"`
		Xmlns                   string                     `xml:"xmlns:tmd,attr"`
		SerialPortConfiguration SerialPortConfigurationXML `xml:"tmd:SerialPortConfiguration"`
	}

	type SetSerialPortConfigurationResponse struct {
		XMLName xml.Name `xml:"SetSerialPortConfigurationResponse"`
	}

	req := SetSerialPortConfiguration{
		Xmlns: deviceIONamespace,
		SerialPortConfiguration: SerialPortConfigurationXML{
			Token:           config.Token,
			Type:            string(config.Type),
			BaudRate:        config.BaudRate,
			ParityBit:       string(config.ParityBit),
			CharacterLength: config.CharacterLength,
			StopBit:         config.StopBit,
		},
	}

	var resp SetSerialPortConfigurationResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return fmt.Errorf("SetSerialPortConfiguration failed: %w", err)
	}

	return nil
}

// SendReceiveSerialCommand sends a serial command and receives a response.
func (s *Service) SendReceiveSerialCommand(ctx context.Context, serialPortToken string, data []byte, timeoutSeconds, dataLength int) ([]byte, error) {
	if serialPortToken == "" {
		return nil, ErrInvalidSerialPortToken
	}

	if len(data) == 0 {
		return nil, ErrInvalidSerialData
	}

	endpoint := s.getDeviceIOEndpoint()

	type SerialData struct {
		Binary string `xml:"tt:Binary,omitempty"`
	}

	type SendReceiveSerialCommand struct {
		XMLName    xml.Name   `xml:"tmd:SendReceiveSerialCommand"`
		Xmlns      string     `xml:"xmlns:tmd,attr"`
		XmlnsTT    string     `xml:"xmlns:tt,attr"`
		Token      string     `xml:"tmd:Token"`
		SerialData SerialData `xml:"tmd:SerialData"`
		TimeOut    string     `xml:"tmd:TimeOut,omitempty"`
		DataLength int        `xml:"tmd:DataLength,omitempty"`
	}

	type SendReceiveSerialCommandResponse struct {
		XMLName    xml.Name `xml:"SendReceiveSerialCommandResponse"`
		SerialData struct {
			Binary string `xml:"Binary"`
		} `xml:"SerialData"`
	}

	req := SendReceiveSerialCommand{
		Xmlns:   deviceIONamespace,
		XmlnsTT: "http://www.onvif.org/ver10/schema",
		Token:   serialPortToken,
		SerialData: SerialData{
			Binary: string(data),
		},
		DataLength: dataLength,
	}

	if timeoutSeconds > 0 {
		req.TimeOut = fmt.Sprintf("PT%dS", timeoutSeconds)
	}

	var resp SendReceiveSerialCommandResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("SendReceiveSerialCommand failed: %w", err)
	}

	return []byte(resp.SerialData.Binary), nil
}

// GetVideoOutputConfiguration retrieves a video output configuration.
func (s *Service) GetVideoOutputConfiguration(ctx context.Context, videoOutputToken string) (*VideoOutputConfiguration, error) {
	if videoOutputToken == "" {
		return nil, ErrInvalidVideoOutputToken
	}

	endpoint := s.getDeviceIOEndpoint()

	type GetVideoOutputConfiguration struct {
		XMLName          xml.Name `xml:"tmd:GetVideoOutputConfiguration"`
		Xmlns            string   `xml:"xmlns:tmd,attr"`
		VideoOutputToken string   `xml:"tmd:VideoOutputToken"`
	}

	type GetVideoOutputConfigurationResponse struct {
		XMLName                  xml.Name `xml:"GetVideoOutputConfigurationResponse"`
		VideoOutputConfiguration struct {
			Token       string `xml:"token,attr"`
			Name        string `xml:"Name"`
			UseCount    int    `xml:"UseCount"`
			OutputToken string `xml:"OutputToken"`
		} `xml:"VideoOutputConfiguration"`
	}

	req := GetVideoOutputConfiguration{
		Xmlns:            deviceIONamespace,
		VideoOutputToken: videoOutputToken,
	}

	var resp GetVideoOutputConfigurationResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoOutputConfiguration failed: %w", err)
	}

	return &VideoOutputConfiguration{
		Token:       resp.VideoOutputConfiguration.Token,
		Name:        resp.VideoOutputConfiguration.Name,
		UseCount:    resp.VideoOutputConfiguration.UseCount,
		OutputToken: resp.VideoOutputConfiguration.OutputToken,
	}, nil
}

// GetVideoOutputConfigurationOptions retrieves video output configuration options.
func (s *Service) GetVideoOutputConfigurationOptions(ctx context.Context, videoOutputToken string) (*VideoOutputConfigurationOptions, error) {
	if videoOutputToken == "" {
		return nil, ErrInvalidVideoOutputToken
	}

	endpoint := s.getDeviceIOEndpoint()

	type GetVideoOutputConfigurationOptions struct {
		XMLName          xml.Name `xml:"tmd:GetVideoOutputConfigurationOptions"`
		Xmlns            string   `xml:"xmlns:tmd,attr"`
		VideoOutputToken string   `xml:"tmd:VideoOutputToken"`
	}

	type GetVideoOutputConfigurationOptionsResponse struct {
		XMLName                         xml.Name `xml:"GetVideoOutputConfigurationOptionsResponse"`
		VideoOutputConfigurationOptions struct {
			Name struct {
				Min int `xml:"Min,attr"`
				Max int `xml:"Max,attr"`
			} `xml:"Name"`
			OutputTokensAvailable []string `xml:"OutputTokensAvailable"`
		} `xml:"VideoOutputConfigurationOptions"`
	}

	req := GetVideoOutputConfigurationOptions{
		Xmlns:            deviceIONamespace,
		VideoOutputToken: videoOutputToken,
	}

	var resp GetVideoOutputConfigurationOptionsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetVideoOutputConfigurationOptions failed: %w", err)
	}

	return &VideoOutputConfigurationOptions{
		Name: StringRange{
			Min: resp.VideoOutputConfigurationOptions.Name.Min,
			Max: resp.VideoOutputConfigurationOptions.Name.Max,
		},
		OutputTokensAvailable: resp.VideoOutputConfigurationOptions.OutputTokensAvailable,
	}, nil
}

// SetVideoOutputConfiguration sets a video output configuration.
func (s *Service) SetVideoOutputConfiguration(ctx context.Context, config *VideoOutputConfiguration) error {
	if config == nil {
		return ErrVideoOutputConfigNil
	}

	if config.Token == "" {
		return ErrInvalidVideoOutputToken
	}

	endpoint := s.getDeviceIOEndpoint()

	type VideoOutputConfigurationXML struct {
		Token       string `xml:"token,attr"`
		Name        string `xml:"tt:Name"`
		UseCount    int    `xml:"tt:UseCount"`
		OutputToken string `xml:"tt:OutputToken"`
	}

	type SetVideoOutputConfiguration struct {
		XMLName          xml.Name                    `xml:"tmd:SetVideoOutputConfiguration"`
		Xmlns            string                      `xml:"xmlns:tmd,attr"`
		XmlnsTT          string                      `xml:"xmlns:tt,attr"`
		Configuration    VideoOutputConfigurationXML `xml:"tmd:Configuration"`
		ForcePersistence bool                        `xml:"tmd:ForcePersistence"`
	}

	type SetVideoOutputConfigurationResponse struct {
		XMLName xml.Name `xml:"SetVideoOutputConfigurationResponse"`
	}

	req := SetVideoOutputConfiguration{
		Xmlns:   deviceIONamespace,
		XmlnsTT: "http://www.onvif.org/ver10/schema",
		Configuration: VideoOutputConfigurationXML{
			Token:       config.Token,
			Name:        config.Name,
			UseCount:    config.UseCount,
			OutputToken: config.OutputToken,
		},
		ForcePersistence: config.ForcePersistence,
	}

	var resp SetVideoOutputConfigurationResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return fmt.Errorf("SetVideoOutputConfiguration failed: %w", err)
	}

	return nil
}

// GetRelayOutputOptions retrieves relay output options.
func (s *Service) GetRelayOutputOptions(ctx context.Context, relayOutputToken string) (*RelayOutputOptions, error) {
	if relayOutputToken == "" {
		return nil, ErrInvalidRelayOutputToken
	}

	endpoint := s.getDeviceIOEndpoint()

	type GetRelayOutputOptions struct {
		XMLName          xml.Name `xml:"tmd:GetRelayOutputOptions"`
		Xmlns            string   `xml:"xmlns:tmd,attr"`
		RelayOutputToken string   `xml:"tmd:RelayOutputToken"`
	}

	type GetRelayOutputOptionsResponse struct {
		XMLName            xml.Name `xml:"GetRelayOutputOptionsResponse"`
		RelayOutputOptions struct {
			Token      string   `xml:"token,attr"`
			Mode       []string `xml:"Mode"`
			DelayTimes []string `xml:"DelayTimes"`
			Discrete   bool     `xml:"Discrete"`
		} `xml:"RelayOutputOptions"`
	}

	req := GetRelayOutputOptions{
		Xmlns:            deviceIONamespace,
		RelayOutputToken: relayOutputToken,
	}

	var resp GetRelayOutputOptionsResponse

	if err := s.c.Call(ctx, endpoint, "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetRelayOutputOptions failed: %w", err)
	}

	modes := make([]RelayMode, len(resp.RelayOutputOptions.Mode))
	for i, m := range resp.RelayOutputOptions.Mode {
		modes[i] = RelayMode(m)
	}

	return &RelayOutputOptions{
		Token:      resp.RelayOutputOptions.Token,
		Mode:       modes,
		DelayTimes: resp.RelayOutputOptions.DelayTimes,
		Discrete:   resp.RelayOutputOptions.Discrete,
	}, nil
}

// GetRelayOutputs gets a list of all available relay outputs and their settings.
func (s *Service) GetRelayOutputs(ctx context.Context) ([]*RelayOutput, error) {
	type GetRelayOutputs struct {
		XMLName xml.Name `xml:"tds:GetRelayOutputs"`
		Xmlns   string   `xml:"xmlns:tds,attr"`
	}

	type GetRelayOutputsResponse struct {
		XMLName      xml.Name `xml:"GetRelayOutputsResponse"`
		RelayOutputs []struct {
			Token      string `xml:"token,attr"`
			Properties struct {
				Mode      string `xml:"Mode"`
				DelayTime string `xml:"DelayTime"`
				IdleState string `xml:"IdleState"`
			} `xml:"Properties"`
		} `xml:"RelayOutputs"`
	}

	req := GetRelayOutputs{
		Xmlns: Namespace,
	}

	var resp GetRelayOutputsResponse

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, &resp); err != nil {
		return nil, fmt.Errorf("GetRelayOutputs failed: %w", err)
	}

	relays := make([]*RelayOutput, len(resp.RelayOutputs))
	for i, relay := range resp.RelayOutputs {
		relays[i] = &RelayOutput{
			Token: relay.Token,
			Properties: RelayOutputSettings{
				Mode:      RelayMode(relay.Properties.Mode),
				IdleState: RelayIdleState(relay.Properties.IdleState),
				// DelayTime parsing would require duration parsing
			},
		}
	}

	return relays, nil
}

// SetRelayOutputSettings sets the settings of a relay output.
// SetRelayOutputSettings sets the settings of a relay output.
func (s *Service) SetRelayOutputSettings(ctx context.Context, token string, settings *RelayOutputSettings) error {
	type SetRelayOutputSettings struct {
		XMLName          xml.Name `xml:"tds:SetRelayOutputSettings"`
		Xmlns            string   `xml:"xmlns:tds,attr"`
		RelayOutputToken string   `xml:"tds:RelayOutputToken"`
		Properties       struct {
			Mode      string `xml:"tt:Mode"`
			DelayTime string `xml:"tt:DelayTime"`
			IdleState string `xml:"tt:IdleState"`
		} `xml:"tds:Properties"`
	}

	req := SetRelayOutputSettings{
		Xmlns:            Namespace,
		RelayOutputToken: token,
	}
	req.Properties.Mode = string(settings.Mode)
	req.Properties.IdleState = string(settings.IdleState)
	// DelayTime would need duration formatting

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetRelayOutputSettings failed: %w", err)
	}

	return nil
}

// SetRelayOutputState sets the state of a relay output.
// SetRelayOutputState sets the state of a relay output.
func (s *Service) SetRelayOutputState(ctx context.Context, token string, state RelayLogicalState) error {
	type SetRelayOutputState struct {
		XMLName          xml.Name          `xml:"tds:SetRelayOutputState"`
		Xmlns            string            `xml:"xmlns:tds,attr"`
		RelayOutputToken string            `xml:"tds:RelayOutputToken"`
		LogicalState     RelayLogicalState `xml:"tds:LogicalState"`
	}

	req := SetRelayOutputState{
		Xmlns:            Namespace,
		RelayOutputToken: token,
		LogicalState:     state,
	}

	if err := s.c.Call(ctx, s.c.EndpointFor(api.ServiceDevice), "", req, nil); err != nil {
		return fmt.Errorf("SetRelayOutputState failed: %w", err)
	}

	return nil
}

// SendAuxiliaryCommand sends an auxiliary command to the device.
