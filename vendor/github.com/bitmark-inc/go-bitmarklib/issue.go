package bitmarklib

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bitmark-inc/bitmarkd/transactionrecord"
	"golang.org/x/crypto/ed25519"
)

// The index prevents generate dupliocated nonces
var nonceIndex uint64

var (
	ErrEmptyMetaKeyValue = fmt.Errorf("key and value of metadata can not be empty")
)

// Asset includes name, meta and fingerprint of the actual digital property
type Asset struct {
	transactionrecord.AssetData
}

// NewAsset will return an asset struct
func NewAsset(name string, fingerprint string) Asset {
	return Asset{
		AssetData: transactionrecord.AssetData{
			Name:        name,
			Fingerprint: fingerprint,
			Signature:   []byte{},
		},
	}
}

// SetMeta will turns a map of string into a string with key
// and value split by \u0000. Both key and value key not be
// empty string
func (a *Asset) SetMeta(metadata map[string]string) error {
	metaStrings := make([]string, 0, len(metadata)*2)
	for key, val := range metadata {
		if key == "" || val == "" {
			return ErrEmptyMetaKeyValue
		}
		metaStrings = append(metaStrings, key, val)
	}
	a.Metadata = strings.Join(metaStrings, "\u0000")
	return nil
}

// Sign an asset with a keypair and write the signature into
// Signature field
func (a *Asset) Sign(kp *KeyPair) error {
	a.Registrant = kp.Account()

	packed, _ := a.Pack(a.Registrant)
	if nil == packed {
		return fmt.Errorf("fail to pack an asset")
	}

	a.Signature = ed25519.Sign(kp.PrivateKeyBytes(), packed)
	_, err := a.Pack(a.Registrant)
	return err
}

// Issue is to claim the ownership to a specific asset.
type Issue struct {
	transactionrecord.BitmarkIssue
}

// NewIssue will return an Issue struct
func NewIssue(assetIndex transactionrecord.AssetIndex) Issue {
	return Issue{
		BitmarkIssue: transactionrecord.BitmarkIssue{
			AssetIndex: assetIndex,
		},
	}
}

// Sign an issue with a keypair and write the signature into
// Signature field
func (i *Issue) Sign(kp *KeyPair) error {
	atomic.AddUint64(&nonceIndex, 1)
	i.Nonce = uint64(time.Now().UTC().Unix())*1000 + nonceIndex%1000
	i.Owner = kp.Account()

	packed, _ := i.Pack(i.Owner)
	if nil == packed {
		return fmt.Errorf("fail to pack an issue")
	}

	i.Signature = ed25519.Sign(kp.PrivateKeyBytes(), packed)
	_, err := i.Pack(i.Owner)
	return err
}
