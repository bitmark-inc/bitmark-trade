package bitmarklib

import (
	"bytes"
	"github.com/bitmark-inc/bitmarkd/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewKeypair(t *testing.T) {
	p, err := NewKeyPair(true, ED25519)
	assert.NoError(t, err)
	assert.Len(t, p.seed, 32)
	assert.Len(t, p.PrivateKeyBytes(), 64)
	assert.True(t, bytes.Equal(p.PrivateKeyBytes()[32:], p.Account().PublicKeyBytes()))
}

func TestFromKIF(t *testing.T) {
	kif := "cYK2SzQnYLG55yiRCSryymEw3EaNYnCD2mtCwkVXdFLSzQ4ReV"
	kp, err := NewKeyPairFromKIF(kif)
	assert.NoError(t, err)
	assert.True(t, kp.PrivateKey.IsTesting())
	privateKey := util.ToBase58(kp.PrivateKey.PrivateKeyBytes())
	assert.Equal(t,
		"2T2LWi7M7Qz9vx3vaMiNCwzreSGEkBMkDcbMhq2Ss5LZTtWWAXUxVjk7N5Gg1guTW1XMHu4wrTXBik8EmVkPvZcK",
		privateKey,
	)
}

func TestNewKeypairFromBase58Seed(t *testing.T) {
	seed := "8VNLU6LSMjnCfMNHG9YftLV1TVWzAphfCSwJsf351974"

	kp, err := NewKeyPairFromBase58Seed(seed, true, ED25519)
	assert.NoError(t, err)
	assert.True(t, kp.PrivateKey.IsTesting())
	privateKey := util.ToBase58(kp.PrivateKey.PrivateKeyBytes())
	assert.Equal(t, "3E2z4v1HjJUtxzh8CQeDN3NJ1THPs2bcuPiqJrEFNZLhw58gcWBuM1ZouCwwZZxBmZztkdEaRJrmLLoBpyb7dEoz", privateKey)
}

func TestKIFEncodeDecode(t *testing.T) {
	p1, err := NewKeyPair(true, ED25519)
	assert.NoError(t, err)

	kif, err := p1.KIF()
	assert.NoError(t, err)

	p2, err := NewKeyPairFromKIF(kif)
	assert.NoError(t, err)

	assert.True(t, bytes.Equal(
		p1.PrivateKey.PrivateKeyBytes(),
		p2.PrivateKey.PrivateKeyBytes()))
}
