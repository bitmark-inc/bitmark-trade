package bitmarklib

import (
	"bytes"
	"crypto/rand"
	"fmt"

	"github.com/bitmark-inc/bitmarkd/account"
	"github.com/bitmark-inc/bitmarkd/util"
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/sha3"
)

const (
	variantPrivateKey = 0x00
	variantPublicKey  = 0x01

	variantLivenet = 0x00
	variantTestnet = 0x01

	variantKeyTypeED25519 = 0x01
)

const (
	seedLength        = 32
	kifChecksumLength = 4
	checksumLength    = 4
)

// The algorithm type of a keypair
type KeyType int

const (
	Nothing KeyType = iota
	ED25519 KeyType = iota
)

var (
	ErrKIFLength        = fmt.Errorf("kif length is invalid")
	ErrInvalidSeed      = fmt.Errorf("invalid seed")
	ErrInvalidKeyType   = fmt.Errorf("invalid key type")
	ErrInvalidAlgorithm = fmt.Errorf("invalid key algorithm")
	ErrChecksumMismatch = fmt.Errorf("checksum mismatch")
)

type PublicKey struct {
	*account.Account
}

// Keypair is the most important part of bitmark. Every action requires
// a signature which is signed from a keypair.
type KeyPair struct {
	*account.PrivateKey
	seed []byte
}

// KIF returns a KIF string for a keypair
func (kp KeyPair) KIF() (string, error) {

	if kp.seed == nil {
		return "", ErrInvalidSeed
	}

	isTest := kp.PrivateKey.IsTesting()

	var variant uint64 = variantKeyTypeED25519 << 4

	if isTest {
		variant |= variantTestnet << 1
	}

	b := append(util.ToVarint64(variant), kp.seed...)
	checksum := sha3.Sum256(b)
	kifBytes := append(b, checksum[:4]...)

	return util.ToBase58(kifBytes), nil
}

// Seed returns the base58 string of a seed of a keypair
func (kp KeyPair) Seed() string {
	return util.ToBase58(kp.seed)
}

// SeedBytes returns bytes of a seed of a keypair
func (kp KeyPair) SeedBytes() []byte {
	return kp.seed
}

// String returns the base58 string of the private key in a keypair
func (kp KeyPair) String() string {
	return util.ToBase58(kp.PrivateKey.PrivateKeyBytes())
}

// KeyType returns the algorithm type of a keypair
func (kp KeyPair) KeyType() string {
	// HARDCODE: only one algorithm type is in the system now.
	return "ed25519"
}

// NewPublicKey generate a PublicKey struct from a key byte
func NewPublicKey(keyByte []byte) (*PublicKey, error) {
	checksumStart := len(keyByte) - checksumLength
	keyLeft := keyByte[:checksumStart]

	// verify the checksum of public key
	checksum := sha3.Sum256(keyLeft)
	if !bytes.Equal(checksum[:checksumLength], keyByte[checksumStart:]) {
		return nil, ErrChecksumMismatch
	}

	variant := keyLeft[0]
	key := keyLeft[1:]

	if variant&0x01 != variantPublicKey {
		return nil, ErrInvalidKeyType
	}
	test := variant&0x02 != 0

	var ai account.AccountInterface
	variant = variant >> 4
	switch variant & 0x01 {
	case variantKeyTypeED25519:
		ai = &account.ED25519Account{
			PublicKey: key,
			Test:      test,
		}
	default:
		return nil, ErrInvalidAlgorithm
	}

	return &PublicKey{
		&account.Account{
			AccountInterface: ai,
		},
	}, nil
}

// NewPubKeyFromAccount will generate a PublicKey struct from an account string
func NewPubKeyFromAccount(account string) (*PublicKey, error) {
	accountBytes := util.FromBase58(account)
	return NewPublicKey(accountBytes)
}

// NewKeyPair will first generate a seed. Then it use the seed
// to generate a new keypair
func NewKeyPair(test bool, algorithm KeyType) (*KeyPair, error) {
	seedCore := make([]byte, seedLength)
	n, err := rand.Read(seedCore)
	if nil != err {
		return nil, err
	}
	if 32 != n {
		panic("too few random bytes")
	}

	return NewKeyPairFromSeed(seedCore, test, algorithm)
}

// Generate a new keypair with specific seed byte.
func NewKeyPairFromSeed(seed []byte, test bool, algorithm KeyType) (*KeyPair, error) {
	_, priv, err := ed25519.GenerateKey(bytes.NewBuffer(seed))
	if nil != err {
		return nil, err
	}

	privateKey := &account.PrivateKey{
		PrivateKeyInterface: &account.ED25519PrivateKey{
			Test:       test,
			PrivateKey: priv,
		},
	}

	return &KeyPair{
		PrivateKey: privateKey,
		seed:       seed,
	}, nil
}

// Generate a new keypair with specific seed string.
func NewKeyPairFromBase58Seed(seed string, test bool, algorithm KeyType) (*KeyPair, error) {
	seedCore := util.FromBase58(seed)
	return NewKeyPairFromSeed(seedCore, test, algorithm)
}

// Generate keypair from base58 private key. The keypair may be lack of
// network info and algorithm.
func NewKeyPairFromBase58PrivateKey(key string, algorithm KeyType) *KeyPair {
	keyBytes := util.FromBase58(key)
	var p *account.PrivateKey

	switch algorithm {
	case ED25519:
		p = &account.PrivateKey{
			PrivateKeyInterface: &account.ED25519PrivateKey{
				PrivateKey: keyBytes,
			},
		}
	default:
	}

	return &KeyPair{
		PrivateKey: p,
	}
}

// Generate a new keypair from a KIF string
func NewKeyPairFromKIF(kif string) (*KeyPair, error) {
	b := util.FromBase58(kif)
	v, n := util.FromVarint64(b)
	if n+seedLength+kifChecksumLength != len(b) {
		return nil, ErrKIFLength
	}

	if v&0x01 != variantPrivateKey {
		return nil, ErrInvalidKeyType
	}

	seedBytes := b[n : n+seedLength]
	kifChecksum := b[n+seedLength:]

	checksum := sha3.Sum256(b[:n+seedLength])
	if !bytes.Equal(checksum[:kifChecksumLength], kifChecksum) {
		return nil, ErrChecksumMismatch
	}

	test := v&0x02 != 0

	v = v >> 4
	switch v & 0x01 {
	case variantKeyTypeED25519:
		return NewKeyPairFromSeed(seedBytes, test, ED25519)
	default:
		return nil, ErrInvalidAlgorithm
	}
}

type EncrKeyPair struct {
	PrivateKey *[32]byte
	PublicKey  *[32]byte
}

// Generate a new encryption keypair with the given seed.
func NewEncrKeyPairFromSeed(seed []byte) (*EncrKeyPair, error) {
	pub, pvt, err := box.GenerateKey(bytes.NewBuffer(seed))
	if err != nil {
		return nil, err
	}

	return &EncrKeyPair{PrivateKey: pvt, PublicKey: pub}, nil
}
