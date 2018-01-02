package bitmarksdk

import (
	"bytes"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/sha3"
)

const (
	pubkeyMask     = 0x01
	testnetMask    = 0x02
	algorithmShift = 4
	checksumLength = 4
)

type Account struct {
	api     *APIClient
	seed    *Seed
	AuthKey AuthKey
	EncrKey EncrKey
}

func NewAccount(network Network) (*Account, error) {
	seed, err := NewSeed(SeedVersion1, network)
	if err != nil {
		return nil, err
	}

	authKey, err := NewAuthKey(seed)
	if err != nil {
		return nil, err
	}

	encrKey, err := NewEncrKey(seed)
	if err != nil {
		return nil, err
	}

	apiClient := NewAPIClient(network, &http.Client{Timeout: 3 * time.Second})

	account := &Account{apiClient, seed, authKey, encrKey}
	err = account.api.registerEncPubkey(account)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func AccountFromCore(network Network, core []byte) (*Account, error) {
	seed := &Seed{SeedVersion1, network, core}

	authKey, err := NewAuthKey(seed)
	if err != nil {
		return nil, err
	}

	encrKey, err := NewEncrKey(seed)
	if err != nil {
		return nil, err
	}

	apiClient := NewAPIClient(seed.network, &http.Client{Timeout: 3 * time.Second})

	return &Account{apiClient, seed, authKey, encrKey}, nil
}

func AccountFromSeed(s string) (*Account, error) {
	seed, err := SeedFromBase58(s)
	if err != nil {
		return nil, err
	}

	authKey, err := NewAuthKey(seed)
	if err != nil {
		return nil, err
	}

	encrKey, err := NewEncrKey(seed)
	if err != nil {
		return nil, err
	}

	apiClient := NewAPIClient(seed.network, &http.Client{Timeout: 3 * time.Second})

	return &Account{apiClient, seed, authKey, encrKey}, nil
}

func AccountFromRecoveryPhrase(s string) (*Account, error) {
	b, err := phraseToBytes(strings.Split(s, " "))
	if err != nil {
		return nil, err
	}

	network := Livenet
	if b[0] == 0x01 {
		network = Testnet
	}
	seed := Seed{
		SeedVersion1,
		network,
		b[1:],
	}

	return AccountFromSeed(seed.String())
}

func (acct *Account) Network() Network {
	return acct.seed.network
}

func (acct *Account) Core() []byte {
	return acct.seed.core
}

func (acct *Account) Seed() string {
	return acct.seed.String()
}

func (acct *Account) RecoveryPhrase() []string {
	buf := new(bytes.Buffer)
	switch acct.Network() {
	case Livenet:
		buf.Write([]byte{00})
	case Testnet:
		buf.Write([]byte{01})
	}
	buf.Write(acct.seed.core)
	phrase, _ := bytesToPhrase(buf.Bytes())
	return phrase
}

func (acct *Account) AccountNumber() string {
	buffer := acct.bytes()
	checksum := sha3.Sum256(buffer)
	buffer = append(buffer, checksum[:checksumLength]...)
	return toBase58(buffer)
}

func (acct *Account) bytes() []byte {
	keyVariant := byte(acct.AuthKey.Algorithm()<<algorithmShift) | pubkeyMask
	if acct.seed.network == Testnet {
		keyVariant |= testnetMask
	}
	return append([]byte{keyVariant}, acct.AuthKey.PublicKeyBytes()...)
}

func AuthPublicKeyFromAccountNumber(acctNo string) []byte {
	buffer := fromBase58(acctNo)
	return buffer[:len(buffer)-checksumLength]
}
