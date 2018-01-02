package main

import (
	"fmt"
	"net/http"

	bmksdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
)

func handleAccountCreation() gin.HandlerFunc {
	return func(c *gin.Context) {
		account, err := createAccount()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"account": account})
	}
}

func createAccount() (string, error) {
	acct, err := bmk.CreateAccount()
	if err != nil {
		return "", err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(getAccountBucketName())
		return b.Put([]byte(acct.AccountNumber()), acct.Core())
	})
	if err != nil {
		return "", err
	}

	return acct.AccountNumber(), nil

	// account, err := NewAccount(testnet)
	// if err != nil {
	// 	return "", err
	// }
	//
	// err = db.Update(func(tx *bolt.Tx) error {
	// 	b := tx.Bucket([]byte(getAccountBucketName()))
	// 	return b.Put([]byte(account.AccountNumber()), account.SeedBytes())
	// })
	// if err != nil {
	// 	return "", err
	// }
	//
	// sig := ed25519.Sign(account.AuthKeyPair.PrivateKeyBytes(), account.EncrKeyPair.PublicKey[:])
	// err = bmservice.RegisterEncrPubkey(
	// 	account.AccountNumber(), account.EncrKeyPair.PublicKey[:], sig)
	// if err != nil {
	// 	return "", err
	// }
	//
	// return account.AccountNumber(), nil
}

func getAccount(accountNo string) (*bmksdk.Account, error) {
	var core []byte

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(getAccountBucketName())
		core = b.Get([]byte(accountNo))
		return nil
	})
	if err != nil {
		return nil, err
	}

	if core == nil {
		return nil, fmt.Errorf("account %s not registered", accountNo)
	}

	return bmksdk.AccountFromCore(bmk.Network, core)

	// var seed []byte
	//
	// err := db.View(func(tx *bolt.Tx) error {
	// 	b := tx.Bucket([]byte(getAccountBucketName()))
	// 	seed = b.Get([]byte(accountNo))
	// 	return nil
	// })
	// if err != nil {
	// 	return nil, err
	// }
	//
	// if seed == nil {
	// 	return nil, fmt.Errorf("account %s not registered", accountNo)
	// }
	//
	// account, err := NewAccountFromSeed(seed, testnet)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// return account, nil
}

// var (
// 	seedNonce = [24]byte{
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 	}
// 	authSeedCountBM = [16]byte{
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe7,
// 	}
// 	encrSeedCountBM = [16]byte{
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe8,
// 	}
// )
//
// type Account struct {
// 	seed        []byte
// 	AuthKeyPair *bitmarklib.KeyPair
// 	EncrKeyPair *bitmarklib.EncrKeyPair
// }
//
// func NewAccountFromSeed(seed []byte, test bool) (*Account, error) {
// 	var secretKey [32]byte
// 	copy(secretKey[:], seed)
//
// 	authSeed := createAuthSeed(secretKey)
// 	authKeypair, err := bitmarklib.NewKeyPairFromSeed(authSeed, test, bitmarklib.ED25519)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	encrSeed := createEncrSeed(secretKey)
// 	encrKeypair, err := bitmarklib.NewEncrKeyPairFromSeed(encrSeed)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &Account{seed, authKeypair, encrKeypair}, nil
// }
//
// func NewAccount(test bool) (*Account, error) {
// 	seed, err := generateSeed(32)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return NewAccountFromSeed(seed, test)
// }
//
// func (a *Account) AccountNumber() string {
// 	return a.AuthKeyPair.Account().String()
// }
//
// func (a *Account) SeedBytes() []byte {
// 	return a.seed
// }
//
// func generateSeed(size int) ([]byte, error) {
// 	seed := make([]byte, size)
// 	if _, err := io.ReadFull(rand.Reader, seed); err != nil {
// 		return nil, err
// 	}
//
// 	return seed, nil
// }
//
// func createAuthSeed(seed [32]byte) []byte {
// 	return secretbox.Seal([]byte{}, authSeedCountBM[:], &seedNonce, &seed)
// }
//
// func createEncrSeed(seed [32]byte) []byte {
// 	return secretbox.Seal([]byte{}, encrSeedCountBM[:], &seedNonce, &seed)
// }
//
// func authPublicKeyFromAccountNumber(acctNo string) (*bitmarklib.PublicKey, error) {
// 	return bitmarklib.NewPubKeyFromAccount(acctNo)
// }
