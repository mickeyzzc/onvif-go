package provider

// OSD is one on-screen overlay text bound to a video source
// configuration (the minimal Text/String OSD closed loop).
type OSD struct {
	Token                         string
	VideoSourceConfigurationToken string
	Type                          string // "Text" for the minimal loop
	Text                          string
}

// OSDStore is the host seam behind the media OSD family: list, upsert
// (create when the token is empty), delete.
type OSDStore interface {
	OSDs(videoSourceConfigurationToken string) []OSD
	UpsertOSD(osd OSD) OSD
	DeleteOSD(token string) bool
}

// RelayOutput is one physical alarm relay with its current logical
// state — the NVR alarm-linkage surface.
type RelayOutput struct {
	Token        string
	Mode         string // Bistable | Monostable
	DelayTime    string // xs:duration, e.g. "PT1S"
	IdleState    string // closed | open
	LogicalState string // active | inactive
}

// DigitalInput is one alarm input line.
type DigitalInput struct {
	Token     string
	IdleState string // closed | open
}

// RelayController is the host seam behind the DeviceIO relay family.
type RelayController interface {
	RelayOutputs() []RelayOutput
	SetRelayState(token, logicalState string) error
	DigitalInputs() []DigitalInput
}
