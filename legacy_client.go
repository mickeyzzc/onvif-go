package onvif

import (
	"context"
	"time"
)

// Transitional *Client delegators over the service facades. These exist only
// while call sites migrate to the facade API and are REMOVED before the
// v1.2.0 tag — do not use in new code.

// --- transitional delegators: MediaService ---
func (c *Client) GetProfiles(ctx context.Context) ([]*Profile, error) {
	return c.Media().GetProfiles(ctx)
}

func (c *Client) GetStreamURI(ctx context.Context, profileToken string) (*MediaURI, error) {
	return c.Media().GetStreamURI(ctx, profileToken)
}

func (c *Client) GetSnapshotURI(ctx context.Context, profileToken string) (*MediaURI, error) {
	return c.Media().GetSnapshotURI(ctx, profileToken)
}

func (c *Client) GetVideoEncoderConfiguration(ctx context.Context, configurationToken string) (*VideoEncoderConfiguration, error) {
	return c.Media().GetVideoEncoderConfiguration(ctx, configurationToken)
}

func (c *Client) GetVideoSources(ctx context.Context) ([]*VideoSource, error) {
	return c.Media().GetVideoSources(ctx)
}

func (c *Client) GetAudioSources(ctx context.Context) ([]*AudioSource, error) {
	return c.Media().GetAudioSources(ctx)
}

func (c *Client) GetAudioOutputs(ctx context.Context) ([]*AudioOutput, error) {
	return c.Media().GetAudioOutputs(ctx)
}

func (c *Client) CreateProfile(ctx context.Context, name, token string) (*Profile, error) {
	return c.Media().CreateProfile(ctx, name, token)
}

func (c *Client) DeleteProfile(ctx context.Context, profileToken string) error {
	return c.Media().DeleteProfile(ctx, profileToken)
}

func (c *Client) SetVideoEncoderConfiguration(ctx context.Context, config *VideoEncoderConfiguration, forcePersistence bool) error {
	return c.Media().SetVideoEncoderConfiguration(ctx, config, forcePersistence)
}

func (c *Client) GetMediaServiceCapabilities(ctx context.Context) (*MediaServiceCapabilities, error) {
	return c.Media().GetMediaServiceCapabilities(ctx)
}

func (c *Client) GetVideoEncoderConfigurationOptions(ctx context.Context, configurationToken string) (*VideoEncoderConfigurationOptions, error) {
	return c.Media().GetVideoEncoderConfigurationOptions(ctx, configurationToken)
}

func (c *Client) GetAudioEncoderConfiguration(ctx context.Context, configurationToken string) (*AudioEncoderConfiguration, error) {
	return c.Media().GetAudioEncoderConfiguration(ctx, configurationToken)
}

func (c *Client) SetAudioEncoderConfiguration(ctx context.Context, config *AudioEncoderConfiguration, forcePersistence bool) error {
	return c.Media().SetAudioEncoderConfiguration(ctx, config, forcePersistence)
}

func (c *Client) GetMetadataConfiguration(ctx context.Context, configurationToken string) (*MetadataConfiguration, error) {
	return c.Media().GetMetadataConfiguration(ctx, configurationToken)
}

func (c *Client) SetMetadataConfiguration(ctx context.Context, config *MetadataConfiguration, forcePersistence bool) error {
	return c.Media().SetMetadataConfiguration(ctx, config, forcePersistence)
}

func (c *Client) GetVideoSourceModes(ctx context.Context, videoSourceToken string) ([]*VideoSourceMode, error) {
	return c.Media().GetVideoSourceModes(ctx, videoSourceToken)
}

func (c *Client) SetVideoSourceMode(ctx context.Context, videoSourceToken, modeToken string) error {
	return c.Media().SetVideoSourceMode(ctx, videoSourceToken, modeToken)
}

func (c *Client) SetSynchronizationPoint(ctx context.Context, profileToken string) error {
	return c.Media().SetSynchronizationPoint(ctx, profileToken)
}

func (c *Client) GetOSDs(ctx context.Context, configurationToken string) ([]*OSDConfiguration, error) {
	return c.Media().GetOSDs(ctx, configurationToken)
}

func (c *Client) GetOSD(ctx context.Context, osdToken string) (*OSDConfiguration, error) {
	return c.Media().GetOSD(ctx, osdToken)
}

func (c *Client) SetOSD(ctx context.Context, osd *OSDConfiguration) error {
	return c.Media().SetOSD(ctx, osd)
}

func (c *Client) CreateOSD(ctx context.Context, videoSourceConfigurationToken string, osd *OSDConfiguration) (*OSDConfiguration, error) {
	return c.Media().CreateOSD(ctx, videoSourceConfigurationToken, osd)
}

func (c *Client) DeleteOSD(ctx context.Context, osdToken string) error {
	return c.Media().DeleteOSD(ctx, osdToken)
}

func (c *Client) StartMulticastStreaming(ctx context.Context, profileToken string) error {
	return c.Media().StartMulticastStreaming(ctx, profileToken)
}

func (c *Client) StopMulticastStreaming(ctx context.Context, profileToken string) error {
	return c.Media().StopMulticastStreaming(ctx, profileToken)
}

func (c *Client) GetProfile(ctx context.Context, profileToken string) (*Profile, error) {
	return c.Media().GetProfile(ctx, profileToken)
}

