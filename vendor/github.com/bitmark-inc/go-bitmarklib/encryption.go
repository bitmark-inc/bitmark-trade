package bitmarklib

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ed25519"
)

const (
	Chacha20poly1305 = iota
)

const (
	ciphertextCountSize = 8
)

type SessionKey interface {
	String() string
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

func SessionKeyFromHex(alg int, sessionKey string) (SessionKey, error) {
	key, err := hex.DecodeString(sessionKey)
	if err != nil {
		return nil, errors.New("invalid session key")
	}

	switch alg {
	case Chacha20poly1305:
		return &ChaCha20SessionKey{key: key}, nil
	default:
		return nil, errors.New("unsupported algorithm")
	}
}

type ChaCha20SessionKey struct {
	key []byte
}

func NewChaCha20SessionKey() (*ChaCha20SessionKey, error) {
	key := make([]byte, chacha20poly1305.KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	return &ChaCha20SessionKey{key: key}, nil
}

func (k *ChaCha20SessionKey) String() string {
	return hex.EncodeToString(k.key)
}

func (k *ChaCha20SessionKey) Encrypt(plaintext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(k.key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	count := make([]byte, ciphertextCountSize)
	binary.LittleEndian.PutUint64(count, uint64(len(ciphertext)))

	buf := make([]byte, 0, ciphertextCountSize+len(ciphertext))
	b := bytes.NewBuffer(buf)
	b.Write(count)
	b.Write(ciphertext)

	return b.Bytes(), nil
}

func (k *ChaCha20SessionKey) Decrypt(encryptedContent []byte) ([]byte, error) {
	b := bytes.NewBuffer(encryptedContent)

	nonce := make([]byte, chacha20poly1305.NonceSize)

	n := make([]byte, ciphertextCountSize)
	_, err := io.ReadFull(b, n)
	if err != nil {
		return nil, errors.New("invalid ciphertext count")
	}
	count := binary.LittleEndian.Uint64(n)

	ciphertext := make([]byte, count)
	_, err = io.ReadFull(b, ciphertext)
	if err != nil {
		return nil, errors.New("invalid ciphertext")
	}

	aead, err := chacha20poly1305.New(k.key)
	if err != nil {
		return nil, err
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func CreateEncryptedFile(src, dst string, key SessionKey, bitmarkPrivatekey []byte) error {
	plaintext, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	encryptedData, err := key.Encrypt(plaintext)
	if err != nil {
		return err
	}

	signature := ed25519.Sign(bitmarkPrivatekey, plaintext)

	f, err := os.Create(dst)
	defer f.Close()
	if err != nil {
		return err
	}

	// TODO: handle errors
	w := bufio.NewWriter(f)
	w.Write(encryptedData)
	w.Write(signature)
	w.Flush()

	return nil
}

func CreateDecryptedFile(src, dst string, key SessionKey, bitmarkPublickey []byte) error {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	if len(content) < ed25519.SignatureSize {
		return errors.New("invalid encrypted file size")
	}
	encryptedData := content[:len(content)-ed25519.SignatureSize]
	signature := content[len(content)-ed25519.SignatureSize:]
	if len(signature) != ed25519.SignatureSize {
		return errors.New("invalid signature size")
	}

	plaintext, err := key.Decrypt(encryptedData)
	if err != nil {
		return err
	}

	if !ed25519.Verify(bitmarkPublickey, plaintext, signature) {
		return errors.New("invalid signature")
	}

	f, err := os.Create(dst)
	defer f.Close()
	if err != nil {
		return err
	}
	// TODO: handle potential error
	f.Write(plaintext)

	return nil
}
