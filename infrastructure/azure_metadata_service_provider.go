package infrastructure

import (
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshplatform "github.com/cloudfoundry/bosh-agent/platform"
)

type azureMetadataServiceProvider struct {
	resolver               DNSResolver
	platform               boshplatform.Platform
	walaLibPath            string
	logger                 boshlog.Logger
	logTag                 string
}

func NewAzureMetadataServiceProvider(
	resolver DNSResolver,
	platform boshplatform.Platform,
	walaLibPath string,
	logger boshlog.Logger,
) azureMetadataServiceProvider {
	return azureMetadataServiceProvider{
		resolver:          resolver,
		platform:          platform,
		walaLibPath:       walaLibPath,
		logger:            logger,
		logTag:            "AzureMetadataServiceProvider",
	}
}

func (inf azureMetadataServiceProvider) Get() MetadataService {
	metadataService := NewAzureFileMetadataService(inf.resolver, inf.platform.GetFs(), inf.walaLibPath, inf.logger)
	return metadataService
}