func (c *Client) SetProfile(ctx context.Context, profile *Profile) error {
	return c.Media().SetProfile(ctx, profile)
}

func (c *Client) AddVideoEncoderConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	return c.Media().AddVideoEncoderConfiguration(ctx, profileToken, configurationToken)
}

func (c *Client) RemoveVideoEncoderConfiguration(ctx context.Context, profileToken string) error {
	return c.Media().RemoveVideoEncoderConfiguration(ctx, profileToken)
}

func (c *Client) AddAudioEncoderConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	return c.Media().AddAudioEncoderConfiguration(ctx, profileToken, configurationToken)
}

func (c *Client) RemoveAudioEncoderConfiguration(ctx context.Context, profileToken string) error {
	return c.Media().RemoveAudioEncoderConfiguration(ctx, profileToken)
}

func (c *Client) AddAudioSourceConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	return c.Media().AddAudioSourceConfiguration(ctx, profileToken, configurationToken)
}

func (c *Client) RemoveAudioSourceConfiguration(ctx context.Context, profileToken string) error {
	return c.Media().RemoveAudioSourceConfiguration(ctx, profileToken)
}

func (c *Client) AddVideoSourceConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	return c.Media().AddVideoSourceConfiguration(ctx, profileToken, configurationToken)
}

func (c *Client) RemoveVideoSourceConfiguration(ctx context.Context, profileToken string) error {
	return c.Media().RemoveVideoSourceConfiguration(ctx, profileToken)
}

func (c *Client) AddPTZConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	return c.Media().AddPTZConfiguration(ctx, profileToken, configurationToken)
}

func (c *Client) RemovePTZConfiguration(ctx context.Context, profileToken string) error {
	return c.Media().RemovePTZConfiguration(ctx, profileToken)
}

func (c *Client) AddMetadataConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	return c.Media().AddMetadataConfiguration(ctx, profileToken, configurationToken)
}

func (c *Client) RemoveMetadataConfiguration(ctx context.Context, profileToken string) error {
	return c.Media().RemoveMetadataConfiguration(ctx, profileToken)
}

func (c *Client) GetAudioEncoderConfigurationOptions(ctx context.Context, configurationToken, profileToken string) (*AudioEncoderConfigurationOptions, error) {
	return c.Media().GetAudioEncoderConfigurationOptions(ctx, configurationToken, profileToken)
}

func (c *Client) GetMetadataConfigurationOptions(ctx context.Context, configurationToken, profileToken string) (*MetadataConfigurationOptions, error) {
	return c.Media().GetMetadataConfigurationOptions(ctx, configurationToken, profileToken)
}

func (c *Client) GetAudioOutputConfiguration(ctx context.Context, configurationToken string) (*AudioOutputConfiguration, error) {
	return c.Media().GetAudioOutputConfiguration(ctx, configurationToken)
}

func (c *Client) SetAudioOutputConfiguration(ctx context.Context, config *AudioOutputConfiguration, forcePersistence bool) error {
	return c.Media().SetAudioOutputConfiguration(ctx, config, forcePersistence)
}

func (c *Client) GetAudioOutputConfigurationOptions(ctx context.Context, configurationToken string) (*AudioOutputConfigurationOptions, error) {
	return c.Media().GetAudioOutputConfigurationOptions(ctx, configurationToken)
}

func (c *Client) GetAudioDecoderConfigurationOptions(ctx context.Context, configurationToken string) (*AudioDecoderConfigurationOptions, error) {
	return c.Media().GetAudioDecoderConfigurationOptions(ctx, configurationToken)
}

func (c *Client) GetGuaranteedNumberOfVideoEncoderInstances(ctx context.Context, configurationToken string) (*GuaranteedNumberOfVideoEncoderInstances, error) {
	return c.Media().GetGuaranteedNumberOfVideoEncoderInstances(ctx, configurationToken)
}

func (c *Client) GetOSDOptions(ctx context.Context, configurationToken string) (*OSDConfigurationOptions, error) {
	return c.Media().GetOSDOptions(ctx, configurationToken)
}

func (c *Client) GetVideoSourceConfigurations(ctx context.Context) ([]*VideoSourceConfiguration, error) {
	return c.Media().GetVideoSourceConfigurations(ctx)
}

func (c *Client) GetAudioSourceConfigurations(ctx context.Context) ([]*AudioSourceConfiguration, error) {
	return c.Media().GetAudioSourceConfigurations(ctx)
}

func (c *Client) GetVideoEncoderConfigurations(ctx context.Context) ([]*VideoEncoderConfiguration, error) {
	return c.Media().GetVideoEncoderConfigurations(ctx)
}

func (c *Client) GetAudioEncoderConfigurations(ctx context.Context) ([]*AudioEncoderConfiguration, error) {
	return c.Media().GetAudioEncoderConfigurations(ctx)
}

func (c *Client) GetVideoSourceConfiguration(ctx context.Context, configurationToken string) (*VideoSourceConfiguration, error) {
	return c.Media().GetVideoSourceConfiguration(ctx, configurationToken)
}

func (c *Client) GetAudioSourceConfiguration(ctx context.Context, configurationToken string) (*AudioSourceConfiguration, error) {
	return c.Media().GetAudioSourceConfiguration(ctx, configurationToken)
}

