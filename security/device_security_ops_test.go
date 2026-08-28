package security

// Table-driven wire-contract tests for the security service operations
// that had zero coverage (IP filters, password policy/history, auth
// failure warning, certificate lifecycle, PKCS#10, client-cert mode)
// through the internal/testutil.FakeCaller decode path.

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

func newOpsService(t *testing.T, wantAction, resp string) (*Service, *testutil.FakeCaller) {
	t.Helper()

	caller := testutil.NewFakeCaller("http://fake/device_service", func(action, _ string) (string, error) {
		if strings.TrimPrefix(action, "tds:") != wantAction {
			return "", errors.New("unexpected action " + action)
		}

		return resp, nil
	})

	return New(caller), caller
}

func TestGetRemoteUserMapping(t *testing.T) {
	s, caller := newOpsService(t, "GetRemoteUser",
		`<GetRemoteUserResponse><RemoteUser><Username>remote</Username><Password>pw</Password><UseDerivedPassword>true</UseDerivedPassword></RemoteUser></GetRemoteUserResponse>`)

	user, err := s.GetRemoteUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if user == nil || user.Username != "remote" || !user.UseDerivedPassword {
		t.Fatalf("mapping wrong: %+v", user)
	}

	if caller.CountAction("tds:GetRemoteUser") != 1 {
		t.Fatal("GetRemoteUser not sent exactly once")
	}
}

