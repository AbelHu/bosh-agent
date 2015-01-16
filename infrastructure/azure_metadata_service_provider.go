package infrastructure

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshplatform "github.com/cloudfoundry/bosh-agent/platform"
)

type azureMetadataServiceProvider struct {
	resolver               DNSResolver
	platform               boshplatform.Platform
	userDataFilePath       string
	goalstateFilePath      string
	ovfenvFilePath         string
	logger                 boshlog.Logger
	logTag                 string
}

func NewAzureMetadataServiceProvider(
	resolver DNSResolver,
	platform boshplatform.Platform,
	userDataFilePath string,
	goalstateFilePath string,
	ovfenvFilePath string,
	logger boshlog.Logger,
) azureMetadataServiceProvider {
	return azureMetadataServiceProvider{
		resolver:               resolver,
		platform:               platform,
		userDataFilePath:       userDataFilePath,
		goalstateFilePath:      goalstateFilePath,
		ovfenvFilePath:         ovfenvFilePath,
		logger:                 logger,
		logTag:                 "AzureMetadataServiceProvider",
	}
}

func (inf azureMetadataServiceProvider) Get() MetadataService {
	metadataService := NewAzureFileMetadataService(inf.resolver, inf.platform.GetFs(), inf.userDataFilePath, inf.goalstateFilePath, inf.ovfenvFilePath, inf.logger)
	return metadataService
}