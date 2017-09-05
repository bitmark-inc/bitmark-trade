package bitmarklib

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
)

const (
	Chacha20poly1305 = iota
)

type SessionData struct {
	EncryptedSessionKey          []byte
	EncryptedSessionKeySignature []byte
	SessionKeySignature          []byte
}

func (d *SessionData) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		EncryptedSessionKey          string `json:"enc_skey"`
		EncryptedSessionKeySignature string `json:"enc_skey_sig"`
		SessionKeySignature          string `json:"skey_sig"`
	}{
		EncryptedSessionKey:          hex.EncodeToString(d.EncryptedSessionKey),
		EncryptedSessionKeySignature: hex.EncodeToString(d.EncryptedSessionKeySignature),
		SessionKeySignature:          hex.EncodeToString(d.SessionKeySignature),
	})
}

func (d *SessionData) UnmarshalJSON(data []byte) error {
	var aux struct {
		EncryptedSessKey          string `json:"enc_skey"`
		EncryptedSessKeySignature string `json:"enc_skey_sig"`
		SessKeySignature          string `json:"skey_sig"`
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	d.EncryptedSessionKey, _ = hex.DecodeString(aux.EncryptedSessKey)
	d.EncryptedSessionKeySignature, _ = hex.DecodeString(aux.EncryptedSessKeySignature)
	d.SessionKeySignature, _ = hex.DecodeString(aux.SessKeySignature)
	return nil
}

type SessionKey interface {
	String() string
	Bytes() []byte
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

func SessionKeyFromSessionData(data *SessionData, sEncryptionPubkey, rEncryptionPvtkey *[32]byte, sAuthPubkey []byte) (SessionKey, error) {
	if !ed25519.Verify(sAuthPubkey, data.EncryptedSessionKey, data.EncryptedSessionKeySignature) {
		return nil, errors.New("invalid encrypted session key signature")
	}

	var decryptNonce [24]byte
	copy(decryptNonce[:], data.EncryptedSessionKey[:24])
	decrypted, ok := box.Open(nil, data.EncryptedSessionKey[24:], &decryptNonce, sEncryptionPubkey, rEncryptionPvtkey)
	if !ok {
		return nil, errors.New("unable to decrypt")
	}

	if !ed25519.Verify(sAuthPubkey, decrypted, data.SessionKeySignature) {
		return nil, errors.New("invalid session key signature")
	}

	return &ChaCha20SessionKey{key: decrypted}, nil
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

func (k *ChaCha20SessionKey) Bytes() []byte {
	return k.key
}

// Encrypt the plaintext using zero nonce
func (k *ChaCha20SessionKey) Encrypt(plaintext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(k.key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	return ciphertext, nil
}

// Decrypt the ciphertext using zero nonce
func (k *ChaCha20SessionKey) Decrypt(ciphertext []byte) ([]byte, error) {
	aead, err := chacha20poly1305.New(k.key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, chacha20poly1305.NonceSize)

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// EncryptAssetFile generates encrypted asset file content, which consists of:
// ciphertext of the asset file
// signature of the asset file in plaintext
func EncryptAssetFile(content []byte, key SessionKey, authPvtkey []byte) ([]byte, error) {
	ciphertext, err := key.Encrypt(content)
	if err != nil {
		return nil, err
	}

	signature := ed25519.Sign(authPvtkey, content)

	b := bytes.NewBuffer(ciphertext)
	b.Write(signature)

	return b.Bytes(), nil
}

// DecryptAssetFile decrypts encrypted asset file content and verifies the included signature
func DecryptAssetFile(content []byte, key SessionKey, authPubkey []byte) ([]byte, error) {
	if len(content) < ed25519.SignatureSize {
		return nil, errors.New("invalid encrypted file size")
	}
	ciphertext := content[:len(content)-ed25519.SignatureSize]
	signature := content[len(content)-ed25519.SignatureSize:]

	if len(signature) != ed25519.SignatureSize {
		return nil, errors.New("invalid signature size")
	}

	plaintext, err := key.Decrypt(ciphertext)
	if err != nil {
		return nil, err
	}

	if !ed25519.Verify(authPubkey, plaintext, signature) {
		return nil, errors.New("invalid signature")
	}

	return plaintext, nil
}

// CreateSessionData creates the SessionData of a SessionKey
func CreateSessionData(sessKey SessionKey, recipientEncryptionPubkey, senderEncryptionPvtkey *[32]byte, senderAuthPvtkey []byte) (*SessionData, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		panic(err)
	}
	encryptedKey := box.Seal(nonce[:], sessKey.Bytes(), &nonce, recipientEncryptionPubkey, senderEncryptionPvtkey)

	return &SessionData{
		EncryptedSessionKey:          encryptedKey,
		EncryptedSessionKeySignature: ed25519.Sign(senderAuthPvtkey, encryptedKey),
		SessionKeySignature:          ed25519.Sign(senderAuthPvtkey, sessKey.Bytes()),
	}, nil
}
