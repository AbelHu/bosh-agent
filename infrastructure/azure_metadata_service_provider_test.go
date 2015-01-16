package infrastructure_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakeinf "github.com/cloudfoundry/bosh-agent/infrastructure/fakes"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	fakeplatform "github.com/cloudfoundry/bosh-agent/platform/fakes"

	. "github.com/cloudfoundry/bosh-agent/infrastructure"
)

var _ = Describe("AzureMetadataServiceProvider", func() {
	var (
		azureMetadataServiceProvider     MetadataServiceProvider
		fakeresolver                     *fakeinf.FakeDNSResolver
		platform                         *fakeplatform.FakePlatform
		logger                           boshlog.Logger
	)

	BeforeEach(func() {
		fakeresolver = &fakeinf.FakeDNSResolver{}
		platform = fakeplatform.NewFakePlatform()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		azureMetadataServiceProvider = NewAzureMetadataServiceProvider(
			fakeresolver,
			platform,
			"/var/lib/waagent/CustomData",
			"/var/lib/waagent/GoalState.1.xml",
			"/var/lib/waagent/ovf-env.xml",
			logger,
		)
	})

	Describe("Get", func() {
		It("returns file metadata service", func() {
			expectedMetadataService := NewAzureFileMetadataService(
				fakeresolver,
				platform.GetFs(),
				"/var/lib/waagent/CustomData",
				"/var/lib/waagent/GoalState.1.xml",
				"/var/lib/waagent/ovf-env.xml",
				logger,
			)
			Expect(azureMetadataServiceProvider.Get()).To(Equal(expectedMetadataService))
		})
	})
})