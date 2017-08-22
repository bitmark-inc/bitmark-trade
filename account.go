package main

import (
	"net/http"

	bitmarklib "github.com/bitmark-inc/go-bitmarklib"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

func handleAccountCreation() gin.HandlerFunc {
	return func(c *gin.Context) {
		account, err := createBitmarkAccount()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"account": account})
	}
}

func createBitmarkAccount() (string, error) {
	keypair, err := bitmarklib.NewKeyPair(testnet, bitmarklib.ED25519)
	if err != nil {
		return "", err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("account"))
		return b.Put([]byte(keypair.Account().String()), []byte(keypair.Seed()))
	})
	if err != nil {
		return "", err
	}

	return keypair.Account().String(), nil
}

func getBitmarkKeypair(account string) (*bitmarklib.KeyPair, error) {
	var seed string

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("account"))
		seed = string(b.Get([]byte(account)))
		return nil
	})
	if err != nil {
		return nil, err
	}

	keypair, err := bitmarklib.NewKeyPairFromBase58Seed(seed, testnet, bitmarklib.ED25519)
	if err != nil {
		return nil, err
	}

	return keypair, nil
}
