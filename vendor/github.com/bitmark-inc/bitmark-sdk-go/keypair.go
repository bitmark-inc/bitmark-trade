package bitmarksdk

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	AlgEd25519    = 1
	AlgCurve25519 = 2
)

var (
	seedNonce = [24]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	authSeedCount = [16]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe7,
	}
	encrSeedCount = [16]byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe8,
	}
)

type AsymmetricKey interface {
	PrivateKeyBytes() []byte
	PublicKeyBytes() []byte
	Algorithm() int
}

type AuthKey interface {
	AsymmetricKey

	Sign(message []byte) (signature []byte)
}

type ED25519AuthKey struct {
	privateKey ed25519.PrivateKey
	test       bool
}

func (e ED25519AuthKey) PrivateKeyBytes() []byte {
	return e.privateKey
}

func (e ED25519AuthKey) PublicKeyBytes() []byte {
	return e.privateKey[ed25519.PrivateKeySize-ed25519.PublicKeySize:]
}

func (e ED25519AuthKey) Algorithm() int {
	return AlgEd25519
}

func (e ED25519AuthKey) Sign(message []byte) []byte {
	return ed25519.Sign(e.PrivateKeyBytes(), message)
}

func NewAuthKey(s *Seed) (AuthKey, error) {
	var seedCore = new([32]byte)
	copy(seedCore[:], s.core)
	authSeed := secretbox.Seal([]byte{}, authSeedCount[:], &seedNonce, seedCore)

	// switch s.version to determine which algorithm to generate auth key
	// if more versions are supported in the future
	_, privateKey, err := ed25519.GenerateKey(bytes.NewBuffer(authSeed))
	return ED25519AuthKey{
		privateKey,
		s.network == Testnet,
	}, err
}

type EncrKey interface {
	AsymmetricKey
	Encrypt(plaintext []byte, peerPublicKey []byte) (ciphertext []byte, err error)
	Decrypt(ciphertext []byte, peerPublicKey []byte) (plaintext []byte, err error)
}

type CURVE25519EncrKey struct {
	publicKey  *[32]byte
	privateKey *[32]byte
}

func (c CURVE25519EncrKey) PrivateKeyBytes() []byte {
	return c.privateKey[:]
}

func (c CURVE25519EncrKey) PublicKeyBytes() []byte {
	return c.publicKey[:]
}

func (c CURVE25519EncrKey) Algorithm() int {
	return AlgCurve25519
}

func (c CURVE25519EncrKey) Encrypt(plaintext []byte, peerPublicKey []byte) ([]byte, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	var publicKey = new([32]byte)
	copy(publicKey[:], peerPublicKey[:])

	ciphertext := box.Seal(nonce[:], plaintext, &nonce, publicKey, c.privateKey)
	return ciphertext, nil
}

func (c CURVE25519EncrKey) Decrypt(ciphertext []byte, peerPublicKey []byte) ([]byte, error) {
	var nonce [24]byte
	copy(nonce[:], ciphertext[:24])

	var publicKey = new([32]byte)
	copy(publicKey[:], peerPublicKey[:])

	plaintext, ok := box.Open(nil, ciphertext[24:], &nonce, publicKey, c.privateKey)
	if !ok {
		return nil, errors.New("decryption failed")
	}

	return plaintext, nil
}

func NewEncrKey(s *Seed) (EncrKey, error) {
	var seedCore = new([32]byte)
	copy(seedCore[:], s.core)
	encrSeed := secretbox.Seal([]byte{}, encrSeedCount[:], &seedNonce, seedCore)

	// switch s.version to determine which algorithm to generate auth key
	// if more versions are supported in the future
	publicKey, privateKey, err := box.GenerateKey(bytes.NewBuffer(encrSeed))
	return CURVE25519EncrKey{publicKey, privateKey}, err
}
