package bitmarksdk

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/sha3"
)

type Network int
type SeedVersion int

// String value for print
func (n Network) String() string {
	switch n {
	case Livenet:
		return "Livenet"
	case Testnet:
		return "Testnet"
	default:
		return "Unknown Network"
	}
}

const SeedVersion1 SeedVersion = 1

const (
	Livenet Network = iota
	Testnet Network = iota
)

const (
	seedHeaderLength   = 3
	seedPrefixLength   = 1
	seedCoreLength     = 32
	seedChecksumLength = 4
	seedLengthNew      = seedHeaderLength + seedPrefixLength + seedCoreLength + seedChecksumLength
)

var (
	seedHeader = []byte{0x5a, 0xfe, 0x01}
)

var (
	ErrSeedSizeMismatch     = errors.New("seed size mismatch")
	ErrSeedHeaderMismatch   = errors.New("seed header mismatch")
	ErrSeedChecksumMismatch = errors.New("seed checksum mismatch")
)

// Seed is used to generate keypairs for authentication and encryption.
type Seed struct {
	version SeedVersion
	network Network
	core    []byte
}

// Returns base58 encoded string on bytes of Seed, which consist of:
//  * Header (3 bytes)
//  * Prefix (1 byte)
//  * Core (32 bytes)
//  * Checksum (4 bytes)
func (s Seed) String() string {
	var b bytes.Buffer
	b.Write(seedHeader)

	seedPrefix := []byte{byte(0x00)}
	if s.network == Testnet {
		seedPrefix = []byte{byte(0x01)}
	}
	b.Write(seedPrefix)

	b.Write(s.core)

	checksum := sha3.Sum256(b.Bytes())
	b.Write(checksum[:seedChecksumLength])

	return toBase58(b.Bytes())
}

func NewSeed(version SeedVersion, network Network) (*Seed, error) {
	var core [32]byte
	if _, err := io.ReadFull(rand.Reader, core[:]); err != nil {
		return nil, err
	}
	return &Seed{version, network, core[:]}, nil
}

func SeedFromBase58(seed string) (*Seed, error) {
	seedBytes := fromBase58(seed)

	if len(seedBytes) != seedLengthNew {
		return nil, ErrSeedSizeMismatch
	}

	if !bytes.Equal(seedBytes[:seedHeaderLength], seedHeader) {
		return nil, ErrSeedHeaderMismatch
	}

	checksum := sha3.Sum256(seedBytes[:seedLengthNew-seedChecksumLength])
	if !bytes.Equal(checksum[:seedChecksumLength], seedBytes[seedLengthNew-seedChecksumLength:]) {
		return nil, ErrSeedChecksumMismatch
	}

	network := Livenet
	if seedBytes[seedHeaderLength : seedHeaderLength+seedPrefixLength][0] == 0x01 {
		network = Testnet
	}

	coreStart := seedHeaderLength + seedPrefixLength
	coreEnd := coreStart + seedCoreLength
	return &Seed{SeedVersion1, network, seedBytes[coreStart:coreEnd]}, nil
}
