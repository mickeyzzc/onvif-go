package onvif

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetDeviceInformation(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "successful device information retrieval",
			handler: func(w http.ResponseWriter, r *http.Request) {
				response := `<?xml version="1.0" encoding="UTF-8"?>
				<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
					<s:Body>
						<tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
							<tds:Manufacturer>Test Manufacturer</tds:Manufacturer>
							<tds:Model>Test Model</tds:Model>
							<tds:FirmwareVersion>1.0.0</tds:FirmwareVersion>
							<tds:SerialNumber>12345</tds:SerialNumber>
							<tds:HardwareId>HW-001</tds:HardwareId>
						</tds:GetDeviceInformationResponse>
					</s:Body>
				</s:Envelope>`
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(response))
			},
			wantErr: false,
		},
		{
			name: "SOAP fault response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				response := `<?xml version="1.0" encoding="UTF-8"?>
				<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
					<s:Body>
						<s:Fault>
							<s:Code><s:Value>s:Receiver</s:Value></s:Code>
							<s:Reason><s:Text xml:lang="en">Internal error</s:Text></s:Reason>
						</s:Fault>
					</s:Body>
				</s:Envelope>`
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(response))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client, err := NewClient(server.URL)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			deviceInfo, err := client.Device().GetDeviceInformation(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDeviceInformation() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !tt.wantErr && deviceInfo == nil {
				t.Error("Expected device information, got nil")
			}

			if !tt.wantErr && deviceInfo != nil {
				if deviceInfo.Manufacturer != "Test Manufacturer" {
					t.Errorf("Expected manufacturer 'Test Manufacturer', got '%s'", deviceInfo.Manufacturer)
				}
			}
		})
	}
}

func TestGetCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:Capabilities>
						<tds:Device>
							<tds:XAddr>http://example.com/onvif/device_service</tds:XAddr>
						</tds:Device>
						<tds:Media>
							<tds:XAddr>http://example.com/onvif/media_service</tds:XAddr>
						</tds:Media>
					</tds:Capabilities>
				</tds:GetCapabilitiesResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	capabilities, err := client.Device().GetCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetCapabilities() error = %v", err)
	}

	if capabilities == nil {
		t.Fatal("Expected capabilities, got nil")
	}

	if capabilities.Device == nil || capabilities.Device.XAddr == "" {
		t.Error("Expected Device capabilities with XAddr")
	}
}

func TestGetHostname(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetHostnameResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:HostnameInformation>
						<tt:FromDHCP>false</tt:FromDHCP>
						<tt:Name>test-camera</tt:Name>
					</tds:HostnameInformation>
				</tds:GetHostnameResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	hostname, err := client.Device().GetHostname(context.Background())
	if err != nil {
		t.Fatalf("GetHostname() error = %v", err)
	}

	if hostname == nil {
		t.Fatal("Expected hostname information, got nil")
	}

	if hostname.Name != "test-camera" {
		t.Errorf("Expected hostname 'test-camera', got '%s'", hostname.Name)
	}
}

func TestSetHostname(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request body contains the new hostname
		var envelope struct {
			Body struct {
				SetHostname struct {
					XMLName xml.Name `xml:"SetHostname"`
					Name    string   `xml:"Name"`
				} `xml:"SetHostname"`
			} `xml:"Body"`
		}

		if err := xml.NewDecoder(r.Body).Decode(&envelope); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if envelope.Body.SetHostname.Name != "new-hostname" {
			t.Errorf("Expected hostname 'new-hostname', got '%s'", envelope.Body.SetHostname.Name)
		}

		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:SetHostnameResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Device().SetHostname(context.Background(), "new-hostname")
	if err != nil {
		t.Fatalf("SetHostname() error = %v", err)
	}
}

func TestGetDNS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetDNSResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:DNSInformation>
						<tt:FromDHCP>true</tt:FromDHCP>
						<tt:SearchDomain>example.com</tt:SearchDomain>
						<tt:DNSFromDHCP>
							<tt:Type>IPv4</tt:Type>
							<tt:IPv4Address>8.8.8.8</tt:IPv4Address>
						</tt:DNSFromDHCP>
					</tds:DNSInformation>
				</tds:GetDNSResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	dns, err := client.Device().GetDNS(context.Background())
	if err != nil {
		t.Fatalf("GetDNS() error = %v", err)
	}

	if dns == nil {
		t.Fatal("Expected DNS information, got nil")
	}

	if !dns.FromDHCP {
		t.Error("Expected DNS from DHCP")
	}
}

func TestGetUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetUsersResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:User>
						<tt:Username>admin</tt:Username>
						<tt:UserLevel>Administrator</tt:UserLevel>
					</tds:User>
					<tds:User>
						<tt:Username>user</tt:Username>
						<tt:UserLevel>User</tt:UserLevel>
					</tds:User>
				</tds:GetUsersResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	users, err := client.Device().GetUsers(context.Background())
	if err != nil {
		t.Fatalf("GetUsers() error = %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	if users[0].Username != "admin" {
		t.Errorf("Expected first user to be 'admin', got '%s'", users[0].Username)
	}
}

func TestCreateUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:CreateUsersResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	users := []*User{
		{
			Username:  "newuser",
			Password:  "password123",
			UserLevel: "User",
		},
	}

	err = client.Device().CreateUsers(context.Background(), users)
	if err != nil {
		t.Fatalf("CreateUsers() error = %v", err)
	}
}

func TestDeleteUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:DeleteUsersResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Device().DeleteUsers(context.Background(), []string{"testuser"})
	if err != nil {
		t.Fatalf("DeleteUsers() error = %v", err)
	}
}

func TestGetNetworkInterfaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetNetworkInterfacesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:NetworkInterfaces token="eth0">
						<tt:Enabled>true</tt:Enabled>
						<tt:Info>
							<tt:Name>eth0</tt:Name>
							<tt:HwAddress>00:11:22:33:44:55</tt:HwAddress>
							<tt:MTU>1500</tt:MTU>
						</tt:Info>
						<tt:IPv4>
							<tt:Enabled>true</tt:Enabled>
							<tt:Config>
								<tt:DHCP>false</tt:DHCP>
								<tt:Manual>
									<tt:Address>192.168.1.100</tt:Address>
									<tt:PrefixLength>24</tt:PrefixLength>
								</tt:Manual>
							</tt:Config>
						</tt:IPv4>
					</tds:NetworkInterfaces>
				</tds:GetNetworkInterfacesResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	interfaces, err := client.Device().GetNetworkInterfaces(context.Background())
	if err != nil {
		t.Fatalf("GetNetworkInterfaces() error = %v", err)
	}

	if len(interfaces) != 1 {
		t.Errorf("Expected 1 interface, got %d", len(interfaces))
	}

	if interfaces[0].Info.Name != "eth0" {
		t.Errorf("Expected interface name 'eth0', got '%s'", interfaces[0].Info.Name)
	}
}

func TestGetServices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetServicesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:Service>
						<tds:Namespace>http://www.onvif.org/ver10/device/wsdl</tds:Namespace>
						<tds:XAddr>http://192.168.1.100/onvif/device_service</tds:XAddr>
						<tds:Version>
							<tt:Major>2</tt:Major>
							<tt:Minor>6</tt:Minor>
						</tds:Version>
					</tds:Service>
				</tds:GetServicesResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	services, err := client.Device().GetServices(context.Background(), true)
	if err != nil {
		t.Fatalf("GetServices() error = %v", err)
	}

	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}

	if services[0].Namespace != "http://www.onvif.org/ver10/device/wsdl" {
		t.Errorf("Expected device namespace, got %s", services[0].Namespace)
	}
}

func TestGetServiceCapabilities(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetServiceCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:Capabilities>
						<tds:Network IPFilter="true" ZeroConfiguration="true"/>
						<tds:Security TLS1.2="true"/>
						<tds:System FirmwareUpgrade="true"/>
					</tds:Capabilities>
				</tds:GetServiceCapabilitiesResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	caps, err := client.Device().GetServiceCapabilities(context.Background())
	if err != nil {
		t.Fatalf("GetServiceCapabilities() error = %v", err)
	}

	if caps.Network == nil || !caps.Network.IPFilter {
		t.Error("Expected Network.IPFilter to be true")
	}
}

func TestGetDiscoveryMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetDiscoveryModeResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:DiscoveryMode>Discoverable</tds:DiscoveryMode>
				</tds:GetDiscoveryModeResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	mode, err := client.Device().GetDiscoveryMode(context.Background())
	if err != nil {
		t.Fatalf("GetDiscoveryMode() error = %v", err)
	}

	if mode != DiscoveryModeDiscoverable {
		t.Errorf("Expected Discoverable mode, got %s", mode)
	}
}

func TestSetDiscoveryMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:SetDiscoveryModeResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Device().SetDiscoveryMode(context.Background(), DiscoveryModeDiscoverable)
	if err != nil {
		t.Fatalf("SetDiscoveryMode() error = %v", err)
	}
}

func TestGetEndpointReference(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetEndpointReferenceResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:GUID>urn:uuid:12345678-1234-1234-1234-123456789abc</tds:GUID>
				</tds:GetEndpointReferenceResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	guid, err := client.Device().GetEndpointReference(context.Background())
	if err != nil {
		t.Fatalf("GetEndpointReference() error = %v", err)
	}

	expected := "urn:uuid:12345678-1234-1234-1234-123456789abc"
	if guid != expected {
		t.Errorf("Expected GUID %s, got %s", expected, guid)
	}
}

func TestGetNetworkProtocols(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetNetworkProtocolsResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:NetworkProtocols>
						<tt:Name>HTTP</tt:Name>
						<tt:Enabled>true</tt:Enabled>
						<tt:Port>80</tt:Port>
					</tds:NetworkProtocols>
					<tds:NetworkProtocols>
						<tt:Name>RTSP</tt:Name>
						<tt:Enabled>true</tt:Enabled>
						<tt:Port>554</tt:Port>
					</tds:NetworkProtocols>
				</tds:GetNetworkProtocolsResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	protocols, err := client.Device().GetNetworkProtocols(context.Background())
	if err != nil {
		t.Fatalf("GetNetworkProtocols() error = %v", err)
	}

	if len(protocols) != 2 {
		t.Fatalf("Expected 2 protocols, got %d", len(protocols))
	}

	if protocols[0].Name != NetworkProtocolHTTP {
		t.Errorf("Expected HTTP protocol, got %s", protocols[0].Name)
	}
}

func TestSetNetworkProtocols(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:SetNetworkProtocolsResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	protocols := []*NetworkProtocol{
		{Name: NetworkProtocolHTTP, Enabled: true, Port: []int{8080}},
	}

	err = client.Device().SetNetworkProtocols(context.Background(), protocols)
	if err != nil {
		t.Fatalf("SetNetworkProtocols() error = %v", err)
	}
}

func TestGetNetworkDefaultGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetNetworkDefaultGatewayResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:NetworkGateway>
						<tt:IPv4Address>192.168.1.1</tt:IPv4Address>
					</tds:NetworkGateway>
				</tds:GetNetworkDefaultGatewayResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	gateway, err := client.Device().GetNetworkDefaultGateway(context.Background())
	if err != nil {
		t.Fatalf("GetNetworkDefaultGateway() error = %v", err)
	}

	if len(gateway.IPv4Address) != 1 || gateway.IPv4Address[0] != "192.168.1.1" {
		t.Errorf("Expected gateway 192.168.1.1, got %v", gateway.IPv4Address)
	}
}

func TestSetNetworkDefaultGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:SetNetworkDefaultGatewayResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	gateway := &NetworkGateway{
		IPv4Address: []string{"192.168.1.1"},
	}

	err = client.Device().SetNetworkDefaultGateway(context.Background(), gateway)
	if err != nil {
		t.Fatalf("SetNetworkDefaultGateway() error = %v", err)
	}
}

func BenchmarkDeviceGetDeviceInformation(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `<?xml version="1.0" encoding="UTF-8"?>
		<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
			<s:Body>
				<tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
					<tds:Manufacturer>Test</tds:Manufacturer>
					<tds:Model>Model</tds:Model>
					<tds:FirmwareVersion>1.0</tds:FirmwareVersion>
					<tds:SerialNumber>123</tds:SerialNumber>
					<tds:HardwareId>HW1</tds:HardwareId>
				</tds:GetDeviceInformationResponse>
			</s:Body>
		</s:Envelope>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		_, _ = client.Device().GetDeviceInformation(ctx)
	}
}

func newMockDeviceAdditionalServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := xml.NewDecoder(r.Body)
		var envelope struct {
			Body struct {
				Content []byte `xml:",innerxml"`
			} `xml:"Body"`
		}
		_ = decoder.Decode(&envelope)
		bodyContent := string(envelope.Body.Content)

		w.Header().Set("Content-Type", "application/soap+xml")

		switch {
		case strings.Contains(bodyContent, "GetGeoLocation"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:tt="http://www.onvif.org/ver10/schema">
	<s:Body>
		<tds:GetGeoLocationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:Location Lon="-122.4194" Lat="37.7749" Elevation="10.5">
				<tt:Entity>Building A</tt:Entity>
				<tt:Token>location1</tt:Token>
				<tt:Fixed>true</tt:Fixed>
			</tds:Location>
		</tds:GetGeoLocationResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetGeoLocation"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetGeoLocationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "DeleteGeoLocation"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:DeleteGeoLocationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetDPAddresses"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetDPAddressesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:DPAddress>
				<tt:Type>IPv4</tt:Type>
				<tt:IPv4Address>239.255.255.250</tt:IPv4Address>
			</tds:DPAddress>
			<tds:DPAddress>
				<tt:Type>IPv6</tt:Type>
				<tt:IPv6Address>ff02::c</tt:IPv6Address>
			</tds:DPAddress>
		</tds:GetDPAddressesResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetDPAddresses"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetDPAddressesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetAccessPolicy"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetAccessPolicyResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:PolicyFile>
				<tt:Data>cG9saWN5IGRhdGE=</tt:Data>
				<tt:ContentType>application/xml</tt:ContentType>
			</tds:PolicyFile>
		</tds:GetAccessPolicyResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetAccessPolicy"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetAccessPolicyResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetWsdlUrl"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetWsdlUrlResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:WsdlUrl>http://192.168.1.100/onvif/device.wsdl</tds:WsdlUrl>
		</tds:GetWsdlUrlResponse>
	</s:Body>
</s:Envelope>`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestGetGeoLocation(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	locations, err := client.Device().GetGeoLocation(ctx)
	if err != nil {
		t.Fatalf("GetGeoLocation failed: %v", err)
	}

	if len(locations) != 1 {
		t.Fatalf("Expected 1 location, got %d", len(locations))
	}

	loc := locations[0]
	if loc.Entity != "Building A" {
		t.Errorf("Expected entity 'Building A', got %s", loc.Entity)
	}

	if loc.Token != "location1" {
		t.Errorf("Expected token 'location1', got %s", loc.Token)
	}

	if !loc.Fixed {
		t.Error("Expected Fixed to be true")
	}

	// Check coordinates (approximate comparison due to float precision)
	if loc.Lon < -122.42 || loc.Lon > -122.41 {
		t.Errorf("Expected longitude around -122.4194, got %f", loc.Lon)
	}

	if loc.Lat < 37.77 || loc.Lat > 37.78 {
		t.Errorf("Expected latitude around 37.7749, got %f", loc.Lat)
	}

	if loc.Elevation < 10.0 || loc.Elevation > 11.0 {
		t.Errorf("Expected elevation around 10.5, got %f", loc.Elevation)
	}
}

func TestSetGeoLocation(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	locations := []LocationEntity{
		{
			Entity:    "Main Office",
			Token:     "loc1",
			Fixed:     true,
			Lon:       -122.4194,
			Lat:       37.7749,
			Elevation: 15.0,
		},
	}

	err = client.Device().SetGeoLocation(ctx, locations)
	if err != nil {
		t.Fatalf("SetGeoLocation failed: %v", err)
	}
}

func TestDeleteGeoLocation(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	locations := []LocationEntity{
		{Token: "location1"},
	}

	err = client.Device().DeleteGeoLocation(ctx, locations)
	if err != nil {
		t.Fatalf("DeleteGeoLocation failed: %v", err)
	}
}

func TestGetDPAddresses(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	addresses, err := client.Device().GetDPAddresses(ctx)
	if err != nil {
		t.Fatalf("GetDPAddresses failed: %v", err)
	}

	if len(addresses) != 2 {
		t.Fatalf("Expected 2 addresses, got %d", len(addresses))
	}

	// Check IPv4 address
	if addresses[0].Type != "IPv4" {
		t.Errorf("Expected Type 'IPv4', got %s", addresses[0].Type)
	}
	if addresses[0].IPv4Address != "239.255.255.250" {
		t.Errorf("Expected IPv4 address '239.255.255.250', got %s", addresses[0].IPv4Address)
	}

	// Check IPv6 address
	if addresses[1].Type != "IPv6" {
		t.Errorf("Expected Type 'IPv6', got %s", addresses[1].Type)
	}
	if addresses[1].IPv6Address != "ff02::c" {
		t.Errorf("Expected IPv6 address 'ff02::c', got %s", addresses[1].IPv6Address)
	}
}

func TestSetDPAddresses(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	addresses := []NetworkHost{
		{
			Type:        "IPv4",
			IPv4Address: "239.255.255.250",
		},
	}

	err = client.Device().SetDPAddresses(ctx, addresses)
	if err != nil {
		t.Fatalf("SetDPAddresses failed: %v", err)
	}
}

func TestGetAccessPolicy(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	policy, err := client.Device().GetAccessPolicy(ctx)
	if err != nil {
		t.Fatalf("GetAccessPolicy failed: %v", err)
	}

	if policy == nil || policy.PolicyFile == nil {
		t.Fatal("Expected policy file, got nil")
	}

	if policy.PolicyFile.ContentType != "application/xml" {
		t.Errorf("Expected content type 'application/xml', got %s", policy.PolicyFile.ContentType)
	}
}

func TestSetAccessPolicy(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	policy := &AccessPolicy{
		PolicyFile: &BinaryData{
			Data:        []byte("policy data"),
			ContentType: "application/xml",
		},
	}

	err = client.Device().SetAccessPolicy(ctx, policy)
	if err != nil {
		t.Fatalf("SetAccessPolicy failed: %v", err)
	}
}

func TestGetWsdlUrl(t *testing.T) {
	server := newMockDeviceAdditionalServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	url, err := client.Device().GetWsdlURL(ctx)
	if err != nil {
		t.Fatalf("GetWsdlURL failed: %v", err)
	}

	expected := "http://192.168.1.100/onvif/device.wsdl"
	if url != expected {
		t.Errorf("Expected URL %s, got %s", expected, url)
	}
}
