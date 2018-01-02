package bitmarksdk

import (
	"errors"
	"fmt"
	"net/http"
)

type Config struct {
	HTTPClient *http.Client
	Network    Network

	APIEndpoint string
	KeyEndpoint string
}

type Client struct {
	Network Network
	service *Service
}

func NewClient(cfg *Config) *Client {
	var apiEndpoint string
	var keyEndpoint string
	switch cfg.Network {
	case Testnet:
		apiEndpoint = "https://api.test.bitmark.com"
		keyEndpoint = "https://assets.test.bitmark.com"
	case Livenet:
		apiEndpoint = "https://api.bitmark.com"
		keyEndpoint = "https://assets.bitmark.com"
	}

	// allow endpoints customization
	if cfg.APIEndpoint != "" {
		apiEndpoint = cfg.APIEndpoint
	}
	if cfg.KeyEndpoint != "" {
		keyEndpoint = cfg.KeyEndpoint
	}

	svc := &Service{cfg.HTTPClient, apiEndpoint, keyEndpoint}
	return &Client{cfg.Network, svc}
}

func (c *Client) CreateAccount() (*Account, error) {
	seed, err := NewSeed(SeedVersion1, c.Network)
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

	account := &Account{seed: seed, AuthKey: authKey, EncrKey: encrKey}

	if err := c.service.registerEncPubkey(account); err != nil {
		return nil, err
	}
	return account, nil
}

func (c *Client) RestoreAccountFromSeed(s string) (*Account, error) {
	seed, err := SeedFromBase58(s)
	if err != nil {
		return nil, err
	}

	if seed.network != c.Network {
		return nil, fmt.Errorf("trying to restore %s account in %s environment", seed.network, c.Network)
	}

	authKey, err := NewAuthKey(seed)
	if err != nil {
		return nil, err
	}

	encrKey, err := NewEncrKey(seed)
	if err != nil {
		return nil, err
	}

	return &Account{seed: seed, AuthKey: authKey, EncrKey: encrKey}, nil
}

func (c *Client) IssueByAssetFile(acct *Account, af *AssetFile, quantity int) ([]string, error) {
	var asset *AssetRecord
	if af.propertyName != "" {
		var err error
		asset, err = NewAssetRecord(af.propertyName, af.Fingerprint, af.propertyMetadata, acct)
		if err != nil {
			return nil, err
		}
	}

	issues, err := NewIssueRecords(af.Id(), acct, quantity)
	if err != nil {
		return nil, err
	}

	if uerr := c.service.uploadAsset(acct, af); uerr != nil {
		return nil, uerr
	}
	bitmarkIds, err := c.service.createIssueTx(asset, issues)
	return bitmarkIds, err
}

func (c *Client) IssueByAssetId(acct *Account, assetId string, quantity int) ([]string, error) {
	issues, err := NewIssueRecords(assetId, acct, quantity)
	if err != nil {
		return nil, err
	}

	bitmarkIds, err := c.service.createIssueTx(nil, issues)
	return bitmarkIds, err
}

func (c *Client) Transfer(acct *Account, bitmarkId, receiver string) (string, error) {
	access, aerr := c.service.getAssetAccess(acct, bitmarkId)
	if aerr != nil {
		return "", aerr
	}

	if access.SessData != nil {
		senderPublicKey, err := c.service.getEncPubkey(access.Sender)
		if err != nil {
			return "", err
		}

		dataKey, err := dataKeyFromSessionData(acct, access.SessData, senderPublicKey)
		if err != nil {
			return "", err
		}

		recipientEncrPubkey, err := c.service.getEncPubkey(receiver)
		if err != nil {
			return "", err
		}

		data, err := createSessionData(acct, dataKey, recipientEncrPubkey)
		if err != nil {
			return "", err
		}

		err = c.service.addSessionData(acct, bitmarkId, receiver, data)
		if err != nil {
			return "", err
		}
	}

	bmk, err := c.service.getBitmark(bitmarkId)
	if err != nil {
		return "", err
	}

	if acct.AccountNumber() != bmk.Owner {
		return "", errors.New("not bitmark owner")
	}

	tr, err := NewTransferRecord(bmk.HeadId, receiver, acct)
	if err != nil {
		return "", err
	}

	return c.service.createTransferTx(tr)
}

func (c *Client) SignTransferOffer(sender *Account, bitmarkId, receiver string) (*TransferOffer, error) {
	access, aerr := c.service.getAssetAccess(sender, bitmarkId)
	if aerr != nil {
		return nil, aerr
	}

	if access.SessData != nil {
		senderPublicKey, err := c.service.getEncPubkey(access.Sender)
		if err != nil {
			return nil, err
		}

		dataKey, err := dataKeyFromSessionData(sender, access.SessData, senderPublicKey)
		if err != nil {
			return nil, err
		}

		recipientEncrPubkey, err := c.service.getEncPubkey(receiver)
		if err != nil {
			return nil, err
		}

		data, err := createSessionData(sender, dataKey, recipientEncrPubkey)
		if err != nil {
			return nil, err
		}

		err = c.service.addSessionData(sender, bitmarkId, receiver, data)
		if err != nil {
			return nil, err
		}
	}

	bmk, err := c.service.getBitmark(bitmarkId)
	if err != nil {
		return nil, err
	}

	if sender.AccountNumber() != bmk.Owner {
		return nil, errors.New("not bitmark owner")
	}

	return NewTransferOffer(bitmarkId, bmk.HeadId, receiver, sender)
}

func (c *Client) CountersignTransfer(receiver *Account, t *TransferOffer) (string, error) {
	record, err := t.Countersign(receiver)
	if err != nil {
		return "", err
	}
	return c.service.createCountersignTransferTx(record)
}

func (c *Client) DownloadAsset(acct *Account, bitmarkId string) (string, []byte, error) {
	access, err := c.service.getAssetAccess(acct, bitmarkId)
	if err != nil {
		return "", nil, err
	}

	fileName, content, err := c.service.getAssetContent(access.URL)
	if err != nil {
		return "", nil, err
	}

	if access.SessData == nil { // public asset
		return fileName, content, nil
	}

	encrPubkey, err := c.service.getEncPubkey(access.Sender)
	if err != nil {
		return "", nil, fmt.Errorf("fail to get enc public key: %s", err.Error())
	}

	dataKey, err := dataKeyFromSessionData(acct, access.SessData, encrPubkey)
	if err != nil {
		return "", nil, err
	}

	plaintext, err := dataKey.Decrypt(content)
	if err != nil {
		return "", nil, err
	}

	return fileName, plaintext, nil
}
