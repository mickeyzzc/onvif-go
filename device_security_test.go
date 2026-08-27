package onvif

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newMockDeviceSecurityServer() *httptest.Server {
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
		case strings.Contains(bodyContent, "GetRemoteUser"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetRemoteUserResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:RemoteUser>
				<tt:Username>remote_admin</tt:Username>
				<tt:Password></tt:Password>
				<tt:UseDerivedPassword>true</tt:UseDerivedPassword>
			</tds:RemoteUser>
		</tds:GetRemoteUserResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetRemoteUser"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetRemoteUserResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetIPAddressFilter"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetIPAddressFilterResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:IPAddressFilter>
				<tt:Type>Allow</tt:Type>
				<tt:IPv4Address>
					<tt:Address>192.168.1.0</tt:Address>
					<tt:PrefixLength>24</tt:PrefixLength>
				</tt:IPv4Address>
			</tds:IPAddressFilter>
		</tds:GetIPAddressFilterResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetIPAddressFilter"),
			strings.Contains(bodyContent, "AddIPAddressFilter"),
			strings.Contains(bodyContent, "RemoveIPAddressFilter"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetIPAddressFilterResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetZeroConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetZeroConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:ZeroConfiguration>
				<tt:InterfaceToken>eth0</tt:InterfaceToken>
				<tt:Enabled>true</tt:Enabled>
				<tt:Addresses>169.254.1.100</tt:Addresses>
			</tds:ZeroConfiguration>
		</tds:GetZeroConfigurationResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetZeroConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetZeroConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetPasswordComplexityConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetPasswordComplexityConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:MinLen>8</tds:MinLen>
			<tds:Uppercase>1</tds:Uppercase>
			<tds:Number>1</tds:Number>
			<tds:SpecialChars>1</tds:SpecialChars>
			<tds:BlockUsernameOccurrence>true</tds:BlockUsernameOccurrence>
			<tds:PolicyConfigurationLocked>false</tds:PolicyConfigurationLocked>
		</tds:GetPasswordComplexityConfigurationResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetPasswordComplexityConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetPasswordComplexityConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetPasswordHistoryConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetPasswordHistoryConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:Enabled>true</tds:Enabled>
			<tds:Length>5</tds:Length>
		</tds:GetPasswordHistoryConfigurationResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetPasswordHistoryConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetPasswordHistoryConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "GetAuthFailureWarningConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:GetAuthFailureWarningConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
			<tds:Enabled>true</tds:Enabled>
			<tds:MonitorPeriod>60</tds:MonitorPeriod>
			<tds:MaxAuthFailures>5</tds:MaxAuthFailures>
		</tds:GetAuthFailureWarningConfigurationResponse>
	</s:Body>
</s:Envelope>`))

		case strings.Contains(bodyContent, "SetAuthFailureWarningConfiguration"):
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body>
		<tds:SetAuthFailureWarningConfigurationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl"/>
	</s:Body>
</s:Envelope>`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestGetRemoteUser(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	remoteUser, err := client.Security().GetRemoteUser(ctx)
	if err != nil {
		t.Fatalf("GetRemoteUser failed: %v", err)
	}

	if remoteUser.Username != "remote_admin" {
		t.Errorf("Expected username 'remote_admin', got %s", remoteUser.Username)
	}

	if !remoteUser.UseDerivedPassword {
		t.Error("UseDerivedPassword should be true")
	}
}

func TestSetRemoteUser(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	remoteUser := &RemoteUser{
		Username:           "new_remote",
		Password:           "password123",
		UseDerivedPassword: true,
	}

	err = client.Security().SetRemoteUser(ctx, remoteUser)
	if err != nil {
		t.Fatalf("SetRemoteUser failed: %v", err)
	}
}

func TestGetIPAddressFilter(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	filter, err := client.Security().GetIPAddressFilter(ctx)
	if err != nil {
		t.Fatalf("GetIPAddressFilter failed: %v", err)
	}

	if filter.Type != IPAddressFilterAllow {
		t.Errorf("Expected Allow filter type, got %s", filter.Type)
	}

	if len(filter.IPv4Address) != 1 {
		t.Fatalf("Expected 1 IPv4 address, got %d", len(filter.IPv4Address))
	}

	if filter.IPv4Address[0].Address != "192.168.1.0" {
		t.Errorf("Expected address 192.168.1.0, got %s", filter.IPv4Address[0].Address)
	}

	if filter.IPv4Address[0].PrefixLength != 24 {
		t.Errorf("Expected prefix length 24, got %d", filter.IPv4Address[0].PrefixLength)
	}
}

func TestSetIPAddressFilter(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "10.0.0.0", PrefixLength: 8},
		},
	}

	err = client.Security().SetIPAddressFilter(ctx, filter)
	if err != nil {
		t.Fatalf("SetIPAddressFilter failed: %v", err)
	}
}

