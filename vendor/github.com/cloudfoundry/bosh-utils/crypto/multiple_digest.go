package crypto

import (
	"strings"
	"io"
	"errors"
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type MultipleDigest struct {
	digests []Digest
}

func MustNewMultipleDigest(digests ...Digest) MultipleDigest {
	if len(digests) == 0 {
		panic("no digests have been provided")
	}
	return MultipleDigest{digests}
}

func MustParseMultipleDigest(json string) MultipleDigest {
	var digest MultipleDigest
	err := (&digest).UnmarshalJSON([]byte(json))
	if err != nil {
		panic(fmt.Sprintf("Parsing multiple digest: %s", err))
	}
	return digest
}

func (m MultipleDigest) String() string { return m.strongestDigest().String() }
func (m MultipleDigest) Algorithm() Algorithm { return m.strongestDigest().Algorithm() }

func (m MultipleDigest) Verify(reader io.Reader) error {
	err := m.validate()
	if err != nil {
		return err
	}

	return m.strongestDigest().Verify(reader)
}

func (m MultipleDigest) validate() error {
	if len(m.digests) == 0 {
		return errors.New("Expected to find at least one digest")
	}

	algosUsed := map[string]struct{}{}

	for _, digest := range m.digests {
		algoName := digest.Algorithm().Name()

		if _, found := algosUsed[algoName]; found {
			return bosherr.Errorf("Multiple digests of the same algorithm '%s' found in digests '%s'", algoName, m.fullString())
		}

		algosUsed[algoName] = struct{}{}
	}

	return nil
}

func (m MultipleDigest) strongestDigest() (Digest) {
	if len(m.digests) == 0 {
		panic("no digests have been provided")
	}

	preferredAlgorithms := []Algorithm{DigestAlgorithmSHA512, DigestAlgorithmSHA256, DigestAlgorithmSHA1}

	for _, algo := range preferredAlgorithms {
		for _, digest := range m.digests {
			if digest.Algorithm().Name() == algo.Name() {
				return digest
			}
		}
	}

	return m.digests[0]
}

func (m *MultipleDigest) UnmarshalJSON(data []byte) error {
	digestString := strings.TrimSuffix(strings.TrimPrefix(string(data), `"`), `"`)

	multiDigest, err := m.parseMultipleDigestString(digestString)
	if err != nil {
		return err
	}

	err = multiDigest.validate()
	if err != nil {
		return err
	}

	*m = multiDigest

	return nil
}

func (m MultipleDigest) fullString() string {
	var result []string

	for _, digest := range m.digests {
		result = append(result, digest.String())
	}

	return strings.Join(result, ";")
}

func (m MultipleDigest) MarshalJSON() ([]byte, error) {
	if len(m.digests) == 0 {
		return nil, errors.New("no digests have been provided")
	}

	return []byte(fmt.Sprintf(`"%s"`, m.fullString())), nil
}

func (m MultipleDigest) parseMultipleDigestString(multipleDigest string) (MultipleDigest, error) {
	pieces := strings.Split(multipleDigest, ";")

	digests := []Digest{}

	for _, digest := range pieces {
		parsedDigest, err := m.parseDigestString(digest)
		if err == nil {
			digests = append(digests, parsedDigest)
		}
	}

	if len(digests) == 0 {
		return MultipleDigest{}, errors.New("No recognizable digest algorithm found. Supported algorithms: sha1, sha256, sha512")
	}

	return MultipleDigest{digests: digests}, nil
}

func (MultipleDigest) parseDigestString(digest string) (Digest, error) {
	if len(digest) == 0 {
		return nil, errors.New("Can not parse empty string.")
	}

	pieces := strings.SplitN(digest, ":", 2)

	if len(pieces) == 1 {
		// historically digests were only sha1 and did not include a prefix.
		// continue to support that behavior.
		pieces = []string{"sha1", pieces[0]}
	}

	switch pieces[0] {
	case "sha1":
		return NewDigest(DigestAlgorithmSHA1, pieces[1]), nil
	case "sha256":
		return NewDigest(DigestAlgorithmSHA256, pieces[1]), nil
	case "sha512":
		return NewDigest(DigestAlgorithmSHA512, pieces[1]), nil
	default:
		return NewDigest(NewUnknownAlgorithm(pieces[0]), pieces[1]), nil
	}
}
