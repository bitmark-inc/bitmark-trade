package main

import (
	"bytes"
	"fmt"

	bmksdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/encoding"
	"github.com/boltdb/bolt"
	"golang.org/x/crypto/sha3"
)

func getAccountBucketName() []byte {
	return []byte(fmt.Sprintf("account-%s", string(bmksdk.GetNetwork())))
}

func getAccount(accountNo string) (account.Account, error) {
	var val []byte

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(getAccountBucketName())
		val = b.Get([]byte(accountNo))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get the account from db: %s", err)
	}

	if val == nil {
		return nil, fmt.Errorf("account %s not registered", accountNo)
	}

	var seed string
	switch len(val) {
	case 32:
		var b bytes.Buffer

		// write the seed header
		b.Write([]byte{0x5a, 0xfe, 0x01})

		// write the network
		switch bmksdk.GetNetwork() {
		case bmksdk.Livenet:
			b.Write([]byte{byte(0x00)})
		case bmksdk.Testnet:
			b.Write([]byte{byte(0x01)})
		}

		// write the core 32 bytes
		b.Write(val)

		// write the checksum
		checksum := sha3.Sum256(b.Bytes())
		b.Write(checksum[:4])

		seed = encoding.ToBase58(b.Bytes())
	case 33:
		seed = string(val)
	default:
		return nil, fmt.Errorf("invalid account format: %s", string(val))
	}

	return account.FromSeed(seed)
}

func addAccount(acct account.Account) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(getAccountBucketName())
		return b.Put([]byte(acct.AccountNumber()), []byte(acct.Seed()))
	})
}