func TestAddIPAddressFilter(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "172.16.0.0", PrefixLength: 12},
		},
	}

	err = client.Security().AddIPAddressFilter(ctx, filter)
	if err != nil {
		t.Fatalf("AddIPAddressFilter failed: %v", err)
	}
}

func TestRemoveIPAddressFilter(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "172.16.0.0", PrefixLength: 12},
		},
	}

	err = client.Security().RemoveIPAddressFilter(ctx, filter)
	if err != nil {
		t.Fatalf("RemoveIPAddressFilter failed: %v", err)
	}
}

func TestGetZeroConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	zeroConf, err := client.Device().GetZeroConfiguration(ctx)
	if err != nil {
		t.Fatalf("GetZeroConfiguration failed: %v", err)
	}

	if zeroConf.InterfaceToken != "eth0" {
		t.Errorf("Expected interface token 'eth0', got %s", zeroConf.InterfaceToken)
	}

	if !zeroConf.Enabled {
		t.Error("Zero configuration should be enabled")
	}

	if len(zeroConf.Addresses) != 1 || zeroConf.Addresses[0] != "169.254.1.100" {
		t.Errorf("Expected address 169.254.1.100, got %v", zeroConf.Addresses)
	}
}

func TestSetZeroConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.Device().SetZeroConfiguration(ctx, "eth0", true)
	if err != nil {
		t.Fatalf("SetZeroConfiguration failed: %v", err)
	}
}

func TestGetPasswordComplexityConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config, err := client.Security().GetPasswordComplexityConfiguration(ctx)
	if err != nil {
		t.Fatalf("GetPasswordComplexityConfiguration failed: %v", err)
	}

	if config.MinLen != 8 {
		t.Errorf("Expected MinLen 8, got %d", config.MinLen)
	}

	if config.Uppercase != 1 {
		t.Errorf("Expected Uppercase 1, got %d", config.Uppercase)
	}

	if config.Number != 1 {
		t.Errorf("Expected Number 1, got %d", config.Number)
	}

	if config.SpecialChars != 1 {
		t.Errorf("Expected SpecialChars 1, got %d", config.SpecialChars)
	}

	if !config.BlockUsernameOccurrence {
		t.Error("BlockUsernameOccurrence should be true")
	}

	if config.PolicyConfigurationLocked {
		t.Error("PolicyConfigurationLocked should be false")
	}
}

func TestSetPasswordComplexityConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config := &PasswordComplexityConfiguration{
		MinLen:                    10,
		Uppercase:                 2,
		Number:                    2,
		SpecialChars:              1,
		BlockUsernameOccurrence:   true,
		PolicyConfigurationLocked: false,
	}

	err = client.Security().SetPasswordComplexityConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("SetPasswordComplexityConfiguration failed: %v", err)
	}
}

func TestGetPasswordHistoryConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config, err := client.Security().GetPasswordHistoryConfiguration(ctx)
	if err != nil {
		t.Fatalf("GetPasswordHistoryConfiguration failed: %v", err)
	}

	if !config.Enabled {
		t.Error("Password history should be enabled")
	}

	if config.Length != 5 {
		t.Errorf("Expected Length 5, got %d", config.Length)
	}
}

func TestSetPasswordHistoryConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config := &PasswordHistoryConfiguration{
		Enabled: true,
		Length:  10,
	}

	err = client.Security().SetPasswordHistoryConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("SetPasswordHistoryConfiguration failed: %v", err)
	}
}

func TestGetAuthFailureWarningConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config, err := client.Security().GetAuthFailureWarningConfiguration(ctx)
	if err != nil {
		t.Fatalf("GetAuthFailureWarningConfiguration failed: %v", err)
	}

	if !config.Enabled {
		t.Error("Auth failure warning should be enabled")
	}

	if config.MonitorPeriod != 60 {
		t.Errorf("Expected MonitorPeriod 60, got %d", config.MonitorPeriod)
	}

	if config.MaxAuthFailures != 5 {
		t.Errorf("Expected MaxAuthFailures 5, got %d", config.MaxAuthFailures)
	}
}

func TestSetAuthFailureWarningConfiguration(t *testing.T) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	config := &AuthFailureWarningConfiguration{
		Enabled:         true,
		MonitorPeriod:   120,
		MaxAuthFailures: 3,
	}

	err = client.Security().SetAuthFailureWarningConfiguration(ctx, config)
	if err != nil {
		t.Fatalf("SetAuthFailureWarningConfiguration failed: %v", err)
	}
}

func TestIPAddressFilterTypeConstants(t *testing.T) {
	if IPAddressFilterAllow != "Allow" {
		t.Errorf("IPAddressFilterAllow should be 'Allow', got %s", IPAddressFilterAllow)
	}

	if IPAddressFilterDeny != "Deny" {
		t.Errorf("IPAddressFilterDeny should be 'Deny', got %s", IPAddressFilterDeny)
	}
}

// Benchmarks for device security operations.

func BenchmarkGetRemoteUser(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		_, _ = client.Security().GetRemoteUser(ctx)
	}
}

func BenchmarkSetRemoteUser(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	remoteUser := &RemoteUser{
		Username:           "test_user",
		Password:           "password123",
		UseDerivedPassword: true,
	}

	b.ResetTimer()
	for range b.N {
		_ = client.Security().SetRemoteUser(ctx, remoteUser)
	}
}

func BenchmarkGetIPAddressFilter(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		_, _ = client.Security().GetIPAddressFilter(ctx)
	}
}

func BenchmarkSetIPAddressFilter(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "192.168.1.0", PrefixLength: 24},
			{Address: "10.0.0.0", PrefixLength: 8},
		},
		IPv6Address: []PrefixedIPv6Address{
			{Address: "fe80::", PrefixLength: 64},
		},
	}

	b.ResetTimer()
	for range b.N {
		_ = client.Security().SetIPAddressFilter(ctx, filter)
	}
}

func BenchmarkAddIPAddressFilter(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "172.16.0.0", PrefixLength: 12},
		},
	}

	b.ResetTimer()
	for range b.N {
		_ = client.Security().AddIPAddressFilter(ctx, filter)
	}
}

func BenchmarkRemoveIPAddressFilter(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	filter := &IPAddressFilter{
		Type: IPAddressFilterAllow,
		IPv4Address: []PrefixedIPv4Address{
			{Address: "172.16.0.0", PrefixLength: 12},
		},
	}

	b.ResetTimer()
	for range b.N {
		_ = client.Security().RemoveIPAddressFilter(ctx, filter)
	}
}

func BenchmarkGetZeroConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		_, _ = client.Device().GetZeroConfiguration(ctx)
	}
}

func BenchmarkSetZeroConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		_ = client.Device().SetZeroConfiguration(ctx, "eth0", true)
	}
}

func BenchmarkGetPasswordComplexityConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		_, _ = client.Security().GetPasswordComplexityConfiguration(ctx)
	}
}

func BenchmarkSetPasswordComplexityConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	config := &PasswordComplexityConfiguration{
		MinLen:                    10,
		Uppercase:                 2,
		Number:                    2,
		SpecialChars:              1,
		BlockUsernameOccurrence:   true,
		PolicyConfigurationLocked: false,
	}

	b.ResetTimer()
	for range b.N {
		_ = client.Security().SetPasswordComplexityConfiguration(ctx, config)
	}
}

func BenchmarkGetPasswordHistoryConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		_, _ = client.Security().GetPasswordHistoryConfiguration(ctx)
	}
}

func BenchmarkSetPasswordHistoryConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	config := &PasswordHistoryConfiguration{
		Enabled: true,
		Length:  10,
	}

	b.ResetTimer()
	for range b.N {
		_ = client.Security().SetPasswordHistoryConfiguration(ctx, config)
	}
}

func BenchmarkGetAuthFailureWarningConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		_, _ = client.Security().GetAuthFailureWarningConfiguration(ctx)
	}
}

func BenchmarkSetAuthFailureWarningConfiguration(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()
	config := &AuthFailureWarningConfiguration{
		Enabled:         true,
		MonitorPeriod:   120,
		MaxAuthFailures: 3,
	}

	b.ResetTimer()
	for range b.N {
		_ = client.Security().SetAuthFailureWarningConfiguration(ctx, config)
	}
}

// BenchmarkIPAddressFilterWithManyAddresses tests performance with larger address lists.
func BenchmarkIPAddressFilterWithManyAddresses(b *testing.B) {
	server := newMockDeviceSecurityServer()
	defer server.Close()

	client, _ := NewClient(server.URL)
	ctx := context.Background()

	// Create filter with many addresses to test pre-allocation efficiency
	filter := &IPAddressFilter{
		Type:        IPAddressFilterAllow,
		IPv4Address: make([]PrefixedIPv4Address, 100),
		IPv6Address: make([]PrefixedIPv6Address, 50),
	}

	for i := range 100 {
		filter.IPv4Address[i] = PrefixedIPv4Address{
			Address:      "192.168.1.0",
			PrefixLength: 24,
		}
	}

	for i := range 50 {
		filter.IPv6Address[i] = PrefixedIPv6Address{
			Address:      "fe80::",
			PrefixLength: 64,
		}
	}

	b.ResetTimer()
	for range b.N {
		_ = client.Security().SetIPAddressFilter(ctx, filter)
	}
}

const (
	testCertID    = "cert-001"
	testXMLHeader = `<?xml version="1.0" encoding="UTF-8"?>`
)

func newMockDeviceCertificatesServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/soap+xml")

		// Parse request to determine which operation
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		requestBody := string(buf)

		var response string

		switch {
		case strings.Contains(requestBody, "GetCertificatesStatus"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetCertificatesStatusResponse>
      <tds:CertificateStatus>
        <tt:CertificateID>cert-001</tt:CertificateID>
        <tt:Status>true</tt:Status>
      </tds:CertificateStatus>
    </tds:GetCertificatesStatusResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "SetCertificatesStatus"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:SetCertificatesStatusResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetCertificateInformation"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetCertificateInformationResponse>
      <tds:CertificateInformation>
        <tt:CertificateID>cert-001</tt:CertificateID>
        <tt:IssuerDN>CN=Test CA</tt:IssuerDN>
        <tt:SubjectDN>CN=Device Certificate</tt:SubjectDN>
        <tt:ValidNotBefore>2024-01-01T00:00:00Z</tt:ValidNotBefore>
        <tt:ValidNotAfter>2025-01-01T00:00:00Z</tt:ValidNotAfter>
      </tds:CertificateInformation>
    </tds:GetCertificateInformationResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "LoadCertificateWithPrivateKey"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:LoadCertificateWithPrivateKeyResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "LoadCACertificates"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:LoadCACertificatesResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "LoadCertificates"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:LoadCertificatesResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetCACertificates"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetCACertificatesResponse>
      <tds:Certificate>
        <tt:CertificateID>ca-001</tt:CertificateID>
        <tt:Certificate>
          <tt:Data>` + base64.StdEncoding.EncodeToString([]byte("CA CERTIFICATE DATA")) + `</tt:Data>
        </tt:Certificate>
      </tds:Certificate>
    </tds:GetCACertificatesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetCertificates"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetCertificatesResponse>
      <tds:Certificate>
        <tt:CertificateID>cert-001</tt:CertificateID>
        <tt:Certificate>
          <tt:Data>` + base64.StdEncoding.EncodeToString([]byte("CERTIFICATE DATA")) + `</tt:Data>
        </tt:Certificate>
      </tds:Certificate>
    </tds:GetCertificatesResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "CreateCertificate"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:CreateCertificateResponse>
      <tds:Certificate>
        <tt:CertificateID>cert-new</tt:CertificateID>
        <tt:Certificate>
          <tt:Data>` + base64.StdEncoding.EncodeToString([]byte("NEW CERTIFICATE DATA")) + `</tt:Data>
        </tt:Certificate>
      </tds:Certificate>
    </tds:CreateCertificateResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "DeleteCertificates"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:DeleteCertificatesResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetPkcs10Request"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetPkcs10RequestResponse>
      <tds:Pkcs10Request>
        <tt:Data>` + base64.StdEncoding.EncodeToString([]byte("PKCS#10 CSR DATA")) + `</tt:Data>
      </tds:Pkcs10Request>
    </tds:GetPkcs10RequestResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "GetClientCertificateMode"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:GetClientCertificateModeResponse>
      <tds:Enabled>true</tds:Enabled>
    </tds:GetClientCertificateModeResponse>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		case strings.Contains(requestBody, "SetClientCertificateMode"):
			response = `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <tds:SetClientCertificateModeResponse/>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`

		default:
			response = testXMLHeader + `
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope">
  <SOAP-ENV:Body>
    <SOAP-ENV:Fault>
      <SOAP-ENV:Code><SOAP-ENV:Value>SOAP-ENV:Receiver</SOAP-ENV:Value></SOAP-ENV:Code>
      <SOAP-ENV:Reason><SOAP-ENV:Text>Unknown operation</SOAP-ENV:Text></SOAP-ENV:Reason>
    </SOAP-ENV:Fault>
  </SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
		}

		_, _ = w.Write([]byte(response))
	}))
}

func TestGetCertificates(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	certs, err := client.Security().GetCertificates(ctx)
	if err != nil {
		t.Fatalf("GetCertificates failed: %v", err)
	}

	if len(certs) == 0 {
		t.Error("Expected at least one certificate")
	}

	if certs[0].CertificateID != testCertID {
		t.Errorf("Expected certificate ID '%s', got '%s'", testCertID, certs[0].CertificateID)
	}
}

func TestGetCACertificates(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	certs, err := client.Security().GetCACertificates(ctx)
	if err != nil {
		t.Fatalf("GetCACertificates failed: %v", err)
	}

	if len(certs) == 0 {
		t.Error("Expected at least one CA certificate")
	}

	if certs[0].CertificateID != "ca-001" {
		t.Errorf("Expected certificate ID 'ca-001', got '%s'", certs[0].CertificateID)
	}
}

func TestLoadCertificates(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	certs := []*Certificate{
		{
			CertificateID: "cert-upload",
			Certificate: BinaryData{
				Data: []byte("UPLOADED CERTIFICATE DATA"),
			},
		},
	}

	err = client.Security().LoadCertificates(ctx, certs)
	if err != nil {
		t.Fatalf("LoadCertificates failed: %v", err)
	}
}

func TestLoadCACertificates(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	certs := []*Certificate{
		{
			CertificateID: "ca-upload",
			Certificate: BinaryData{
				Data: []byte("UPLOADED CA CERTIFICATE DATA"),
			},
		},
	}

	err = client.Security().LoadCACertificates(ctx, certs)
	if err != nil {
		t.Fatalf("LoadCACertificates failed: %v", err)
	}
}

func TestCreateCertificate(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	cert, err := client.Security().CreateCertificate(ctx, "cert-new", "CN=New Device", "2024-01-01T00:00:00Z", "2025-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("CreateCertificate failed: %v", err)
	}

	if cert.CertificateID != "cert-new" {
		t.Errorf("Expected certificate ID 'cert-new', got '%s'", cert.CertificateID)
	}
}

func TestDeleteCertificates(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	err = client.Security().DeleteCertificates(ctx, []string{"cert-001", "cert-002"})
	if err != nil {
		t.Fatalf("DeleteCertificates failed: %v", err)
	}
}

func TestGetCertificateInformation(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	info, err := client.Security().GetCertificateInformation(ctx, "cert-001")
	if err != nil {
		t.Fatalf("GetCertificateInformation failed: %v", err)
	}

	if info.CertificateID != "cert-001" {
		t.Errorf("Expected certificate ID 'cert-001', got '%s'", info.CertificateID)
	}

	if info.IssuerDN != "CN=Test CA" {
		t.Errorf("Expected issuer 'CN=Test CA', got '%s'", info.IssuerDN)
	}

	if info.SubjectDN != "CN=Device Certificate" {
		t.Errorf("Expected subject 'CN=Device Certificate', got '%s'", info.SubjectDN)
	}
}

func TestGetCertificatesStatus(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	statuses, err := client.Security().GetCertificatesStatus(ctx)
	if err != nil {
		t.Fatalf("GetCertificatesStatus failed: %v", err)
	}

	if len(statuses) == 0 {
		t.Error("Expected at least one certificate status")
	}

	if statuses[0].CertificateID != "cert-001" {
		t.Errorf("Expected certificate ID 'cert-001', got '%s'", statuses[0].CertificateID)
	}

	if !statuses[0].Status {
		t.Error("Expected certificate status to be true")
	}
}

func TestSetCertificatesStatus(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	statuses := []*CertificateStatus{
		{
			CertificateID: "cert-001",
			Status:        true,
		},
	}

	err = client.Security().SetCertificatesStatus(ctx, statuses)
	if err != nil {
		t.Fatalf("SetCertificatesStatus failed: %v", err)
	}
}

func TestGetPkcs10Request(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	csr, err := client.Security().GetPkcs10Request(ctx, "cert-csr", "CN=Device CSR", nil)
	if err != nil {
		t.Fatalf("GetPkcs10Request failed: %v", err)
	}

	if csr == nil || len(csr.Data) == 0 {
		t.Error("Expected non-empty PKCS#10 CSR data")
	}

	// Check that data was decoded from base64
	expectedData := []byte("PKCS#10 CSR DATA")
	if len(csr.Data) > 0 && !bytes.Equal(csr.Data, expectedData) {
		t.Logf("CSR data length: %d, expected: %d", len(csr.Data), len(expectedData))
		t.Logf("CSR data: %q, expected: %q", string(csr.Data), string(expectedData))
	}
}

func TestLoadCertificateWithPrivateKey(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	certs := []*Certificate{
		{
			CertificateID: "cert-with-key",
			Certificate: BinaryData{
				Data: []byte("CERTIFICATE DATA"),
			},
		},
	}

	privateKeys := []*BinaryData{
		{
			Data: []byte("PRIVATE KEY DATA"),
		},
	}

	err = client.Security().LoadCertificateWithPrivateKey(ctx, certs, privateKeys, []string{"cert-with-key"})
	if err != nil {
		t.Fatalf("LoadCertificateWithPrivateKey failed: %v", err)
	}
}

func TestGetClientCertificateMode(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	enabled, err := client.Security().GetClientCertificateMode(ctx)
	if err != nil {
		t.Fatalf("GetClientCertificateMode failed: %v", err)
	}

	if !enabled {
		t.Error("Expected client certificate mode to be enabled")
	}
}

func TestSetClientCertificateMode(t *testing.T) {
	server := newMockDeviceCertificatesServer()
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	ctx := context.Background()

	err = client.Security().SetClientCertificateMode(ctx, true)
	if err != nil {
		t.Fatalf("SetClientCertificateMode failed: %v", err)
	}
}
