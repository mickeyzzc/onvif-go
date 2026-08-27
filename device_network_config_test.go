package onvif

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const networkInterfacesResponse = `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tds:GetNetworkInterfacesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
<tds:NetworkInterfaces token="eth0">
<tt:Enabled>true</tt:Enabled>
<tt:Info><tt:Name>eth0</tt:Name><tt:HwAddress>00:11:22:33:44:55</tt:HwAddress><tt:MTU>1500</tt:MTU></tt:Info>
<tt:IPv4><tt:Enabled>true</tt:Enabled>
<tt:Config><tt:Manual><tt:Address>192.168.1.64</tt:Address><tt:PrefixLength>24</tt:PrefixLength></tt:Manual>
<tt:DHCP>false</tt:DHCP></tt:Config></tt:IPv4>
</tds:NetworkInterfaces>
</tds:GetNetworkInterfacesResponse>
</s:Body></s:Envelope>`

// TestSetNetworkInterfacesRequestShape asserts the SOAP the device receives
// for DHCP and manual configurations, and the RebootNeeded round trip.
func TestSetNetworkInterfacesRequestShape(t *testing.T) {
	tests := []struct {
		name string
		cfg  NetworkInterfaceConfig
		want []string
		bad  []string
	}{
		{
			name: "DHCP enabled",
			cfg:  NetworkInterfaceConfig{Enabled: true, IPv4Enabled: true, DHCP: true},
			want: []string{
				"<tds:InterfaceToken>eth0</tds:InterfaceToken>",
				"<tt:DHCP>true</tt:DHCP>",
			},
			bad: []string{"<tt:Manual>"},
		},
		{
			name: "manual IPv4",
			cfg: NetworkInterfaceConfig{
				Enabled: true, IPv4Enabled: true, DHCP: false,
				ManualAddress: "192.168.1.50", ManualPrefixLength: 24,
			},
			want: []string{
				"<tt:Address>192.168.1.50</tt:Address>",
				"<tt:PrefixLength>24</tt:PrefixLength>",
				"<tt:DHCP>false</tt:DHCP>",
			},
			bad: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lastRequest string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				lastRequest = string(body)
				w.Header().Set("Content-Type", "application/soap+xml")
				_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope"><s:Body>
<tds:SetNetworkInterfacesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
<tds:RebootNeeded>true</tds:RebootNeeded>
</tds:SetNetworkInterfacesResponse>
</s:Body></s:Envelope>`))
			}))
			t.Cleanup(server.Close)

			client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			reboot, err := client.Device().SetNetworkInterfaces(
				context.Background(), "eth0", tt.cfg)
			if err != nil {
				t.Fatalf("SetNetworkInterfaces() error = %v", err)
			}

			if !reboot {
				t.Error("RebootNeeded = false, want true")
			}

			for _, want := range tt.want {
				if !strings.Contains(lastRequest, want) {
					t.Errorf("request missing %q: %s", want, lastRequest)
				}
			}

			for _, bad := range tt.bad {
				if strings.Contains(lastRequest, bad) {
					t.Errorf("request must not contain %q: %s", bad, lastRequest)
				}
			}
		})
	}
}

func TestSetNetworkInterfacesValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("device must not be contacted for invalid configurations")
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if _, err := client.Device().SetNetworkInterfaces(context.Background(), "", NetworkInterfaceConfig{}); err == nil {
		t.Error("empty token accepted, want error")
	}

	if _, err := client.Device().SetNetworkInterfaces(context.Background(), "eth0",
		NetworkInterfaceConfig{DHCP: false, ManualAddress: ""}); err == nil {
		t.Error("static config without manual address accepted, want error")
	}
}

// TestGetSetNetworkInterfacesRoundTrip verifies Get→Set consistency on the
// parsed structures.
func TestGetSetNetworkInterfacesRoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")
		_, _ = w.Write([]byte(networkInterfacesResponse))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	interfaces, err := client.Device().GetNetworkInterfaces(context.Background())
	if err != nil {
		t.Fatalf("GetNetworkInterfaces() error = %v", err)
	}

	if len(interfaces) != 1 {
		t.Fatalf("got %d interfaces, want 1", len(interfaces))
	}

	eth0 := interfaces[0]
	if eth0.Token != "eth0" || !eth0.Enabled || eth0.IPv4 == nil {
		t.Fatalf("interface parsed wrong: %+v", eth0)
	}

	manual := eth0.IPv4.Config.Manual[0]
	if manual.Address != "192.168.1.64" || manual.PrefixLength != 24 {
		t.Errorf("manual address parsed wrong: %+v", manual)
	}

	if manual.Netmask != "255.255.255.0" {
		t.Errorf("Netmask = %q, want 255.255.255.0", manual.Netmask)
	}
}

func TestNetmaskFromPrefixLength(t *testing.T) {
	tests := []struct {
		prefix int
		want   string
	}{
		{0, "0.0.0.0"},
		{8, "255.0.0.0"},
		{24, "255.255.255.0"},
		{25, "255.255.255.128"},
		{32, "255.255.255.255"},
		{-1, ""},
		{33, ""},
	}

	for _, tt := range tests {
		if got := NetmaskFromPrefixLength(tt.prefix); got != tt.want {
			t.Errorf("NetmaskFromPrefixLength(%d) = %q, want %q", tt.prefix, got, tt.want)
		}
	}
}