func TestGetRemoteUserNoneConfigured(t *testing.T) {
	s, _ := newOpsService(t, "GetRemoteUser", `<GetRemoteUserResponse/>`)

	user, err := s.GetRemoteUser(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if user != nil {
		t.Fatalf("no RemoteUser element must map to nil, got %+v", user)
	}
}

func TestSetRemoteUserWire(t *testing.T) {
	s, caller := newOpsService(t, "SetRemoteUser", `<SetRemoteUserResponse/>`)

	if err := s.SetRemoteUser(context.Background(), &RemoteUser{Username: "u", Password: "p"}); err != nil {
		t.Fatal(err)
	}

	if body := caller.Requests()[0].Body; !strings.Contains(body, ">u<") || !strings.Contains(body, ">p<") {
		t.Fatalf("request body missing fields: %s", body)
	}
}

func TestGetIPAddressFilterMapping(t *testing.T) {
	s, _ := newOpsService(t, "GetIPAddressFilter",
		`<GetIPAddressFilterResponse><IPAddressFilter><Type>Allow</Type><IPv4Address><Address>192.168.1.1</Address><PrefixLength>32</PrefixLength></IPv4Address></IPAddressFilter></GetIPAddressFilterResponse>`)

	filter, err := s.GetIPAddressFilter(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if filter.Type != "Allow" || len(filter.IPv4Address) != 1 || filter.IPv4Address[0].Address != "192.168.1.1" {
		t.Fatalf("mapping wrong: %+v", filter)
	}
}

func TestIPAddressFilterWriters(t *testing.T) {
	filter := &IPAddressFilter{Type: "Deny"}

	for _, tc := range []struct {
		name   string
		action string
		invoke func(s *Service) error
	}{
		{"SetIPAddressFilter", "SetIPAddressFilter", func(s *Service) error {
			return s.SetIPAddressFilter(context.Background(), filter)
		}},
		{"AddIPAddressFilter", "AddIPAddressFilter", func(s *Service) error {
			return s.AddIPAddressFilter(context.Background(), filter)
		}},
		{"RemoveIPAddressFilter", "RemoveIPAddressFilter", func(s *Service) error {
			return s.RemoveIPAddressFilter(context.Background(), filter)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := newOpsService(t, tc.action, "<"+tc.action+"Response/>")

			if err := tc.invoke(s); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPasswordPolicyRoundtrip(t *testing.T) {
	s, _ := newOpsService(t, "GetPasswordComplexityConfiguration",
		`<GetPasswordComplexityConfigurationResponse><MinLen>12</MinLen><Uppercase>1</Uppercase><Number>2</Number><SpecialChars>1</SpecialChars><BlockUsernameOccurrence>true</BlockUsernameOccurrence></GetPasswordComplexityConfigurationResponse>`)

	cfg, err := s.GetPasswordComplexityConfiguration(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if cfg.MinLen != 12 || cfg.Uppercase != 1 || cfg.Number != 2 || cfg.SpecialChars != 1 || !cfg.BlockUsernameOccurrence {
		t.Fatalf("mapping wrong: %+v", cfg)
	}

	s2, caller := newOpsService(t, "SetPasswordComplexityConfiguration", `<SetPasswordComplexityConfigurationResponse/>`)
	if err := s2.SetPasswordComplexityConfiguration(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	if body := caller.Requests()[0].Body; !strings.Contains(body, "MinLen>12<") {
		t.Fatalf("writer dropped fields: %s", body)
	}
}

func TestPasswordHistoryRoundtrip(t *testing.T) {
	s, _ := newOpsService(t, "GetPasswordHistoryConfiguration",
		`<GetPasswordHistoryConfigurationResponse><Enabled>true</Enabled><Length>5</Length></GetPasswordHistoryConfigurationResponse>`)

	cfg, err := s.GetPasswordHistoryConfiguration(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.Enabled || cfg.Length != 5 {
		t.Fatalf("mapping wrong: %+v", cfg)
	}

	s2, _ := newOpsService(t, "SetPasswordHistoryConfiguration", `<SetPasswordHistoryConfigurationResponse/>`)
	if err := s2.SetPasswordHistoryConfiguration(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
}

func TestAuthFailureWarningRoundtrip(t *testing.T) {
	s, _ := newOpsService(t, "GetAuthFailureWarningConfiguration",
		`<GetAuthFailureWarningConfigurationResponse><Enabled>true</Enabled><MonitorPeriod>30</MonitorPeriod><MaxAuthFailures>3</MaxAuthFailures></GetAuthFailureWarningConfigurationResponse>`)

	cfg, err := s.GetAuthFailureWarningConfiguration(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.Enabled || cfg.MaxAuthFailures != 3 {
		t.Fatalf("mapping wrong: %+v", cfg)
	}

	s2, _ := newOpsService(t, "SetAuthFailureWarningConfiguration", `<SetAuthFailureWarningConfigurationResponse/>`)

	if err := s2.SetAuthFailureWarningConfiguration(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
}

func TestCertificateLifecycle(t *testing.T) {
	cert := &Certificate{CertificateID: "cert-1"}

	t.Run("GetCertificates", func(t *testing.T) {
		s, _ := newOpsService(t, "GetCertificates",
			`<GetCertificatesResponse><Certificate><CertificateID>cert-1</CertificateID></Certificate></GetCertificatesResponse>`)

		certs, err := s.GetCertificates(context.Background())
		if err != nil || len(certs) != 1 || certs[0].CertificateID != "cert-1" {
			t.Fatalf("result: %+v err %v", certs, err)
		}
	})

	t.Run("GetCACertificates", func(t *testing.T) {
		s, _ := newOpsService(t, "GetCACertificates", `<GetCACertificatesResponse/>`)

		certs, err := s.GetCACertificates(context.Background())
		if err != nil || len(certs) != 0 {
			t.Fatalf("result: %+v err %v", certs, err)
		}
	})

	t.Run("LoadCertificates", func(t *testing.T) {
		s, _ := newOpsService(t, "LoadCertificates", `<LoadCertificatesResponse/>`)

		if err := s.LoadCertificates(context.Background(), []*Certificate{cert}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("LoadCACertificates", func(t *testing.T) {
		s, _ := newOpsService(t, "LoadCACertificates", `<LoadCACertificatesResponse/>`)

		if err := s.LoadCACertificates(context.Background(), []*Certificate{cert}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("CreateCertificate", func(t *testing.T) {
		s, _ := newOpsService(t, "CreateCertificate",
			`<CreateCertificateResponse><Certificate><CertificateID>created</CertificateID></Certificate></CreateCertificateResponse>`)

		created, err := s.CreateCertificate(context.Background(), "created", "CN=test", "", "")
		if err != nil || created == nil || created.CertificateID != "created" {
			t.Fatalf("result: %+v err %v", created, err)
		}
	})

	t.Run("DeleteCertificates", func(t *testing.T) {
		s, _ := newOpsService(t, "DeleteCertificates", `<DeleteCertificatesResponse/>`)

		if err := s.DeleteCertificates(context.Background(), []string{"cert-1"}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetCertificateInformation", func(t *testing.T) {
		s, _ := newOpsService(t, "GetCertificateInformation",
			`<GetCertificateInformationResponse><CertificateInformation><CertificateID>cert-1</CertificateID><IssuerDN>CN=ca</IssuerDN></CertificateInformation></GetCertificateInformationResponse>`)

		info, err := s.GetCertificateInformation(context.Background(), "cert-1")
		if err != nil || info == nil || info.IssuerDN != "CN=ca" {
			t.Fatalf("result: %+v err %v", info, err)
		}
	})

	t.Run("GetCertificatesStatus", func(t *testing.T) {
		s, _ := newOpsService(t, "GetCertificatesStatus",
			`<GetCertificatesStatusResponse><CertificateStatus><CertificateID>cert-1</CertificateID><Status>true</Status></CertificateStatus></GetCertificatesStatusResponse>`)

		statuses, err := s.GetCertificatesStatus(context.Background())
		if err != nil || len(statuses) != 1 || statuses[0].CertificateID != "cert-1" {
			t.Fatalf("result: %+v err %v", statuses, err)
		}
	})

	t.Run("SetCertificatesStatus", func(t *testing.T) {
		s, _ := newOpsService(t, "SetCertificatesStatus", `<SetCertificatesStatusResponse/>`)

		if err := s.SetCertificatesStatus(context.Background(), []*CertificateStatus{{CertificateID: "cert-1", Status: true}}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("GetPkcs10Request", func(t *testing.T) {
		s, _ := newOpsService(t, "GetPkcs10Request",
			`<GetPkcs10RequestResponse><Pkcs10Request><ContentType>application/pkcs10</ContentType><Data>QUJD</Data></Pkcs10Request></GetPkcs10RequestResponse>`)

		data, err := s.GetPkcs10Request(context.Background(), "cert-1", "CN=test", nil)
		if err != nil || data == nil || data.ContentType != "application/pkcs10" {
			t.Fatalf("result: %+v err %v", data, err)
		}
	})

	t.Run("LoadCertificateWithPrivateKey", func(t *testing.T) {
		s, _ := newOpsService(t, "LoadCertificateWithPrivateKey", `<LoadCertificateWithPrivateKeyResponse/>`)

		if err := s.LoadCertificateWithPrivateKey(context.Background(), []*Certificate{cert}, nil, []string{"cert-1"}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestClientCertificateMode(t *testing.T) {
	s, _ := newOpsService(t, "GetClientCertificateMode",
		`<GetClientCertificateModeResponse><Enabled>true</Enabled></GetClientCertificateModeResponse>`)

	enabled, err := s.GetClientCertificateMode(context.Background())
	if err != nil || !enabled {
		t.Fatalf("enabled = %v err %v", enabled, err)
	}

	s2, _ := newOpsService(t, "SetClientCertificateMode", `<SetClientCertificateModeResponse/>`)

	if err := s2.SetClientCertificateMode(context.Background(), true); err != nil {
		t.Fatal(err)
	}
}
