package infrastructure

import (
	"encoding/json"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsettings "github.com/cloudfoundry/bosh-agent/settings"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type fileMetadataService struct {
	metaDataFilePath string
	userDataFilePath string
	settingsFilePath string
	resolver         DNSResolver
	fs               boshsys.FileSystem

	logger boshlog.Logger
	logTag string
}

func NewFileMetadataService(
	metaDataFilePath string,
	userDataFilePath string,
	settingsFilePath string,
	resolver DNSResolver,
	fs boshsys.FileSystem,
	logger boshlog.Logger,
) MetadataService {
	return fileMetadataService{
		metaDataFilePath: metaDataFilePath,
		userDataFilePath: userDataFilePath,
		settingsFilePath: settingsFilePath,
		resolver:         resolver,
		fs:               fs,

		logTag: "fileMetadataService",
		logger: logger,
	}
}

func (ms fileMetadataService) Load() error {
	return nil
}

func (ms fileMetadataService) GetPublicKey() (string, error) {
	return "", nil
}

func (ms fileMetadataService) GetInstanceID() (string, error) {
	if ms.metaDataFilePath == "" {
		return "", nil
	}

	var metadata MetadataContentsType

	contents, err := ms.fs.ReadFile(ms.metaDataFilePath)
	if err != nil {
		return "", bosherr.WrapError(err, "Reading metadata file")
	}

	err = json.Unmarshal([]byte(contents), &metadata)
	if err != nil {
		return "", bosherr.WrapError(err, "Unmarshalling metadata")
	}

	ms.logger.Debug(ms.logTag, "Read metadata '%#v'", metadata)

	return metadata.InstanceID, nil
}

func (ms fileMetadataService) GetServerName() (string, error) {
	var userData UserDataContentsType

	contents, err := ms.fs.ReadFile(ms.userDataFilePath)
	if err != nil {
		return "", bosherr.WrapError(err, "Reading user data")
	}

	err = json.Unmarshal([]byte(contents), &userData)
	if err != nil {
		return "", bosherr.WrapError(err, "Unmarshalling user data")
	}

	ms.logger.Debug(ms.logTag, "Read user data '%#v'", userData)

	return userData.Server.Name, nil
}

func (ms fileMetadataService) GetRegistryEndpoint() (string, error) {
	var userData UserDataContentsType

	contents, err := ms.fs.ReadFile(ms.userDataFilePath)
	if err != nil {
		// Older versions of bosh-warden-cpi placed
		// full settings file at a specific location.
		return ms.settingsFilePath, nil
	}

	err = json.Unmarshal([]byte(contents), &userData)
	if err != nil {
		return "", bosherr.WrapError(err, "Unmarshalling user data")
	}

	ms.logger.Debug(ms.logTag, "Read user data '%#v'", userData)

	endpoint := userData.Registry.Endpoint
	nameServers := userData.DNS.Nameserver

	if len(nameServers) > 0 {
		endpoint, err = ms.resolver.LookupHost(nameServers, endpoint)
		if err != nil {
			return "", bosherr.WrapError(err, "Resolving registry endpoint")
		}
	}

	return endpoint, nil
}

func (ms fileMetadataService) GetNetworks() (boshsettings.Networks, error) {
	var userData UserDataContentsType

	contents, err := ms.fs.ReadFile(ms.userDataFilePath)
	if err != nil {
		return nil, bosherr.WrapError(err, "Reading user data")
	}

	err = json.Unmarshal([]byte(contents), &userData)
	if err != nil {
		return nil, bosherr.WrapError(err, "Unmarshalling user data")
	}

	ms.logger.Debug(ms.logTag, "Read user data '%#v'", userData)

	return userData.Networks, nil
}

func (ms fileMetadataService) IsAvailable() bool { return true }
