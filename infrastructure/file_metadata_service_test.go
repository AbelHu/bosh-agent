package infrastructure_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-agent/infrastructure"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeinf "github.com/cloudfoundry/bosh-agent/infrastructure/fakes"
)

var _ = Describe("FileMetadataService", func() {
	var (
		dnsResolver     *fakeinf.FakeDNSResolver
		fs              *fakesys.FakeFileSystem
		metadataService MetadataService
	)

	Describe("All files exist", func() {
		BeforeEach(func() {
			dnsResolver = &fakeinf.FakeDNSResolver{}
			fs = fakesys.NewFakeFileSystem()
			logger := boshlog.NewLogger(boshlog.LevelNone)
			metadataService = NewFileMetadataService(
				"fake-metadata-file-path",
				"fake-userdata-file-path",
				"fake-settings-file-path",
				dnsResolver,
				fs,
				logger,
			)
		})

		Describe("GetInstanceID", func() {
			Context("when metadata service file exists", func() {
				BeforeEach(func() {
					metadataContents := `{"instance-id":"fake-instance-id"}`
					fs.WriteFileString("fake-metadata-file-path", metadataContents)
				})

				It("returns instance id", func() {
					instanceID, err := metadataService.GetInstanceID()
					Expect(err).NotTo(HaveOccurred())
					Expect(instanceID).To(Equal("fake-instance-id"))
				})
			})

			Context("when metadata service file does not exist", func() {
				It("returns an error", func() {
					instanceID, err := metadataService.GetInstanceID()
					Expect(err).To(HaveOccurred())
					Expect(instanceID).To(BeEmpty())
				})
			})
		})

		Describe("GetServerName", func() {
			Context("when userdata file exists", func() {
				BeforeEach(func() {
					userDataContents := `{"server":{"name":"fake-server-name"}}`
					fs.WriteFileString("fake-userdata-file-path", userDataContents)
				})

				It("returns server name", func() {
					serverName, err := metadataService.GetServerName()
					Expect(err).NotTo(HaveOccurred())
					Expect(serverName).To(Equal("fake-server-name"))
				})
			})

			Context("when userdata file does not exist", func() {
				It("returns an error", func() {
					serverName, err := metadataService.GetServerName()
					Expect(err).To(HaveOccurred())
					Expect(serverName).To(BeEmpty())
				})
			})
		})

		Describe("GetNetworks", func() {
			It("returns the network settings", func() {
				userDataContents := `
					{
						"networks": {
							"network_1": {"type": "manual", "ip": "1.2.3.4", "netmask": "2.3.4.5", "gateway": "3.4.5.6", "default": ["dns"], "dns": ["8.8.8.8"], "mac": "fake-mac-address-1"},
							"network_2": {"type": "dynamic", "default": ["dns"], "dns": ["8.8.8.8"], "mac": "fake-mac-address-2"}
						}
					}`
				fs.WriteFileString("fake-userdata-file-path", userDataContents)

				networks, err := metadataService.GetNetworks()
				Expect(err).ToNot(HaveOccurred())
				Expect(networks).To(Equal(boshsettings.Networks{
					"network_1": boshsettings.Network{
						Type:    "manual",
						IP:      "1.2.3.4",
						Netmask: "2.3.4.5",
						Gateway: "3.4.5.6",
						Default: []string{"dns"},
						DNS:     []string{"8.8.8.8"},
						Mac:     "fake-mac-address-1",
					},
					"network_2": boshsettings.Network{
						Type:    "dynamic",
						Default: []string{"dns"},
						DNS:     []string{"8.8.8.8"},
						Mac:     "fake-mac-address-2",
					},
				}))
			})

			It("returns a nil Networks if the settings are missing (from an old CPI version)", func() {
				userDataContents := `{}`
				fs.WriteFileString("fake-userdata-file-path", userDataContents)

				networks, err := metadataService.GetNetworks()
				Expect(err).ToNot(HaveOccurred())
				Expect(networks).To(BeNil())
			})

			It("raises an error if we can't read the file", func() {
				networks, err := metadataService.GetNetworks()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Reading user data: File not found"))
				Expect(networks).To(BeNil())
			})
		})

		Describe("GetRegistryEndpoint", func() {
			Context("when userdata file contains a dns server", func() {
				BeforeEach(func() {
					userDataContents := `{
						"registry":{"endpoint":"http://fake-registry.com"},
						"dns":{"nameserver":["fake-dns-server-ip"]}
					}`
					fs.WriteFileString("fake-userdata-file-path", userDataContents)
				})

				Context("when registry endpoint is successfully resolved", func() {
					BeforeEach(func() {
						dnsResolver.RegisterRecord(fakeinf.FakeDNSRecord{
							DNSServers: []string{"fake-dns-server-ip"},
							Host:       "http://fake-registry.com",
							IP:         "http://fake-registry-ip",
						})
					})

					It("returns the successfully resolved registry endpoint", func() {
						endpoint, err := metadataService.GetRegistryEndpoint()
						Expect(err).NotTo(HaveOccurred())
						Expect(endpoint).To(Equal("http://fake-registry-ip"))
					})
				})

				Context("when registry endpoint is not successfully resolved", func() {
					BeforeEach(func() {
						dnsResolver.LookupHostErr = errors.New("fake-lookup-host-err")
					})

					It("returns error because it failed to resolve registry endpoint", func() {
						endpoint, err := metadataService.GetRegistryEndpoint()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-lookup-host-err"))
						Expect(endpoint).To(BeEmpty())
					})
				})
			})

			Context("when userdata file does not contain dns servers", func() {
				Context("when userdata file exists", func() {
					BeforeEach(func() {
						userDataContents := `{"registry":{"endpoint":"fake-registry-endpoint"}}`
						fs.WriteFileString("fake-userdata-file-path", userDataContents)
					})

					It("returns registry endpoint", func() {
						registryEndpoint, err := metadataService.GetRegistryEndpoint()
						Expect(err).NotTo(HaveOccurred())
						Expect(registryEndpoint).To(Equal("fake-registry-endpoint"))
					})
				})

				Context("when userdata file does not exist", func() {
					It("returns registry endpoint pointing to a settings file", func() {
						registryEndpoint, err := metadataService.GetRegistryEndpoint()
						Expect(err).NotTo(HaveOccurred())
						Expect(registryEndpoint).To(Equal("fake-settings-file-path"))
					})
				})
			})
		})

		Describe("IsAvailable", func() {
			It("returns true", func() {
				Expect(metadataService.IsAvailable()).To(BeTrue())
			})
		})
	})

	Describe("Only userdata file exists", func() {
		BeforeEach(func() {
			dnsResolver = &fakeinf.FakeDNSResolver{}
			fs = fakesys.NewFakeFileSystem()
			logger := boshlog.NewLogger(boshlog.LevelNone)
			metadataService = NewFileMetadataService(
				"",
				"fake-userdata-file-path",
				"",
				dnsResolver,
				fs,
				logger,
			)
		})

		Describe("GetInstanceID", func() {
			It("returns empty", func() {
				instanceID, err := metadataService.GetInstanceID()
				Expect(err).NotTo(HaveOccurred())
				Expect(instanceID).To(BeEmpty())
			})
		})

		Describe("GetServerName", func() {
			Context("when userdata file exists", func() {
				BeforeEach(func() {
					userDataContents := `{"server":{"name":"fake-server-name"}}`
					fs.WriteFileString("fake-userdata-file-path", userDataContents)
				})

				It("returns server name", func() {
					serverName, err := metadataService.GetServerName()
					Expect(err).NotTo(HaveOccurred())
					Expect(serverName).To(Equal("fake-server-name"))
				})
			})

			Context("when userdata file does not exist", func() {
				It("returns an error", func() {
					serverName, err := metadataService.GetServerName()
					Expect(err).To(HaveOccurred())
					Expect(serverName).To(BeEmpty())
				})
			})
		})

		Describe("GetNetworks", func() {
			It("returns the network settings", func() {
				userDataContents := `
					{
						"networks": {
							"network_1": {"type": "manual", "ip": "1.2.3.4", "netmask": "2.3.4.5", "gateway": "3.4.5.6", "default": ["dns"], "dns": ["8.8.8.8"], "mac": "fake-mac-address-1"},
							"network_2": {"type": "dynamic", "default": ["dns"], "dns": ["8.8.8.8"], "mac": "fake-mac-address-2"}
						}
					}`
				fs.WriteFileString("fake-userdata-file-path", userDataContents)

				networks, err := metadataService.GetNetworks()
				Expect(err).ToNot(HaveOccurred())
				Expect(networks).To(Equal(boshsettings.Networks{
					"network_1": boshsettings.Network{
						Type:    "manual",
						IP:      "1.2.3.4",
						Netmask: "2.3.4.5",
						Gateway: "3.4.5.6",
						Default: []string{"dns"},
						DNS:     []string{"8.8.8.8"},
						Mac:     "fake-mac-address-1",
					},
					"network_2": boshsettings.Network{
						Type:    "dynamic",
						Default: []string{"dns"},
						DNS:     []string{"8.8.8.8"},
						Mac:     "fake-mac-address-2",
					},
				}))
			})

			It("returns a nil Networks if the settings are missing (from an old CPI version)", func() {
				userDataContents := `{}`
				fs.WriteFileString("fake-userdata-file-path", userDataContents)

				networks, err := metadataService.GetNetworks()
				Expect(err).ToNot(HaveOccurred())
				Expect(networks).To(BeNil())
			})

			It("raises an error if we can't read the file", func() {
				networks, err := metadataService.GetNetworks()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Reading user data: File not found"))
				Expect(networks).To(BeNil())
			})
		})

		Describe("GetRegistryEndpoint", func() {
			Context("when userdata file contains a dns server", func() {
				BeforeEach(func() {
					userDataContents := `{
						"registry":{"endpoint":"http://fake-registry.com"},
						"dns":{"nameserver":["fake-dns-server-ip"]}
					}`
					fs.WriteFileString("fake-userdata-file-path", userDataContents)
				})

				Context("when registry endpoint is successfully resolved", func() {
					BeforeEach(func() {
						dnsResolver.RegisterRecord(fakeinf.FakeDNSRecord{
							DNSServers: []string{"fake-dns-server-ip"},
							Host:       "http://fake-registry.com",
							IP:         "http://fake-registry-ip",
						})
					})

					It("returns the successfully resolved registry endpoint", func() {
						endpoint, err := metadataService.GetRegistryEndpoint()
						Expect(err).NotTo(HaveOccurred())
						Expect(endpoint).To(Equal("http://fake-registry-ip"))
					})
				})

				Context("when registry endpoint is not successfully resolved", func() {
					BeforeEach(func() {
						dnsResolver.LookupHostErr = errors.New("fake-lookup-host-err")
					})

					It("returns error because it failed to resolve registry endpoint", func() {
						endpoint, err := metadataService.GetRegistryEndpoint()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-lookup-host-err"))
						Expect(endpoint).To(BeEmpty())
					})
				})
			})

			Context("when userdata file does not contain dns servers", func() {
				Context("when userdata file exists", func() {
					BeforeEach(func() {
						userDataContents := `{"registry":{"endpoint":"fake-registry-endpoint"}}`
						fs.WriteFileString("fake-userdata-file-path", userDataContents)
					})

					It("returns registry endpoint", func() {
						registryEndpoint, err := metadataService.GetRegistryEndpoint()
						Expect(err).NotTo(HaveOccurred())
						Expect(registryEndpoint).To(Equal("fake-registry-endpoint"))
					})
				})

				Context("when userdata file does not exist", func() {
					It("returns registry endpoint pointing to a settings file", func() {
						registryEndpoint, err := metadataService.GetRegistryEndpoint()
						Expect(err).NotTo(HaveOccurred())
						Expect(registryEndpoint).To(BeEmpty())
					})
				})
			})
		})

		Describe("IsAvailable", func() {
			It("returns true", func() {
				Expect(metadataService.IsAvailable()).To(BeTrue())
			})
		})
	})
})
