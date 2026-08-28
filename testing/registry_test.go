package onviftesting

// Tests for the capture registry: load/save round-trip, entry CRUD,
// coverage aggregation, ID sanitization, validation, and entry creation
// from archives.

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistryLoadSaveRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "registry.json")

	loaded, err := LoadRegistry(path)
	if err != nil {
		t.Fatalf("LoadRegistry(missing): %v", err)
	}

	if loaded.Version != RegistryVersion || len(loaded.Cameras) != 0 {
		t.Fatalf("fresh registry = %+v", loaded)
	}

	loaded.AddCamera(&CameraEntry{ID: "cam-1", Manufacturer: "Acme", Model: "CamX", Firmware: "1.0"})
	if err := SaveRegistry(loaded, path); err != nil {
		t.Fatalf("SaveRegistry: %v", err)
	}

	reloaded, err := LoadRegistry(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(reloaded.Cameras) != 1 || reloaded.Cameras[0].ID != "cam-1" {
		t.Fatalf("roundtrip lost camera: %+v", reloaded.Cameras)
	}

	if reloaded.Cameras[0].AddedDate == "" {
		t.Error("AddCamera must stamp AddedDate")
	}
}

func TestRegistryLoadInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "registry.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadRegistry(path); err == nil {
		t.Error("invalid JSON accepted")
	}
}

func TestRegistryEntryCRUD(t *testing.T) {
	r := &Registry{}
	r.AddCamera(&CameraEntry{ID: "a", Manufacturer: "Acme"})
	r.AddCamera(&CameraEntry{ID: "b", Manufacturer: "Other"})

	// Add existing ID updates in place.
	r.AddCamera(&CameraEntry{ID: "a", Manufacturer: "Acme2"})
	if len(r.Cameras) != 2 {
		t.Fatalf("update added a duplicate: %d cameras", len(r.Cameras))
	}

	if got := r.GetCamera("b"); got == nil || got.Manufacturer != "Other" {
		t.Errorf("GetCamera(b) = %+v", got)
	}

	if r.GetCamera("missing") != nil {
		t.Error("GetCamera(missing) must be nil")
	}

	acme := r.GetCamerasByManufacturer("Acme2")
	if len(acme) != 1 || acme[0].ID != "a" {
		t.Errorf("GetCamerasByManufacturer = %+v", acme)
	}

	if !r.RemoveCamera("a") || r.GetCamera("a") != nil {
		t.Error("RemoveCamera(a) failed")
	}

	if r.RemoveCamera("a") {
		t.Error("double remove must report false")
	}
}

func TestRegistryCoverage(t *testing.T) {
	r := &Registry{}
	r.UpdateCoverage()

	if len(r.Coverage) == 0 {
		t.Fatal("UpdateCoverage produced no services")
	}

	for service, cov := range r.Coverage {
		if cov.Total == 0 {
			t.Errorf("service %s has zero total operations", service)
		}
	}

	r.Coverage = map[string]Coverage{
		"Device": {Total: 10, Captured: 5},
		"Media":  {Total: 30, Captured: 6},
	}

	total, captured := r.GetTotalCoverage()
	if total != 40 || captured != 11 {
		t.Fatalf("GetTotalCoverage = %d/%d, want 40/11", total, captured)
	}

	summary := r.GetSummary()
	if summary.TotalCameras != 0 || summary.TotalOperations != 40 {
		t.Fatalf("summary = %+v", summary)
	}

	if summary.ServiceCoverage["Device"] != 50 {
		t.Errorf("Device coverage = %v, want 50%%", summary.ServiceCoverage["Device"])
	}
}

func TestGenerateCameraIDSanitization(t *testing.T) {
	id := GenerateCameraID("Acme Corp", "Cam X-2", "1.0.5")
	if id != "acme_corp_cam_x_2_1_0_5" {
		t.Errorf("GenerateCameraID = %q", id)
	}

	// Invalid characters are dropped; space/-/_/. collapse to underscores.
	if got := sanitizeID("A B!@#C"); got != "a_bc" {
		t.Errorf("sanitizeID = %q", got)
	}
}

func TestValidateRegistry(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "ok.tar.gz"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := &Registry{Cameras: []CameraEntry{
		{ID: "good", CaptureFile: "ok.tar.gz"},
		{ID: "bad", CaptureFile: "missing.tar.gz"},
	}}

	errs := ValidateRegistry(r, dir)
	if len(errs) != 1 || errs[0] == "" {
		t.Fatalf("ValidateRegistry = %v, want one missing-file error", errs)
	}
}

func TestCreateCameraEntryFromV1Archive(t *testing.T) {
	path := writeCaptureArchive(t, map[string]string{
		"001.json": `{"operation_name":"GetDeviceInformation","request_body":"<GetDeviceInformation/>","response_body":"<GetDeviceInformationResponse><Manufacturer>Acme</Manufacturer><Model>CamX</Model><FirmwareVersion>1.0</FirmwareVersion></GetDeviceInformationResponse>","status_code":200}`,
		"002.json": `{"operation_name":"GetProfiles","request_body":"<GetProfiles/>","response_body":"<GetProfilesResponse/>","status_code":200}`,
	})

	entry, err := CreateCameraEntryFromCapture(path)
	if err != nil {
		t.Fatal(err)
	}

	if entry.Manufacturer != "Acme" || entry.Model != "CamX" || entry.Firmware != "1.0" {
		t.Errorf("camera info = %+v", entry)
	}

	if entry.OperationsCaptured != 2 {
		t.Errorf("OperationsCaptured = %d", entry.OperationsCaptured)
	}

	if entry.ID != GenerateCameraID("Acme", "CamX", "1.0") {
		t.Errorf("ID = %q", entry.ID)
	}
}

func TestDetectCapabilitiesAndServiceInference(t *testing.T) {
	capture := &CameraCaptureV2{Exchanges: []CapturedExchangeV2{
		{OperationName: "GetProfiles"},
		{OperationName: "ContinuousMove"},
		{OperationName: "GetDeviceInformation"},
	}}

	caps := detectCapabilities(capture)
	if len(caps) != 3 {
		t.Fatalf("capabilities = %v, want Media/PTZ/Device", caps)
	}

	cases := map[string]string{
		"GetProfiles":                 "Media",
		"GetStreamUri":                "Media",
		"GotoPreset":                  "PTZ",
		"ContinuousMove":              "PTZ",
		"GetImagingSettings":          "Imaging",
		"CreatePullPointSubscription": "Event",
		"GetSystemDateAndTime":        "Device",
	}

	for op, want := range cases {
		if got := inferServiceFromOperation(op); got != want {
			t.Errorf("inferServiceFromOperation(%q) = %q, want %q", op, got, want)
		}
	}
}
