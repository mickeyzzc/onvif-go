package device

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
	"github.com/mickeyzzc/onvif-go/v2/types"
)

func newDeviceCaller(handler func(action, reqXML string) (string, error)) *Service {
	return New(testutil.NewFakeCaller("http://fake/device", handler))
}

func TestSystemReboot(t *testing.T) {
	svc := newDeviceCaller(func(action, _ string) (string, error) {
		if action != "tds:SystemReboot" {
			return "", errors.New("unexpected action " + action)
		}

		return `<SystemRebootResponse><Message>Rebooting now</Message></SystemRebootResponse>`, nil
	})

	msg, err := svc.SystemReboot(context.Background())
	if err != nil {
		t.Fatalf("SystemReboot: %v", err)
	}

	if msg != "Rebooting now" {
		t.Errorf("message = %q", msg)
	}
}

func TestGetNTPParses(t *testing.T) {
	svc := newDeviceCaller(func(action, _ string) (string, error) {
		if action != "tds:GetNTP" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetNTPResponse>
	<NTPInformation><FromDHCP>false</FromDHCP>
		<NTPManual><Type>IPv4</Type><IPv4Address>10.0.0.9</IPv4Address></NTPManual>
	</NTPInformation>
</GetNTPResponse>`, nil
	})

	ntp, err := svc.GetNTP(context.Background())
	if err != nil {
		t.Fatalf("GetNTP: %v", err)
	}

	if ntp.FromDHCP {
		t.Error("FromDHCP not parsed")
	}

	if len(ntp.NTPManual) == 0 || ntp.NTPManual[0].IPv4Address != "10.0.0.9" {
		t.Errorf("NTPManual = %+v", ntp.NTPManual)
	}
}

func TestRemoteDiscoveryMode(t *testing.T) {
	svc := newDeviceCaller(func(action, reqXML string) (string, error) {
		switch action {
		case "tds:GetRemoteDiscoveryMode":
			return `<GetRemoteDiscoveryModeResponse><RemoteDiscoveryMode>Discoverable</RemoteDiscoveryMode></GetRemoteDiscoveryModeResponse>`, nil
		case "tds:SetRemoteDiscoveryMode":
			if !strings.Contains(reqXML, "NonDiscoverable") {
				t.Errorf("mode not encoded: %s", reqXML)
			}

			return `<SetRemoteDiscoveryModeResponse/>`, nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	})

	mode, err := svc.GetRemoteDiscoveryMode(context.Background())
	if err != nil {
		t.Fatalf("GetRemoteDiscoveryMode: %v", err)
	}

	if mode != DiscoveryModeDiscoverable {
		t.Errorf("mode = %q, want Discoverable", mode)
	}

	if err := svc.SetRemoteDiscoveryMode(context.Background(), DiscoveryModeNonDiscoverable); err != nil {
		t.Fatalf("SetRemoteDiscoveryMode: %v", err)
	}
}

func TestDNSAndNTPSetters(t *testing.T) {
	svc := newDeviceCaller(func(action, reqXML string) (string, error) {
		switch action {
		case "tds:SetDNS":
			for _, want := range []string{"corp.example", "10.1.1.53"} {
				if !strings.Contains(reqXML, want) {
					t.Errorf("SetDNS body misses %q: %s", want, reqXML)
				}
			}

			return `<SetDNSResponse/>`, nil
		case "tds:SetNTP":
			if !strings.Contains(reqXML, "10.0.0.9") {
				t.Errorf("SetNTP body misses server: %s", reqXML)
			}

			return `<SetNTPResponse/>`, nil
		case "tds:SetHostnameFromDHCP":
			return `<SetHostnameFromDHCPResponse><RebootNeeded>true</RebootNeeded></SetHostnameFromDHCPResponse>`, nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	})

	ctx := context.Background()

	if err := svc.SetDNS(ctx, false, []string{"corp.example"}, []types.IPAddress{{Type: "IPv4", IPv4Address: "10.1.1.53"}}); err != nil {
		t.Fatalf("SetDNS: %v", err)
	}

	if err := svc.SetNTP(ctx, false, []NetworkHost{{Type: "IPv4", IPv4Address: "10.0.0.9"}}); err != nil {
		t.Fatalf("SetNTP: %v", err)
	}

	if reboot, err := svc.SetHostnameFromDHCP(ctx, true); err != nil || !reboot {
		t.Errorf("SetHostnameFromDHCP = %v, %v, want reboot=true", reboot, err)
	}
}

func TestSystemDateTimeSetAndRead(t *testing.T) {
	svc := newDeviceCaller(func(action, reqXML string) (string, error) {
		switch action {
		case "tds:SetSystemDateAndTime":
			for _, want := range []string{"2026", "12", "30", "Manual"} {
				if !strings.Contains(reqXML, want) {
					t.Errorf("SetSystemDateAndTime misses %q: %s", want, reqXML)
				}
			}

			return `<SetSystemDateAndTimeResponse/>`, nil
		case "tds:GetSystemDateAndTime":
			return `<GetSystemDateAndTimeResponse>
	<SystemDateAndTime><DateTimeType>NTP</DateTimeType><DaylightSavings>false</DaylightSavings>
		<TimeZone><TZ>UTC</TZ></TimeZone>
		<UTCDateTime><Time><Hour>9</Hour><Minute>8</Minute><Second>7</Second></Time><Date><Year>2026</Year><Month>8</Month><Day>28</Day></Date></UTCDateTime>
	</SystemDateAndTime>
</GetSystemDateAndTimeResponse>`, nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	})

	ctx := context.Background()

	err := svc.SetSystemDateAndTime(ctx, &SystemDateTime{
		DateTimeType: SetDateTimeManual,
		UTCDateTime: &DateTime{
			Date: Date{Year: 2026, Month: 12, Day: 30},
			Time: Time{Hour: 10, Minute: 11, Second: 12},
		},
	})
	if err != nil {
		t.Fatalf("SetSystemDateAndTime: %v", err)
	}

	dt, err := svc.FixedGetSystemDateAndTime(ctx)
	if err != nil {
		t.Fatalf("FixedGetSystemDateAndTime: %v", err)
	}

	if dt.DateTimeType != "NTP" || dt.UTCDateTime.Time.Hour != 9 {
		t.Errorf("date/time not parsed: %+v", dt)
	}
}

func TestBackupAndRestoreOps(t *testing.T) {
	svc := newDeviceCaller(func(action, _ string) (string, error) {
		switch action {
		case "tds:GetSystemBackup":
			return `<GetSystemBackupResponse><BackupFiles><Name>config.tar.gz</Name><Data contentType="application/gzip">aGVsbG8=</Data></BackupFiles></GetSystemBackupResponse>`, nil
		case "tds:StartSystemRestore":
			return `<StartSystemRestoreResponse><UploadUri>http://fake/upload</UploadUri><ExpectedDownTime>PT1M</ExpectedDownTime></StartSystemRestoreResponse>`, nil
		case "tds:RestoreSystem":
			return `<RestoreSystemResponse/>`, nil
		case "tds:GetSystemUris":
			return `<GetSystemUrisResponse>
	<SystemLogUris><SystemLog><Type>System</Type><Uri>http://fake/log</Uri></SystemLog></SystemLogUris>
	<SupportInfoUri>http://fake/support</SupportInfoUri>
	<SystemBackupUri>http://fake/backup</SystemBackupUri>
</GetSystemUrisResponse>`, nil
		case "tds:GetSystemSupportInformation":
			return `<GetSystemSupportInformationResponse><SupportInformation><String>blob</String></SupportInformation></GetSystemSupportInformationResponse>`, nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	})

	ctx := context.Background()

	backup, err := svc.GetSystemBackup(ctx)
	if err != nil {
		t.Fatalf("GetSystemBackup: %v", err)
	}

	if len(backup) == 0 || backup[0].Name != "config.tar.gz" {
		t.Errorf("backup files = %+v", backup)
	}

	uploadURI, downtime, err := svc.StartSystemRestore(ctx)
	if err != nil || uploadURI != "http://fake/upload" || downtime != "PT1M" {
		t.Errorf("StartSystemRestore = %q, %q, %v", uploadURI, downtime, err)
	}

	if err := svc.RestoreSystem(ctx, backup); err != nil {
		t.Fatalf("RestoreSystem: %v", err)
	}

	logs, supportURI, backupURI, err := svc.GetSystemUris(ctx)
	if err != nil {
		t.Fatalf("GetSystemUris: %v", err)
	}

	if supportURI != "http://fake/support" || backupURI != "http://fake/backup" {
		t.Errorf("URIs = support %q, backup %q", supportURI, backupURI)
	}

	if logs == nil || len(logs.SystemLog) != 1 || logs.SystemLog[0].URI != "http://fake/log" {
		t.Errorf("log URIs = %+v", logs)
	}

	info, err := svc.GetSystemSupportInformation(ctx)
	if err != nil || info.String != "blob" {
		t.Errorf("GetSystemSupportInformation = %+v, %v", info, err)
	}
}

func TestDynamicDNS(t *testing.T) {
	svc := newDeviceCaller(func(action, reqXML string) (string, error) {
		switch action {
		case "tds:GetDynamicDNS":
			return `<GetDynamicDNSResponse><DynamicDNSInformation><Type>ClientUpdates</Type><Name>cam.example</Name></DynamicDNSInformation></GetDynamicDNSResponse>`, nil
		case "tds:SetDynamicDNS":
			if !strings.Contains(reqXML, "cam2.example") {
				t.Errorf("SetDynamicDNS body misses name: %s", reqXML)
			}

			return `<SetDynamicDNSResponse/>`, nil
		default:
			return "", errors.New("unexpected action " + action)
		}
	})

	ctx := context.Background()

	dns, err := svc.GetDynamicDNS(ctx)
	if err != nil {
		t.Fatalf("GetDynamicDNS: %v", err)
	}

	if dns.Type != DynamicDNSClientUpdates || dns.Name != "cam.example" {
		t.Errorf("DynamicDNS = %+v", dns)
	}

	if err := svc.SetDynamicDNS(ctx, DynamicDNSClientUpdates, "cam2.example"); err != nil {
		t.Fatalf("SetDynamicDNS: %v", err)
	}
}
