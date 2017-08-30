package bitmarklib

import (
	"testing"

	"github.com/bitmark-inc/bitmarkd/transactionrecord"
	"github.com/stretchr/testify/assert"
)

func TestNewAsset(t *testing.T) {
	a := NewAsset("testcase", "test_fingerprint")

	assert.Equal(t, "testcase", a.Name)
	assert.Equal(t, "test_fingerprint", a.Fingerprint)
}

func TestAssetSetMeta(t *testing.T) {
	a := NewAsset("testcase", "test_fingerpring")
	err := a.SetMeta(map[string]string{
		"test1": "test",
		"test2": "test",
	})
	assert.NoError(t, err)
	assert.Contains(t, a.Metadata, "test1\u0000test")
	assert.Contains(t, a.Metadata, "test2\u0000test")
}

func TestAssetSetMetaWithEmptyKey(t *testing.T) {
	a := NewAsset("testcase", "test_fingerpring")
	err := a.SetMeta(map[string]string{
		"": "test",
	})
	assert.Equal(t, ErrEmptyMetaKeyValue, err)
}

func TestAssetSetMetaWithEmptyValue(t *testing.T) {
	a := NewAsset("testcase", "test_fingerpring")
	err := a.SetMeta(map[string]string{
		"test": "",
	})
	assert.Equal(t, ErrEmptyMetaKeyValue, err)
}
func TestSignAsset(t *testing.T) {
	owner, err := NewKeyPair(true, ED25519)
	assert.NoError(t, err)

	a := NewAsset("testcase", "test_fingerprint")
	err = a.SetMeta(map[string]string{
		"test1": "test",
		"test2": "test",
	})
	assert.NoError(t, err)
	err = a.Sign(owner)
	assert.NoError(t, err)
	assert.NotEmpty(t, a.Signature)
}

func TestAssetSignWithoutMetadata(t *testing.T) {
	owner, err := NewKeyPair(true, ED25519)
	assert.NoError(t, err)

	a := NewAsset("testcase", "test_fingerprint")
	err = a.Sign(owner)
	assert.Equal(t, "", a.Metadata)
}

func TestNewIssue(t *testing.T) {
	i := NewIssue(transactionrecord.NewAssetIndex([]byte{0, 0, 0, 0, 0, 0, 0, 0}))
	assert.Len(t, i.AssetIndex, 64)
}

func TestSignIssue(t *testing.T) {
	owner, err := NewKeyPair(true, ED25519)
	assert.NoError(t, err)

	i := NewIssue(transactionrecord.NewAssetIndex([]byte{0, 0, 0, 0, 0, 0, 0, 0}))
	err = i.Sign(owner)
	assert.NoError(t, err)
	assert.NotEmpty(t, i.Nonce)
	assert.NotEmpty(t, i.Signature)
}