func (c *Client) GetVideoSourceConfigurationOptions(ctx context.Context, configurationToken, profileToken string) (*VideoSourceConfigurationOptions, error) {
	return c.Media().GetVideoSourceConfigurationOptions(ctx, configurationToken, profileToken)
}

func (c *Client) GetAudioSourceConfigurationOptions(ctx context.Context, configurationToken, profileToken string) (*AudioSourceConfigurationOptions, error) {
	return c.Media().GetAudioSourceConfigurationOptions(ctx, configurationToken, profileToken)
}

func (c *Client) SetVideoSourceConfiguration(ctx context.Context, config *VideoSourceConfiguration, forcePersistence bool) error {
	return c.Media().SetVideoSourceConfiguration(ctx, config, forcePersistence)
}

func (c *Client) SetAudioSourceConfiguration(ctx context.Context, config *AudioSourceConfiguration, forcePersistence bool) error {
	return c.Media().SetAudioSourceConfiguration(ctx, config, forcePersistence)
}

func (c *Client) GetCompatibleVideoEncoderConfigurations(ctx context.Context, profileToken string) ([]*VideoEncoderConfiguration, error) {
	return c.Media().GetCompatibleVideoEncoderConfigurations(ctx, profileToken)
}

func (c *Client) GetCompatibleVideoSourceConfigurations(ctx context.Context, profileToken string) ([]*VideoSourceConfiguration, error) {
	return c.Media().GetCompatibleVideoSourceConfigurations(ctx, profileToken)
}

func (c *Client) GetCompatibleAudioEncoderConfigurations(ctx context.Context, profileToken string) ([]*AudioEncoderConfiguration, error) {
	return c.Media().GetCompatibleAudioEncoderConfigurations(ctx, profileToken)
}

func (c *Client) GetCompatibleAudioSourceConfigurations(ctx context.Context, profileToken string) ([]*AudioSourceConfiguration, error) {
	return c.Media().GetCompatibleAudioSourceConfigurations(ctx, profileToken)
}

func (c *Client) GetCompatiblePTZConfigurations(ctx context.Context, profileToken string) ([]*PTZConfiguration, error) {
	return c.Media().GetCompatiblePTZConfigurations(ctx, profileToken)
}

func (c *Client) GetCompatibleMetadataConfigurations(ctx context.Context, profileToken string) ([]*MetadataConfiguration, error) {
	return c.Media().GetCompatibleMetadataConfigurations(ctx, profileToken)
}

func (c *Client) GetCompatibleAudioOutputConfigurations(ctx context.Context, profileToken string) ([]*AudioOutputConfiguration, error) {
	return c.Media().GetCompatibleAudioOutputConfigurations(ctx, profileToken)
}

func (c *Client) GetCompatibleAudioDecoderConfigurations(ctx context.Context, profileToken string) ([]*AudioDecoderConfiguration, error) {
	return c.Media().GetCompatibleAudioDecoderConfigurations(ctx, profileToken)
}

func (c *Client) GetMetadataConfigurations(ctx context.Context) ([]*MetadataConfiguration, error) {
	return c.Media().GetMetadataConfigurations(ctx)
}

func (c *Client) GetAudioOutputConfigurations(ctx context.Context) ([]*AudioOutputConfiguration, error) {
	return c.Media().GetAudioOutputConfigurations(ctx)
}

func (c *Client) GetAudioDecoderConfigurations(ctx context.Context) ([]*AudioDecoderConfiguration, error) {
	return c.Media().GetAudioDecoderConfigurations(ctx)
}

func (c *Client) GetAudioDecoderConfiguration(ctx context.Context, configurationToken string) (*AudioDecoderConfiguration, error) {
	return c.Media().GetAudioDecoderConfiguration(ctx, configurationToken)
}

func (c *Client) SetAudioDecoderConfiguration(ctx context.Context, config *AudioDecoderConfiguration, forcePersistence bool) error {
	return c.Media().SetAudioDecoderConfiguration(ctx, config, forcePersistence)
}

func (c *Client) GetVideoAnalyticsConfigurations(ctx context.Context) ([]*VideoAnalyticsConfiguration, error) {
	return c.Media().GetVideoAnalyticsConfigurations(ctx)
}

func (c *Client) GetVideoAnalyticsConfiguration(ctx context.Context, configurationToken string) (*VideoAnalyticsConfiguration, error) {
	return c.Media().GetVideoAnalyticsConfiguration(ctx, configurationToken)
}

func (c *Client) GetCompatibleVideoAnalyticsConfigurations(ctx context.Context, profileToken string) ([]*VideoAnalyticsConfiguration, error) {
	return c.Media().GetCompatibleVideoAnalyticsConfigurations(ctx, profileToken)
}

func (c *Client) SetVideoAnalyticsConfiguration(ctx context.Context, config *VideoAnalyticsConfiguration, forcePersistence bool) error {
	return c.Media().SetVideoAnalyticsConfiguration(ctx, config, forcePersistence)
}

func (c *Client) GetVideoAnalyticsConfigurationOptions(ctx context.Context, configurationToken, profileToken string) (*VideoAnalyticsConfigurationOptions, error) {
	return c.Media().GetVideoAnalyticsConfigurationOptions(ctx, configurationToken, profileToken)
}

func (c *Client) AddVideoAnalyticsConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	return c.Media().AddVideoAnalyticsConfiguration(ctx, profileToken, configurationToken)
}

