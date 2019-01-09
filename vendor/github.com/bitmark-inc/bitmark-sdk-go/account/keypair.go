package account

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
)

const (
	AlgEd25519 = 1
	AlgNaclBox = 2
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

func NewAuthKey(entropy []byte) (AuthKey, error) {
	_, privateKey, err := ed25519.GenerateKey(bytes.NewBuffer(entropy))
	return ED25519AuthKey{
		privateKey,
	}, err
}

type EncrKey interface {
	AsymmetricKey
	Encrypt(plaintext []byte, peerPublicKey []byte) (ciphertext []byte, err error)
	Decrypt(ciphertext []byte, peerPublicKey []byte) (plaintext []byte, err error)
}

type NaclBoxEncrKey struct {
	publicKey  *[32]byte
	privateKey *[32]byte
}

func (n NaclBoxEncrKey) PrivateKeyBytes() []byte {
	return n.privateKey[:]
}

func (n NaclBoxEncrKey) PublicKeyBytes() []byte {
	return n.publicKey[:]
}

func (n NaclBoxEncrKey) Algorithm() int {
	return AlgNaclBox
}

func (n NaclBoxEncrKey) Encrypt(plaintext []byte, peerPublicKey []byte) ([]byte, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	var publicKey = new([32]byte)
	copy(publicKey[:], peerPublicKey[:])

	ciphertext := box.Seal(nonce[:], plaintext, &nonce, publicKey, n.privateKey)
	return ciphertext, nil
}

func (n NaclBoxEncrKey) Decrypt(ciphertext []byte, peerPublicKey []byte) ([]byte, error) {
	var nonce [24]byte
	copy(nonce[:], ciphertext[:24])

	var publicKey = new([32]byte)
	copy(publicKey[:], peerPublicKey[:])

	plaintext, ok := box.Open(nil, ciphertext[24:], &nonce, publicKey, n.privateKey)
	if !ok {
		return nil, errors.New("decryption failed")
	}

	return plaintext, nil
}

func NewEncrKey(entropy []byte) (EncrKey, error) {
	publicKey, privateKey, err := box.GenerateKey(bytes.NewBuffer(entropy))
	return NaclBoxEncrKey{publicKey, privateKey}, err
}
