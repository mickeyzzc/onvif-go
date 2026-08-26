package onvif

import "context"

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
