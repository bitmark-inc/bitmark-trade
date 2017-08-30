package bitmarklib

import (
	"bytes"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

var (
	seedNonce = [24]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	accountSeedCountBM = [16]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe7,
	}
	accessSeedCountBM = [16]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe8,
	}
)

type AccessKeyPair struct {
	PrivateKey *[32]byte
	PublicKey  *[32]byte
}

func NewAccessKeyPairFromSeed(seed []byte) (*AccessKeyPair, error) {
	var secretKey [32]byte
	copy(secretKey[:], seed)

	encryptedAccessSeed := secretbox.Seal([]byte{}, accessSeedCountBM[:], &seedNonce, &secretKey)

	pub, pri, err := box.GenerateKey(bytes.NewBuffer(encryptedAccessSeed))
	if err != nil {
		return nil, err
	}

	return &AccessKeyPair{PrivateKey: pri, PublicKey: pub}, nil
}
