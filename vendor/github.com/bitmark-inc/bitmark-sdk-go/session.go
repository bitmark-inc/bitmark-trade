package bitmarksdk

import (
	"net/http"
	"time"
)

type Session struct {
	HTTPClient *http.Client
}

func NewSession(c *http.Client) *Session {
	if c == nil {
		c = &http.Client{Timeout: 5 * time.Second}
	}
	return &Session{c}
}

func (sess *Session) CreateAccount(n Network) (*Account, error) {
	seed, err := NewSeed(SeedVersion1, n)
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

	apiClient := NewAPIClient(n, sess.HTTPClient)

	account := &Account{apiClient, seed, authKey, encrKey}
	err = account.api.registerEncPubkey(account)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (sess *Session) RestoreAccountFromSeed(s string) (*Account, error) {
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

	apiClient := NewAPIClient(seed.network, sess.HTTPClient)

	return &Account{apiClient, seed, authKey, encrKey}, nil
}