func (c *Client) RemoveVideoAnalyticsConfiguration(ctx context.Context, profileToken string) error {
	return c.Media().RemoveVideoAnalyticsConfiguration(ctx, profileToken)
}

func (c *Client) AddAudioOutputConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	return c.Media().AddAudioOutputConfiguration(ctx, profileToken, configurationToken)
}

func (c *Client) RemoveAudioOutputConfiguration(ctx context.Context, profileToken string) error {
	return c.Media().RemoveAudioOutputConfiguration(ctx, profileToken)
}

func (c *Client) AddAudioDecoderConfiguration(ctx context.Context, profileToken, configurationToken string) error {
	return c.Media().AddAudioDecoderConfiguration(ctx, profileToken, configurationToken)
}

func (c *Client) RemoveAudioDecoderConfiguration(ctx context.Context, profileToken string) error {
	return c.Media().RemoveAudioDecoderConfiguration(ctx, profileToken)
}

// --- transitional delegators: DeviceService ---
func (c *Client) GetDeviceInformation(ctx context.Context) (*DeviceInformation, error) {
	return c.Device().GetDeviceInformation(ctx)
}

func (c *Client) GetCapabilities(ctx context.Context) (*Capabilities, error) {
	return c.Device().GetCapabilities(ctx)
}

func (c *Client) SystemReboot(ctx context.Context) (string, error) {
	return c.Device().SystemReboot(ctx)
}

func (c *Client) GetSystemDateAndTime(ctx context.Context) (interface{}, error) {
	return c.Device().GetSystemDateAndTime(ctx)
}

func (c *Client) GetHostname(ctx context.Context) (*HostnameInformation, error) {
	return c.Device().GetHostname(ctx)
}

func (c *Client) SetHostname(ctx context.Context, name string) error {
	return c.Device().SetHostname(ctx, name)
}

func (c *Client) GetDNS(ctx context.Context) (*DNSInformation, error) { return c.Device().GetDNS(ctx) }

func (c *Client) GetNTP(ctx context.Context) (*NTPInformation, error) { return c.Device().GetNTP(ctx) }

func (c *Client) GetNetworkInterfaces(ctx context.Context) ([]*NetworkInterface, error) {
	return c.Device().GetNetworkInterfaces(ctx)
}

func (c *Client) GetScopes(ctx context.Context) ([]*Scope, error) { return c.Device().GetScopes(ctx) }

func (c *Client) GetUsers(ctx context.Context) ([]*User, error) { return c.Device().GetUsers(ctx) }

func (c *Client) CreateUsers(ctx context.Context, users []*User) error {
	return c.Device().CreateUsers(ctx, users)
}

func (c *Client) DeleteUsers(ctx context.Context, usernames []string) error {
	return c.Device().DeleteUsers(ctx, usernames)
}

func (c *Client) SetUser(ctx context.Context, user *User) error { return c.Device().SetUser(ctx, user) }

func (c *Client) GetServices(ctx context.Context, includeCapability bool) ([]*Service, error) {
	return c.Device().GetServices(ctx, includeCapability)
}

func (c *Client) GetServiceCapabilities(ctx context.Context) (*DeviceServiceCapabilities, error) {
	return c.Device().GetServiceCapabilities(ctx)
}

func (c *Client) GetDiscoveryMode(ctx context.Context) (DiscoveryMode, error) {
	return c.Device().GetDiscoveryMode(ctx)
}

func (c *Client) SetDiscoveryMode(ctx context.Context, mode DiscoveryMode) error {
	return c.Device().SetDiscoveryMode(ctx, mode)
}

func (c *Client) GetRemoteDiscoveryMode(ctx context.Context) (DiscoveryMode, error) {
	return c.Device().GetRemoteDiscoveryMode(ctx)
}

func (c *Client) SetRemoteDiscoveryMode(ctx context.Context, mode DiscoveryMode) error {
	return c.Device().SetRemoteDiscoveryMode(ctx, mode)
}

func (c *Client) GetEndpointReference(ctx context.Context) (string, error) {
	return c.Device().GetEndpointReference(ctx)
}

func (c *Client) GetNetworkProtocols(ctx context.Context) ([]*NetworkProtocol, error) {
	return c.Device().GetNetworkProtocols(ctx)
}

func (c *Client) SetNetworkProtocols(ctx context.Context, protocols []*NetworkProtocol) error {
	return c.Device().SetNetworkProtocols(ctx, protocols)
}

func (c *Client) GetNetworkDefaultGateway(ctx context.Context) (*NetworkGateway, error) {
	return c.Device().GetNetworkDefaultGateway(ctx)
}

func (c *Client) SetNetworkDefaultGateway(ctx context.Context, gateway *NetworkGateway) error {
	return c.Device().SetNetworkDefaultGateway(ctx, gateway)
}

func (c *Client) GetGeoLocation(ctx context.Context) ([]LocationEntity, error) {
	return c.Device().GetGeoLocation(ctx)
}

func (c *Client) SetGeoLocation(ctx context.Context, location []LocationEntity) error {
	return c.Device().SetGeoLocation(ctx, location)
}

func (c *Client) DeleteGeoLocation(ctx context.Context, location []LocationEntity) error {
	return c.Device().DeleteGeoLocation(ctx, location)
}

func (c *Client) GetDPAddresses(ctx context.Context) ([]NetworkHost, error) {
	return c.Device().GetDPAddresses(ctx)
}

