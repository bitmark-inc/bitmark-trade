package asset

import (
	"encoding/hex"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/utils"
	"golang.org/x/crypto/sha3"
)

const (
	minNameLength     = 1
	maxNameLength     = 64
	maxMetadataLength = 2048
)

var (
	ErrInvalidNameLength     = errors.New("property name not set or exceeds the maximum length (64 Unicode characters)")
	ErrInvalidMetadataLength = errors.New("property metadata exceeds the maximum length (1024 Unicode characters)")
	ErrEmptyContent          = errors.New("asset content is empty")
	ErrNullRegistrant        = errors.New("registrant is null")
)

type RegistrationParams struct {
	Name        string `json:"name" pack:"utf8"`
	Fingerprint string `json:"fingerprint" pack:"utf8"`
	Metadata    string `json:"metadata" pack:"utf8"`
	Registrant  string `json:"registrant" pack:"account"`
	Signature   string `json:"signature"`
}

func NewRegistrationParams(name string, metadata map[string]string) (*RegistrationParams, error) {
	parts := make([]string, 0, len(metadata)*2)
	for key, val := range metadata {
		if key == "" || val == "" {
			continue
		}
		parts = append(parts, key, val)
	}
	compactMetadata := strings.Join(parts, "\u0000")

	if utf8.RuneCountInString(name) < minNameLength || utf8.RuneCountInString(name) > maxNameLength {
		return nil, ErrInvalidNameLength
	}

	if utf8.RuneCountInString(compactMetadata) > maxMetadataLength {
		return nil, ErrInvalidMetadataLength
	}

	return &RegistrationParams{
		Name:     name,
		Metadata: compactMetadata,
	}, nil
}

func (r *RegistrationParams) SetFingerprint(content []byte) error {
	if len(content) == 0 {
		return ErrEmptyContent
	}
	r.Fingerprint = computeFingerprint(1, content)
	return nil
}

func (r *RegistrationParams) Sign(registrant account.Account) error {
	if registrant == nil {
		return ErrNullRegistrant
	}
	r.Registrant = registrant.AccountNumber()

	message, err := utils.Pack(r)
	if err != nil {
		return err
	}
	r.Signature = hex.EncodeToString(registrant.Sign(message))

	return nil
}

func computeFingerprint(version int, content []byte) string {
	// TODO: support more fingerptint versions if necessary
	digest := sha3.Sum512(content)
	return "01" + hex.EncodeToString(digest[:])
}
