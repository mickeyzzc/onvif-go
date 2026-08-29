package device

// Coverage backfill for the device service: full GetCapabilities
// mapping, network interfaces with IPv4/manual addresses, GetServices,
// netmask conversion, and the single-flight/fallback paths of
// GetCapabilitiesCached.

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mickeyzzc/onvif-go/v2/internal/testutil"
)

func TestGetCapabilitiesFullMapping(t *testing.T) {
	svc := newDeviceCaller(func(action, _ string) (string, error) {
		if action != "tds:GetCapabilities" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetCapabilitiesResponse><Capabilities>
		<Analytics><XAddr>http://192.0.2.1/onvif/analytics</XAddr><RuleSupport>true</RuleSupport><AnalyticsModuleSupport>false</AnalyticsModuleSupport></Analytics>
		<Device>
			<XAddr>http://192.0.2.1/onvif/device_service</XAddr>
			<Network><IPFilter>true</IPFilter><ZeroConfiguration>false</ZeroConfiguration><IPVersion6>true</IPVersion6><DynDNS>false</DynDNS></Network>
			<System><DiscoveryResolve>true</DiscoveryResolve><DiscoveryBye>true</DiscoveryBye><RemoteDiscovery>false</RemoteDiscovery><SystemBackup>true</SystemBackup><SystemLogging>false</SystemLogging><FirmwareUpgrade>true</FirmwareUpgrade><SupportedVersions><Major>16</Major><Major>20</Major></SupportedVersions></System>
			<IO><InputConnectors>2</InputConnectors><RelayOutputs>1</RelayOutputs></IO>
			<Security><TLS1.1>false</TLS1.1><TLS1.2>true</TLS1.2><OnboardKeyGeneration>true</OnboardKeyGeneration><AccessPolicyConfig>false</AccessPolicyConfig><X.509Token>true</X.509Token><SAMLToken>false</SAMLToken><KerberosToken>false</KerberosToken><RELToken>false</RELToken></Security>
		</Device>
		<Events><XAddr>http://192.0.2.1/onvif/events</XAddr><WSSubscriptionPolicySupport>true</WSSubscriptionPolicySupport><WSPullPointSupport>true</WSPullPointSupport><WSPausableSubscriptionManagerInterfaceSupport>false</WSPausableSubscriptionManagerInterfaceSupport></Events>
		<Imaging><XAddr>http://192.0.2.1/onvif/imaging</XAddr></Imaging>
		<Media><XAddr>http://192.0.2.1/onvif/media</XAddr><StreamingCapabilities><RTPMulticast>false</RTPMulticast><RTP_TCP>true</RTP_TCP><RTP_RTSP_TCP>true</RTP_RTSP_TCP></StreamingCapabilities></Media>
		<PTZ><XAddr>http://192.0.2.1/onvif/ptz</XAddr></PTZ>
		</Capabilities></GetCapabilitiesResponse>`, nil
	})

	caps, err := svc.GetCapabilities(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if caps.Analytics == nil || !caps.Analytics.RuleSupport || caps.Analytics.AnalyticsModuleSupport {
		t.Errorf("analytics = %+v", caps.Analytics)
	}

	dev := caps.Device
	if dev == nil || dev.XAddr == "http://192.0.2.1/onvif/device_service" {
		if dev == nil {
			t.Fatal("device capabilities missing")
		}
	}

	if !dev.Network.IPFilter || !dev.Network.IPVersion6 || dev.Network.ZeroConfiguration {
		t.Errorf("network caps = %+v", dev.Network)
	}

	if len(dev.System.SupportedVersions) != 2 || !dev.System.FirmwareUpgrade {
		t.Errorf("system caps = %+v", dev.System)
	}

	if dev.IO == nil || dev.IO.InputConnectors != 2 || dev.IO.RelayOutputs != 1 {
		t.Errorf("io caps = %+v", dev.IO)
	}

	if !dev.Security.TLS12 || dev.Security.TLS11 || !dev.Security.X509Token {
		t.Errorf("security caps = %+v", dev.Security)
	}

	if !caps.Events.WSPullPointSupport || caps.Events.WSPausableSubscriptionSupport {
		t.Errorf("events caps = %+v", caps.Events)
	}

	if caps.Imaging == nil || caps.PTZ == nil {
		t.Errorf("imaging/ptz caps missing: %+v %+v", caps.Imaging, caps.PTZ)
	}

	if !caps.Media.StreamingCapabilities.RTPTCP || caps.Media.StreamingCapabilities.RTPMulticast {
		t.Errorf("media streaming caps = %+v", caps.Media.StreamingCapabilities)
	}
}

func TestGetNetworkInterfacesParses(t *testing.T) {
	svc := newDeviceCaller(func(action, _ string) (string, error) {
		if action != "tds:GetNetworkInterfaces" {
			return "", errors.New("unexpected action " + action)
		}

		return `<GetNetworkInterfacesResponse>
		<NetworkInterfaces token="eth0">
			<Enabled>true</Enabled>
			<Info><Name>eth0</Name><HwAddress>00:11:22:33:44:55</HwAddress><MTU>1500</MTU></Info>
			<IPv4><Enabled>true</Enabled><Config>
				<Manual><Address>192.168.1.42</Address><PrefixLength>24</PrefixLength></Manual>
				<DHCP>false</DHCP>
			</Config></IPv4>
		</NetworkInterfaces>
		<NetworkInterfaces token="eth1"><Enabled>false</Enabled></NetworkInterfaces>
		</GetNetworkInterfacesResponse>`, nil
	})

	ifaces, err := svc.GetNetworkInterfaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(ifaces) != 2 {
		t.Fatalf("interfaces = %d, want 2", len(ifaces))
	}

	eth0 := ifaces[0]
	if eth0.Token != "eth0" || !eth0.Enabled || eth0.Info.MTU != 1500 || eth0.Info.HwAddress != "00:11:22:33:44:55" {
		t.Errorf("eth0 = %+v", eth0)
	}

	if eth0.IPv4 == nil || len(eth0.IPv4.Config.Manual) != 1 {
		t.Fatalf("eth0 IPv4 = %+v", eth0.IPv4)
	}

	manual := eth0.IPv4.Config.Manual[0]
	if manual.Address != "192.168.1.42" || manual.PrefixLength != 24 || manual.Netmask != "255.255.255.0" {
		t.Errorf("manual entry = %+v (netmask derived from prefix)", manual)
	}

	if ifaces[1].IPv4 != nil {
		t.Errorf("disabled interface got IPv4 config: %+v", ifaces[1])
	}
}

func TestGetServicesParses(t *testing.T) {
	var gotReq string

	svc := newDeviceCaller(func(action, reqXML string) (string, error) {
		if action != "tds:GetServices" {
			return "", errors.New("unexpected action " + action)
		}

		gotReq = reqXML

		return `<GetServicesResponse>
		<Service><Namespace>http://www.onvif.org/ver10/device/wsdl</Namespace><XAddr>http://192.0.2.1/onvif/device_service</XAddr><Version><Major>2</Major><Minor>0</Minor></Version></Service>
		<Service><Namespace>http://www.onvif.org/ver10/media/wsdl</Namespace><XAddr>http://192.0.2.1/onvif/media</XAddr><Version><Major>2</Major><Minor>0</Minor></Version></Service>
		</GetServicesResponse>`, nil
	})

	services, err := svc.GetServices(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}

	if len(services) != 2 {
		t.Fatalf("services = %d, want 2", len(services))
	}

	if services[0].Namespace != "http://www.onvif.org/ver10/device/wsdl" || services[0].XAddr == "" {
		t.Errorf("first service = %+v", services[0])
	}

	if services[1].Version.Major != 2 || services[1].Version.Minor != 0 {
		t.Errorf("version mapping = %+v", services[1].Version)
	}

	if !strings.Contains(gotReq, "<tds:IncludeCapability>true</tds:IncludeCapability>") {
		t.Errorf("IncludeCapability not serialized: %s", gotReq)
	}
}

func TestGetCapabilitiesCachedSingleFlight(t *testing.T) {
	release := make(chan struct{})

	var calls int32

	var mu sync.Mutex

	caller := testutil.NewFakeCaller("http://fake/device", func(action, _ string) (string, error) {
		if action != "tds:GetCapabilities" {
			return "", errors.New("unexpected action " + action)
		}

		mu.Lock()
		calls++
		mu.Unlock()

		<-release // hold the first fetch so waiters pile up

		return `<GetCapabilitiesResponse><Capabilities><Device><XAddr>http://192.0.2.1/onvif/device_service</XAddr></Device></Capabilities></GetCapabilitiesResponse>`, nil
	})

	svc := New(caller)

	const waiters = 4

	var wg sync.WaitGroup

	results := make([]*Capabilities, waiters)

	errs := make([]error, waiters)

	for i := range waiters {
		wg.Add(1)

		go func() {
			defer wg.Done()

			caps, err := svc.GetCapabilitiesCached(context.Background())
			results[i], errs[i] = caps, err
		}()
	}

	// Let the waiters register before releasing the fetch.
	time.Sleep(50 * time.Millisecond)
	close(release)
	wg.Wait()

	mu.Lock()
	got := calls
	mu.Unlock()

	if got != 1 {
		t.Errorf("SOAP GetCapabilities calls = %d, want 1 (single-flight)", got)
	}

	for i := range waiters {
		if errs[i] != nil || results[i] == nil || results[i].Device == nil {
			t.Errorf("waiter %d: result = %+v err = %v", i, results[i], errs[i])
		}
	}

	// A later call must hit the cache without another SOAP round-trip.
	if _, err := svc.GetCapabilitiesCached(context.Background()); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	got = calls
	mu.Unlock()

	if got != 1 {
		t.Errorf("cached call re-fetched: calls = %d", got)
	}
}

func TestGetCapabilitiesCachedMinimalFallback(t *testing.T) {
	var calls int32

	caller := testutil.NewFakeCaller("http://fake/device", func(action, _ string) (string, error) {
		if action != "tds:GetCapabilities" {
			return "", errors.New("unexpected action " + action)
		}

		calls++

		return "", errors.New("device faults on GetCapabilities")
	})

	svc := NewWithFallback(caller, true)

	caps, err := svc.GetCapabilitiesCached(context.Background())
	if err != nil {
		t.Fatalf("fallback not applied: %v", err)
	}

	if caps == nil || caps.Device != nil {
		t.Errorf("fallback caps = %+v, want minimal (nil sub-capabilities)", caps)
	}

	// The degraded set is cached: no re-hammering on the next call.
	if _, err := svc.GetCapabilitiesCached(context.Background()); err != nil {
		t.Fatal(err)
	}

	if calls != 1 {
		t.Errorf("calls = %d, want 1 (fallback result cached)", calls)
	}

	// Without the fallback the error propagates.
	strict := NewWithFallback(caller, false)

	if _, err := strict.GetCapabilitiesCached(context.Background()); err == nil {
		t.Error("strict service swallowed GetCapabilities failure")
	}
}