func (c *Client) SetDPAddresses(ctx context.Context, dpAddress []NetworkHost) error {
	return c.Device().SetDPAddresses(ctx, dpAddress)
}

func (c *Client) GetAccessPolicy(ctx context.Context) (*AccessPolicy, error) {
	return c.Device().GetAccessPolicy(ctx)
}

func (c *Client) SetAccessPolicy(ctx context.Context, policy *AccessPolicy) error {
	return c.Device().SetAccessPolicy(ctx, policy)
}

func (c *Client) GetWsdlURL(ctx context.Context) (string, error) { return c.Device().GetWsdlURL(ctx) }

func (c *Client) GetCertificates(ctx context.Context) ([]*Certificate, error) {
	return c.Device().GetCertificates(ctx)
}

func (c *Client) GetCACertificates(ctx context.Context) ([]*Certificate, error) {
	return c.Device().GetCACertificates(ctx)
}

func (c *Client) LoadCertificates(ctx context.Context, certificates []*Certificate) error {
	return c.Device().LoadCertificates(ctx, certificates)
}

func (c *Client) LoadCACertificates(ctx context.Context, certificates []*Certificate) error {
	return c.Device().LoadCACertificates(ctx, certificates)
}

func (c *Client) CreateCertificate(ctx context.Context, certificateID, subject, validNotBefore, validNotAfter string) (*Certificate, error) {
	return c.Device().CreateCertificate(ctx, certificateID, subject, validNotBefore, validNotAfter)
}

func (c *Client) DeleteCertificates(ctx context.Context, certificateIDs []string) error {
	return c.Device().DeleteCertificates(ctx, certificateIDs)
}

func (c *Client) GetCertificateInformation(ctx context.Context, certificateID string) (*CertificateInformation, error) {
	return c.Device().GetCertificateInformation(ctx, certificateID)
}

func (c *Client) GetCertificatesStatus(ctx context.Context) ([]*CertificateStatus, error) {
	return c.Device().GetCertificatesStatus(ctx)
}

func (c *Client) SetCertificatesStatus(ctx context.Context, statuses []*CertificateStatus) error {
	return c.Device().SetCertificatesStatus(ctx, statuses)
}

func (c *Client) GetPkcs10Request(ctx context.Context, certificateID, subject string, attributes *BinaryData) (*BinaryData, error) {
	return c.Device().GetPkcs10Request(ctx, certificateID, subject, attributes)
}

func (c *Client) LoadCertificateWithPrivateKey(ctx context.Context, certificates []*Certificate, privateKey []*BinaryData, certificateIDs []string) error {
	return c.Device().LoadCertificateWithPrivateKey(ctx, certificates, privateKey, certificateIDs)
}

func (c *Client) GetClientCertificateMode(ctx context.Context) (bool, error) {
	return c.Device().GetClientCertificateMode(ctx)
}

func (c *Client) SetClientCertificateMode(ctx context.Context, enabled bool) error {
	return c.Device().SetClientCertificateMode(ctx, enabled)
}

func (c *Client) SetDNS(ctx context.Context, fromDHCP bool, searchDomain []string, dnsManual []IPAddress) error {
	return c.Device().SetDNS(ctx, fromDHCP, searchDomain, dnsManual)
}

func (c *Client) SetNTP(ctx context.Context, fromDHCP bool, ntpManual []NetworkHost) error {
	return c.Device().SetNTP(ctx, fromDHCP, ntpManual)
}

func (c *Client) SetHostnameFromDHCP(ctx context.Context, fromDHCP bool) (bool, error) {
	return c.Device().SetHostnameFromDHCP(ctx, fromDHCP)
}

func (c *Client) FixedGetSystemDateAndTime(ctx context.Context) (*SystemDateTime, error) {
	return c.Device().FixedGetSystemDateAndTime(ctx)
}

func (c *Client) SetSystemDateAndTime(ctx context.Context, dateTime *SystemDateTime) error {
	return c.Device().SetSystemDateAndTime(ctx, dateTime)
}

func (c *Client) AddScopes(ctx context.Context, scopeItems []string) error {
	return c.Device().AddScopes(ctx, scopeItems)
}

func (c *Client) RemoveScopes(ctx context.Context, scopeItems []string) ([]string, error) {
	return c.Device().RemoveScopes(ctx, scopeItems)
}

func (c *Client) SetScopes(ctx context.Context, scopes []string) error {
	return c.Device().SetScopes(ctx, scopes)
}

func (c *Client) GetRelayOutputs(ctx context.Context) ([]*RelayOutput, error) {
	return c.Device().GetRelayOutputs(ctx)
}

func (c *Client) SetRelayOutputSettings(ctx context.Context, token string, settings *RelayOutputSettings) error {
	return c.Device().SetRelayOutputSettings(ctx, token, settings)
}

func (c *Client) SetRelayOutputState(ctx context.Context, token string, state RelayLogicalState) error {
	return c.Device().SetRelayOutputState(ctx, token, state)
}

func (c *Client) SendAuxiliaryCommand(ctx context.Context, command AuxiliaryData) (AuxiliaryData, error) {
	return c.Device().SendAuxiliaryCommand(ctx, command)
}

func (c *Client) GetSystemLog(ctx context.Context, logType SystemLogType) (*SystemLog, error) {
	return c.Device().GetSystemLog(ctx, logType)
}

