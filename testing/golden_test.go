package onviftesting

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTokenGetters(t *testing.T) {
	exchange := &CapturedExchangeV2{
		Parameters: map[string]interface{}{
			"ProfileToken":       "profile-1",
			"Token":              "config-9",
			"VideoSourceToken":   "vs-1",
			"AudioSourceToken":   "as-1",
			"PresetToken":        "preset-1",
			"NodeToken":          "node-1",
			"OSDToken":           "osd-1",
			"ConfigurationToken": "config-1",
		},
	}

	if got := exchange.GetProfileToken(); got != "profile-1" {
		t.Errorf("ProfileToken = %q", got)
	}

	if got := exchange.GetConfigurationToken(); got != "config-1" {
		t.Errorf("ConfigurationToken = %q", got)
	}

	if got := exchange.GetVideoSourceToken(); got != "vs-1" {
		t.Errorf("VideoSourceToken = %q", got)
	}

	if got := exchange.GetAudioSourceToken(); got != "as-1" {
		t.Errorf("AudioSourceToken = %q", got)
	}

	if got := exchange.GetPresetToken(); got != "preset-1" {
		t.Errorf("PresetToken = %q", got)
	}

	if got := exchange.GetNodeToken(); got != "node-1" {
		t.Errorf("NodeToken = %q", got)
	}

	if got := exchange.GetOSDToken(); got != "osd-1" {
		t.Errorf("OSDToken = %q", got)
	}

	// Token fallback when ConfigurationToken is absent.
	fallback := &CapturedExchangeV2{Parameters: map[string]interface{}{"Token": "t-9"}}
	if got := fallback.GetConfigurationToken(); got != "t-9" {
		t.Errorf("fallback Token = %q", got)
	}

	// nil parameters: every getter degrades to "".
	empty := &CapturedExchangeV2{}
	for name, got := range map[string]string{
		"profile": empty.GetProfileToken(),
		"config":  empty.GetConfigurationToken(),
		"video":   empty.GetVideoSourceToken(),
		"audio":   empty.GetAudioSourceToken(),
		"preset":  empty.GetPresetToken(),
		"node":    empty.GetNodeToken(),
		"osd":     empty.GetOSDToken(),
	} {
		if got != "" {
			t.Errorf("%s on nil params = %q, want empty", name, got)
		}
	}
}

func TestMatchKeyString(t *testing.T) {
	key := MatchKey{
		OperationName: "GetStreamUri",
		ProfileToken:  "p1",
	}
	if s := key.String(); s != "GetStreamUri[Profile:p1]" {
		t.Errorf("String() = %q", s)
	}

	if s := (MatchKey{OperationName: "Op"}).String(); s != "Op" {
		t.Errorf("bare String() = %q", s)
	}
}

func TestBuildMatchKeyFromExchange(t *testing.T) {
	exchange := &CapturedExchangeV2{
		OperationName: "GetProfiles",
		Parameters: map[string]interface{}{
			"ProfileToken": "profile-7",
		},
	}

	key := BuildMatchKeyFromExchange(exchange)
	if key.OperationName != "GetProfiles" || key.ProfileToken != "profile-7" {
		t.Errorf("key = %+v", key)
	}
}

func TestGoldenRoundTrip(t *testing.T) {
	dir := t.TempDir()

	golden := &GoldenFile{
		Operation:  "GetProfiles",
		Service:    "media",
		Parameters: map[string]string{"ProfileToken": "p1"},
		Request:    "<req/>",
		Response:   "<resp/>",
		ExpectedFields: map[string]interface{}{
			"Profiles": nil,
		},
		VariableFields: []string{"UtcTime"},
	}

	path := filepath.Join(dir, GenerateGoldenFileName(golden.Operation, golden.Parameters))
	if err := SaveGoldenFile(golden, path); err != nil {
		t.Fatalf("SaveGoldenFile: %v", err)
	}

	set, err := LoadGoldenFiles(dir)
	if err != nil {
		t.Fatalf("LoadGoldenFiles: %v", err)
	}

	if len(set.Files) != 1 {
		t.Fatalf("files = %d, want 1", len(set.Files))
	}

	loaded := set.GetGoldenFile("GetProfiles", map[string]string{"ProfileToken": "p1"})
	if loaded == nil || loaded.Response != "<resp/>" {
		t.Fatalf("GetGoldenFile = %+v", loaded)
	}

	// Parameterless lookup falls back to operation-only match.
	if g := set.GetGoldenFile("GetProfiles", nil); g == nil {
		t.Error("operation-only fallback failed")
	}

	if g := set.GetGoldenFile("Missing", nil); g != nil {
		t.Error("unknown operation must return nil")
	}

	ops := set.GetOperations()
	if len(ops) != 1 || ops[0] != "GetProfiles" {
		t.Errorf("operations = %v", ops)
	}

	// Manifest persistence.
	manifest := &GoldenManifest{Version: "2.0"}
	manifestPath := filepath.Join(dir, "manifest.json")

	if err := SaveGoldenManifest(manifest, manifestPath); err != nil {
		t.Fatalf("SaveGoldenManifest: %v", err)
	}

	loadedManifest, err := LoadGoldenManifest(dir)
	if err != nil || loadedManifest.Version != "2.0" {
		t.Errorf("manifest = %+v, %v", loadedManifest, err)
	}
}

