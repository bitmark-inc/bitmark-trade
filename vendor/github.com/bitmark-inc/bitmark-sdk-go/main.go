package bitmarksdk

import (
	"errors"
	"fmt"
)

func (acct *Account) IssueByAssetFile(af *AssetFile, quantity int) ([]string, error) {
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

	if uerr := acct.api.uploadAsset(acct, af); uerr != nil {
		return nil, uerr
	}
	bitmarkIds, err := acct.api.issue(asset, issues)
	return bitmarkIds, err
}

func (acct *Account) IssueByAssetId(assetId string, quantity int) ([]string, error) {
	issues, err := NewIssueRecords(assetId, acct, quantity)
	if err != nil {
		return nil, err
	}

	bitmarkIds, err := acct.api.issue(nil, issues)
	return bitmarkIds, err
}

// TransferBitmark will transfer a bitmark to others. It will check the owner of a bitmark
// which is going to transfer. If it is valid, a transfer request will be submitted.
// If the target bitmark is private, it will generate a new session data for the new
// receiver.
func (acct *Account) TransferBitmark(bitmarkId, receiver string) (string, error) {
	access, err := acct.api.getAssetAccess(acct, bitmarkId)
	if err != nil {
		return "", err
	}

	if access.SessData != nil {
		senderPublicKey, err := acct.api.getEncPubkey(access.Sender)
		if err != nil {
			return "", err
		}

		dataKey, err := dataKeyFromSessionData(acct, access.SessData, senderPublicKey)
		if err != nil {
			return "", err
		}

		recipientEncrPubkey, err := acct.api.getEncPubkey(receiver)
		if err != nil {
			return "", err
		}

		data, err := createSessionData(acct, dataKey, recipientEncrPubkey)
		if err != nil {
			return "", err
		}

		err = acct.api.addSessionData(acct, bitmarkId, receiver, data)
		if err != nil {
			return "", err
		}
	}

	bmk, err := acct.api.getBitmark(bitmarkId)
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

	return acct.api.transfer(tr)
}

func (acct *Account) DownloadAsset(bitmarkId string) (string, []byte, error) {
	access, err := acct.api.getAssetAccess(acct, bitmarkId)
	if err != nil {
		return "", nil, err
	}

	fileName, content, err := acct.api.getAssetContent(access.URL)
	if err != nil {
		return "", nil, err
	}

	if access.SessData == nil { // public asset
		return fileName, content, nil
	}

	encrPubkey, err := acct.api.getEncPubkey(access.Sender)
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

func (acct *Account) RentBitmark(bitmarkId, receiver string, days uint) error {
	access, err := acct.api.getAssetAccess(acct, bitmarkId)
	if access.SessData == nil {
		return errors.New("no need to rent public assets")
	}

	dataKey, err := dataKeyFromSessionData(acct, access.SessData, acct.EncrKey.PublicKeyBytes())
	if err != nil {
		return err
	}

	recipientEncrPubkey, err := acct.api.getEncPubkey(receiver)
	if err != nil {
		return err
	}

	data, err := createSessionData(acct, dataKey, recipientEncrPubkey)
	if err != nil {
		return err
	}

	return acct.api.updateLease(acct, bitmarkId, receiver, days, data)
}

func (acct *Account) ListLeases() ([]*accessByRenting, error) {
	return acct.api.listLeases(acct)
}

func (acct *Account) DownloadAssetByLease(access *accessByRenting) ([]byte, error) {
	req, _ := newAPIRequest("GET", access.URL, nil)
	content, err := acct.api.submitRequest(req, nil)
	if err != nil {
		return nil, err
	}

	encrPubkey, err := acct.api.getEncPubkey(access.Owner)
	if err != nil {
		return nil, fmt.Errorf("fail to get enc public key: %s", err.Error())
	}

	dataKey, err := dataKeyFromSessionData(acct, access.SessData, encrPubkey)
	if err != nil {
		return nil, err
	}

	plaintext, err := dataKey.Decrypt(content)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