func (c *Client) GetSystemBackup(ctx context.Context) ([]*BackupFile, error) {
	return c.Device().GetSystemBackup(ctx)
}

func (c *Client) RestoreSystem(ctx context.Context, backupFiles []*BackupFile) error {
	return c.Device().RestoreSystem(ctx, backupFiles)
}

func (c *Client) GetSystemUris(ctx context.Context) (uriList *SystemLogURIList, systemBackupURI, systemLogURI string, err error) {
	return c.Device().GetSystemUris(ctx)
}

func (c *Client) GetSystemSupportInformation(ctx context.Context) (*SupportInformation, error) {
	return c.Device().GetSystemSupportInformation(ctx)
}

func (c *Client) SetSystemFactoryDefault(ctx context.Context, factoryDefault FactoryDefaultType) error {
	return c.Device().SetSystemFactoryDefault(ctx, factoryDefault)
}

func (c *Client) StartFirmwareUpgrade(ctx context.Context) (uploadURI, uploadDelay, expectedDownTime string, err error) {
	return c.Device().StartFirmwareUpgrade(ctx)
}

func (c *Client) StartSystemRestore(ctx context.Context) (uploadURI, expectedDownTime string, err error) {
	return c.Device().StartSystemRestore(ctx)
}

func (c *Client) GetStorageConfigurations(ctx context.Context) ([]*StorageConfiguration, error) {
	return c.Device().GetStorageConfigurations(ctx)
}

func (c *Client) GetStorageConfiguration(ctx context.Context, token string) (*StorageConfiguration, error) {
	return c.Device().GetStorageConfiguration(ctx, token)
}

func (c *Client) CreateStorageConfiguration(ctx context.Context, config *StorageConfiguration) (string, error) {
	return c.Device().CreateStorageConfiguration(ctx, config)
}

func (c *Client) SetStorageConfiguration(ctx context.Context, config *StorageConfiguration) error {
	return c.Device().SetStorageConfiguration(ctx, config)
}

func (c *Client) DeleteStorageConfiguration(ctx context.Context, token string) error {
	return c.Device().DeleteStorageConfiguration(ctx, token)
}

func (c *Client) SetHashingAlgorithm(ctx context.Context, algorithm string) error {
	return c.Device().SetHashingAlgorithm(ctx, algorithm)
}

func (c *Client) GetDot11Capabilities(ctx context.Context) (*Dot11Capabilities, error) {
	return c.Device().GetDot11Capabilities(ctx)
}

func (c *Client) GetDot11Status(ctx context.Context, interfaceToken string) (*Dot11Status, error) {
	return c.Device().GetDot11Status(ctx, interfaceToken)
}

func (c *Client) GetDot1XConfiguration(ctx context.Context, configToken string) (*Dot1XConfiguration, error) {
	return c.Device().GetDot1XConfiguration(ctx, configToken)
}

func (c *Client) GetDot1XConfigurations(ctx context.Context) ([]*Dot1XConfiguration, error) {
	return c.Device().GetDot1XConfigurations(ctx)
}

func (c *Client) SetDot1XConfiguration(ctx context.Context, config *Dot1XConfiguration) error {
	return c.Device().SetDot1XConfiguration(ctx, config)
}

func (c *Client) CreateDot1XConfiguration(ctx context.Context, config *Dot1XConfiguration) error {
	return c.Device().CreateDot1XConfiguration(ctx, config)
}

func (c *Client) DeleteDot1XConfiguration(ctx context.Context, configToken string) error {
	return c.Device().DeleteDot1XConfiguration(ctx, configToken)
}

func (c *Client) ScanAvailableDot11Networks(ctx context.Context, interfaceToken string) ([]*Dot11AvailableNetworks, error) {
	return c.Device().ScanAvailableDot11Networks(ctx, interfaceToken)
}

// --- transitional delegators: SecurityService ---
func (c *Client) GetRemoteUser(ctx context.Context) (*RemoteUser, error) {
	return c.Security().GetRemoteUser(ctx)
}

func (c *Client) SetRemoteUser(ctx context.Context, remoteUser *RemoteUser) error {
	return c.Security().SetRemoteUser(ctx, remoteUser)
}

func (c *Client) GetIPAddressFilter(ctx context.Context) (*IPAddressFilter, error) {
	return c.Security().GetIPAddressFilter(ctx)
}

func (c *Client) SetIPAddressFilter(ctx context.Context, filter *IPAddressFilter) error {
	return c.Security().SetIPAddressFilter(ctx, filter)
}

func (c *Client) AddIPAddressFilter(ctx context.Context, filter *IPAddressFilter) error {
	return c.Security().AddIPAddressFilter(ctx, filter)
}

func (c *Client) RemoveIPAddressFilter(ctx context.Context, filter *IPAddressFilter) error {
	return c.Security().RemoveIPAddressFilter(ctx, filter)
}

func (c *Client) GetZeroConfiguration(ctx context.Context) (*NetworkZeroConfiguration, error) {
	return c.Security().GetZeroConfiguration(ctx)
}

func (c *Client) SetZeroConfiguration(ctx context.Context, interfaceToken string, enabled bool) error {
	return c.Security().SetZeroConfiguration(ctx, interfaceToken, enabled)
}

func (c *Client) GetDynamicDNS(ctx context.Context) (*DynamicDNSInformation, error) {
	return c.Security().GetDynamicDNS(ctx)
}