func TestBuildGoldenKeyDeterministic(t *testing.T) {
	a := &GoldenFile{Operation: "Op", Parameters: map[string]string{"B": "2", "A": "1"}}
	b := &GoldenFile{Operation: "Op", Parameters: map[string]string{"A": "1", "B": "2"}}

	if buildGoldenKey(a) != buildGoldenKey(b) {
		t.Error("key must not depend on map iteration order")
	}
}

func TestValidateResponse(t *testing.T) {
	type resp struct {
		Manufacturer string `json:"Manufacturer"`
		Model        string `json:"Model"`
	}

	golden := &GoldenFile{
		ExpectedFields: map[string]interface{}{
			"Manufacturer": "ACME",
			"Model":        nil,
		},
		VariableFields: []string{"Model"},
	}

	if errs := ValidateResponse(resp{Manufacturer: "ACME", Model: "anything"}, golden); len(errs) != 0 {
		t.Errorf("unexpected mismatches: %v (variable fields must be skipped)", errs)
	}

	if errs := ValidateResponse(resp{Manufacturer: "Other", Model: "x"}, golden); len(errs) != 1 {
		t.Errorf("mismatches = %v, want the Manufacturer delta", errs)
	}

	if errs := ValidateResponse(resp{}, &GoldenFile{}); errs != nil {
		t.Errorf("no expectations = no errors, got %v", errs)
	}
}

func TestValuesEqual(t *testing.T) {
	if !valuesEqual(map[string]int{"a": 1}, map[string]int{"a": 1}) {
		t.Error("equal maps must compare equal")
	}

	if valuesEqual(1, 2) {
		t.Error("different values must not compare equal")
	}

	if valuesEqual(nil, nil) != true || valuesEqual(nil, 1) != false {
		t.Error("nil comparisons broken")
	}
}

func TestIsVariableField(t *testing.T) {
	if !isVariableField("UtcTime", []string{"UtcTime"}) || isVariableField("Token", nil) {
		t.Error("variable-field matching broken")
	}
}

func TestCreateGoldenFromCapture(t *testing.T) {
	capture := &CapturedExchangeV2{
		OperationName: "GetDeviceInformation",
		ServiceType:   ServiceDevice,
		RequestBody:   "<req/>",
		ResponseBody:  "<resp/>",
		Parameters:    map[string]interface{}{},
	}

	golden := CreateGoldenFromCapture(capture)
	if golden.Operation != "GetDeviceInformation" || golden.Response != "<resp/>" {
		t.Errorf("golden = %+v", golden)
	}
}

func TestGoldenTestRunner(t *testing.T) {
	dir := t.TempDir()

	golden := &GoldenFile{
		Operation:      "GetDeviceInformation",
		Service:        "device",
		Request:        "<req/>",
		Response:       "<resp/>",
		ExpectedFields: map[string]interface{}{"Manufacturer": "ACME"},
	}

	if err := SaveGoldenFile(golden, filepath.Join(dir, "gd.json")); err != nil {
		t.Fatalf("save: %v", err)
	}

	runner, err := NewGoldenTestRunner(dir)
	if err != nil {
		t.Fatalf("NewGoldenTestRunner: %v", err)
	}

	errs := runner.ValidateOperation("GetDeviceInformation", nil, map[string]string{
		"Manufacturer": "ACME",
	})
	if len(errs) != 0 {
		t.Errorf("ValidateOperation errors = %v", errs)
	}

	errs = runner.ValidateOperation("GetDeviceInformation", nil, map[string]string{
		"Manufacturer": "Other",
	})
	if len(errs) != 1 {
		t.Errorf("mismatch must report exactly the delta, got %v", errs)
	}

	if errs := runner.ValidateOperation("Nope", nil, nil); len(errs) != 1 {
		t.Errorf("unknown operation must report one error, got %v", errs)
	}
}

func TestLoadGoldenManifestMissing(t *testing.T) {
	if _, err := LoadGoldenManifest(t.TempDir()); err == nil {
		t.Error("missing manifest must error")
	}
}

// Silence unused-import helpers if assertions evolve.
var (
	_ = os.ReadFile
	_ = json.Marshal
)
