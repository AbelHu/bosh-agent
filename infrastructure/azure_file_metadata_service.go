package infrastructure

import (
	"regexp"
	"encoding/json"
	"encoding/base64"
	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	boshsys "github.com/cloudfoundry/bosh-agent/system"
)

type azureFileMetadataService struct {
	resolver          DNSResolver
	fs                boshsys.FileSystem
	walaLibPath       string
	logger            boshlog.Logger
	logTag            string
}

func NewAzureFileMetadataService(
	resolver DNSResolver,
	fs boshsys.FileSystem,
	walaLibPath string,
	logger boshlog.Logger,
) azureFileMetadataService {
	return azureFileMetadataService{
		resolver:     resolver,
		fs:           fs,
		walaLibPath:  walaLibPath,
		logger:       logger,
		logTag:       "azureFileMetadataService",
	}
}

func (ms azureFileMetadataService) Load() error {
	return nil
}

func (ms azureFileMetadataService) GetPublicKey() (string, error) {
	contents, err := ms.fs.ReadFileString(ms.walaLibPath + "/ovf-env.xml")
	if err != nil {
		return "", bosherr.WrapError(err, "Reading ovf-env.xml")
	}

	re := regexp.MustCompile("<UserName>(.*)</UserName>")
	match := re.FindStringSubmatch(contents)
	if match == nil {
		return "", bosherr.WrapError(err, "Reading ovf-env file")
	}

	publicKey, err := ms.fs.ReadFileString("/home/" + match[1] + "/.ssh/authorized_keys")
	if err != nil {
		return "", bosherr.WrapError(err, "Reading public key file")
	}
	return publicKey, nil
}

func (ms azureFileMetadataService) GetInstanceID() (string, error) {
	return ms.GetServerName()
}

func (ms azureFileMetadataService) GetServerName() (string, error) {
	userData, err := ms.getUserData()
	if err != nil {
		return "", bosherr.WrapError(err, "Getting user data")
	}

	serverName := userData.Server.Name

	if len(serverName) == 0 {
		return "", bosherr.Error("Empty server name")
	}

	return serverName, nil
}

func (ms azureFileMetadataService) GetRegistryEndpoint() (string, error) {
	userData, err := ms.getUserData()
	if err != nil {
		return "", bosherr.WrapError(err, "Getting user data")
	}

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

func (ms azureFileMetadataService) getUserData() (UserDataContentsType, error) {
	var userData UserDataContentsType
	
	contents, err := ms.fs.ReadFile(ms.walaLibPath + "/CustomData")
	if err != nil {
		return userData, bosherr.WrapError(err, "Reading user data file")
	}
	
	data := make([]byte, base64.StdEncoding.DecodedLen(len(contents)))
	n, err := base64.StdEncoding.Decode(data, contents)
	if err != nil {
		return userData, bosherr.WrapError(err, "Decoding user data")
	}
	
	err = json.Unmarshal(data[0:n], &userData)
	if err != nil {
		return userData, bosherr.WrapError(err, "Unmarshalling user data")
	}

	return userData, nil
}


func (ms azureFileMetadataService) IsAvailable() bool {
	return true
}