func (c *Client) SetDynamicDNS(ctx context.Context, dnsType DynamicDNSType, name string) error {
	return c.Security().SetDynamicDNS(ctx, dnsType, name)
}

func (c *Client) GetPasswordComplexityConfiguration(ctx context.Context) (*PasswordComplexityConfiguration, error) {
	return c.Security().GetPasswordComplexityConfiguration(ctx)
}

func (c *Client) SetPasswordComplexityConfiguration(ctx context.Context, config *PasswordComplexityConfiguration) error {
	return c.Security().SetPasswordComplexityConfiguration(ctx, config)
}

func (c *Client) GetPasswordHistoryConfiguration(ctx context.Context) (*PasswordHistoryConfiguration, error) {
	return c.Security().GetPasswordHistoryConfiguration(ctx)
}

func (c *Client) SetPasswordHistoryConfiguration(ctx context.Context, config *PasswordHistoryConfiguration) error {
	return c.Security().SetPasswordHistoryConfiguration(ctx, config)
}

func (c *Client) GetAuthFailureWarningConfiguration(ctx context.Context) (*AuthFailureWarningConfiguration, error) {
	return c.Security().GetAuthFailureWarningConfiguration(ctx)
}

func (c *Client) SetAuthFailureWarningConfiguration(ctx context.Context, config *AuthFailureWarningConfiguration) error {
	return c.Security().SetAuthFailureWarningConfiguration(ctx, config)
}

// --- transitional delegators: PTZService ---
func (c *Client) ContinuousMove(ctx context.Context, profileToken string, velocity *PTZSpeed, timeout *string) error {
	return c.PTZ().ContinuousMove(ctx, profileToken, velocity, timeout)
}

func (c *Client) AbsoluteMove(ctx context.Context, profileToken string, position *PTZVector, speed *PTZSpeed) error {
	return c.PTZ().AbsoluteMove(ctx, profileToken, position, speed)
}

func (c *Client) RelativeMove(ctx context.Context, profileToken string, translation *PTZVector, speed *PTZSpeed) error {
	return c.PTZ().RelativeMove(ctx, profileToken, translation, speed)
}

func (c *Client) Stop(ctx context.Context, profileToken string, panTilt, zoom bool) error {
	return c.PTZ().Stop(ctx, profileToken, panTilt, zoom)
}

func (c *Client) GetStatus(ctx context.Context, profileToken string) (*PTZStatus, error) {
	return c.PTZ().GetStatus(ctx, profileToken)
}

func (c *Client) GetPresets(ctx context.Context, profileToken string) ([]*PTZPreset, error) {
	return c.PTZ().GetPresets(ctx, profileToken)
}

func (c *Client) GotoPreset(ctx context.Context, profileToken, presetToken string, speed *PTZSpeed) error {
	return c.PTZ().GotoPreset(ctx, profileToken, presetToken, speed)
}

func (c *Client) SetPreset(ctx context.Context, profileToken, presetName, presetToken string) (string, error) {
	return c.PTZ().SetPreset(ctx, profileToken, presetName, presetToken)
}

func (c *Client) RemovePreset(ctx context.Context, profileToken, presetToken string) error {
	return c.PTZ().RemovePreset(ctx, profileToken, presetToken)
}

func (c *Client) GotoHomePosition(ctx context.Context, profileToken string, speed *PTZSpeed) error {
	return c.PTZ().GotoHomePosition(ctx, profileToken, speed)
}

func (c *Client) SetHomePosition(ctx context.Context, profileToken string) error {
	return c.PTZ().SetHomePosition(ctx, profileToken)
}

func (c *Client) GetConfiguration(ctx context.Context, configurationToken string) (*PTZConfiguration, error) {
	return c.PTZ().GetConfiguration(ctx, configurationToken)
}

func (c *Client) GetConfigurations(ctx context.Context) ([]*PTZConfiguration, error) {
	return c.PTZ().GetConfigurations(ctx)
}

// --- transitional delegators: ImagingService ---
func (c *Client) GetImagingSettings(ctx context.Context, videoSourceToken string) (*ImagingSettings, error) {
	return c.Imaging().GetImagingSettings(ctx, videoSourceToken)
}

func (c *Client) SetImagingSettings(ctx context.Context, videoSourceToken string, settings *ImagingSettings, forcePersistence bool) error {
	return c.Imaging().SetImagingSettings(ctx, videoSourceToken, settings, forcePersistence)
}

func (c *Client) Move(ctx context.Context, videoSourceToken string, focus *FocusMove) error {
	return c.Imaging().Move(ctx, videoSourceToken, focus)
}

func (c *Client) GetOptions(ctx context.Context, videoSourceToken string) (*ImagingOptions, error) {
	return c.Imaging().GetOptions(ctx, videoSourceToken)
}

func (c *Client) GetMoveOptions(ctx context.Context, videoSourceToken string) (*MoveOptions, error) {
	return c.Imaging().GetMoveOptions(ctx, videoSourceToken)
}

func (c *Client) StopFocus(ctx context.Context, videoSourceToken string) error {
	return c.Imaging().StopFocus(ctx, videoSourceToken)
}

func (c *Client) GetImagingStatus(ctx context.Context, videoSourceToken string) (*ImagingStatus, error) {
	return c.Imaging().GetImagingStatus(ctx, videoSourceToken)
}

