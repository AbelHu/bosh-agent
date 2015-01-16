package infrastructure

import (
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshdpresolv "github.com/cloudfoundry/bosh-agent/infrastructure/devicepathresolver"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshplatform "github.com/cloudfoundry/bosh-agent/platform"
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
)

type azureInfrastructure struct {
	metadataService    MetadataService
	registry           Registry
	platform           boshplatform.Platform
	devicePathResolver boshdpresolv.DevicePathResolver
	logger             boshlog.Logger
}

func NewAzureInfrastructure(
	metadataService MetadataService,
	registry Registry,
	platform boshplatform.Platform,
	devicePathResolver boshdpresolv.DevicePathResolver,
	logger boshlog.Logger,
) azureInfrastructure {

	return azureInfrastructure{
		metadataService:    metadataService,
		registry:           registry,
		platform:           platform,
		devicePathResolver: devicePathResolver,
		logger:             logger,
	}
}

func NewAzureRegistry(metadataService MetadataService) Registry {
	return NewHTTPRegistry(metadataService, false)
}

func (inf azureInfrastructure) GetDevicePathResolver() boshdpresolv.DevicePathResolver {
	return inf.devicePathResolver
}

func (inf azureInfrastructure) SetupSSH(username string) error {
	publicKey, err := inf.metadataService.GetPublicKey()
	if err != nil {
		return bosherr.WrapError(err, "Error getting public key")
	}

	return inf.platform.SetupSSH(publicKey, username)
}

func (inf azureInfrastructure) GetSettings() (boshsettings.Settings, error) {
	registry := inf.registry
	settings, err := registry.GetSettings()
	if err != nil {
		return settings, bosherr.WrapError(err, "Getting settings from registry")
	}

	return settings, nil
}

func (inf azureInfrastructure) SetupNetworking(networks boshsettings.Networks) (err error) {
	return inf.platform.SetupDhcp(networks)
}

func (inf azureInfrastructure) GetEphemeralDiskPath(diskSettings boshsettings.DiskSettings) string {
	return "/dev/sdb"
}