// --- transitional delegators: EventService ---
func (c *Client) SetEventEndpoint(endpoint string) { c.Events().SetEventEndpoint(endpoint) }

func (c *Client) GetEventServiceCapabilities(ctx context.Context) (*EventServiceCapabilities, error) {
	return c.Events().GetEventServiceCapabilities(ctx)
}

func (c *Client) CreatePullPointSubscription(ctx context.Context, filter string, initialTerminationTime *time.Duration, subscriptionPolicy string) (*PullPointSubscription, error) {
	return c.Events().CreatePullPointSubscription(ctx, filter, initialTerminationTime, subscriptionPolicy)
}

func (c *Client) PullMessages(ctx context.Context, subscriptionReference string, timeout time.Duration, messageLimit int) ([]NotificationMessage, error) {
	return c.Events().PullMessages(ctx, subscriptionReference, timeout, messageLimit)
}

func (c *Client) Seek(ctx context.Context, subscriptionReference string, utcTime time.Time, reverse bool) error {
	return c.Events().Seek(ctx, subscriptionReference, utcTime, reverse)
}

func (c *Client) SetEventSynchronizationPoint(ctx context.Context, subscriptionReference string) error {
	return c.Events().SetEventSynchronizationPoint(ctx, subscriptionReference)
}

func (c *Client) Unsubscribe(ctx context.Context, subscriptionReference string) error {
	return c.Events().Unsubscribe(ctx, subscriptionReference)
}

func (c *Client) RenewSubscription(ctx context.Context, subscriptionReference string, terminationTime time.Duration) (time.Time, time.Time, error) {
	return c.Events().RenewSubscription(ctx, subscriptionReference, terminationTime)
}

func (c *Client) GetEventProperties(ctx context.Context) (*EventProperties, error) {
	return c.Events().GetEventProperties(ctx)
}

func (c *Client) AddEventBroker(ctx context.Context, config *EventBrokerConfig) error {
	return c.Events().AddEventBroker(ctx, config)
}

func (c *Client) DeleteEventBroker(ctx context.Context, address string) error {
	return c.Events().DeleteEventBroker(ctx, address)
}

func (c *Client) GetEventBrokers(ctx context.Context) ([]*EventBrokerConfig, error) {
	return c.Events().GetEventBrokers(ctx)
}

// --- transitional delegators: DeviceIOService ---
func (c *Client) GetDeviceIOServiceCapabilities(ctx context.Context) (*DeviceIOServiceCapabilities, error) {
	return c.DeviceIO().GetDeviceIOServiceCapabilities(ctx)
}

func (c *Client) GetDigitalInputs(ctx context.Context) ([]*DigitalInput, error) {
	return c.DeviceIO().GetDigitalInputs(ctx)
}

func (c *Client) GetDigitalInputConfigurationOptions(ctx context.Context, token string) (*DigitalInputConfigurationOptions, error) {
	return c.DeviceIO().GetDigitalInputConfigurationOptions(ctx, token)
}

func (c *Client) SetDigitalInputConfigurations(ctx context.Context, inputs []*DigitalInput) error {
	return c.DeviceIO().SetDigitalInputConfigurations(ctx, inputs)
}

func (c *Client) GetVideoOutputs(ctx context.Context) ([]*VideoOutput, error) {
	return c.DeviceIO().GetVideoOutputs(ctx)
}

func (c *Client) GetSerialPorts(ctx context.Context) ([]*SerialPort, error) {
	return c.DeviceIO().GetSerialPorts(ctx)
}

func (c *Client) GetSerialPortConfiguration(ctx context.Context, serialPortToken string) (*SerialPortConfiguration, error) {
	return c.DeviceIO().GetSerialPortConfiguration(ctx, serialPortToken)
}

func (c *Client) GetSerialPortConfigurationOptions(ctx context.Context, serialPortToken string) (*SerialPortConfigurationOptions, error) {
	return c.DeviceIO().GetSerialPortConfigurationOptions(ctx, serialPortToken)
}

func (c *Client) SetSerialPortConfiguration(ctx context.Context, config *SerialPortConfiguration) error {
	return c.DeviceIO().SetSerialPortConfiguration(ctx, config)
}

func (c *Client) SendReceiveSerialCommand(ctx context.Context, serialPortToken string, data []byte, timeoutSeconds, dataLength int) ([]byte, error) {
	return c.DeviceIO().SendReceiveSerialCommand(ctx, serialPortToken, data, timeoutSeconds, dataLength)
}

func (c *Client) GetVideoOutputConfiguration(ctx context.Context, videoOutputToken string) (*VideoOutputConfiguration, error) {
	return c.DeviceIO().GetVideoOutputConfiguration(ctx, videoOutputToken)
}

func (c *Client) GetVideoOutputConfigurationOptions(ctx context.Context, videoOutputToken string) (*VideoOutputConfigurationOptions, error) {
	return c.DeviceIO().GetVideoOutputConfigurationOptions(ctx, videoOutputToken)
}

func (c *Client) SetVideoOutputConfiguration(ctx context.Context, config *VideoOutputConfiguration) error {
	return c.DeviceIO().SetVideoOutputConfiguration(ctx, config)
}

func (c *Client) GetRelayOutputOptions(ctx context.Context, relayOutputToken string) (*RelayOutputOptions, error) {
	return c.DeviceIO().GetRelayOutputOptions(ctx, relayOutputToken)
}
