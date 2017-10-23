package main

import (
	"net/http"

	"github.com/bitmark-inc/bitmark-trade/bmservice"
	bitmarklib "github.com/bitmark-inc/go-bitmarklib"
	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	TxId      string `json:"txid"`
	NextOnwer string `json:"owner"`
}

func handleTransfer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req transferRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"message": "invalid request body"})
			return
		}

		// query the current owner
		tx, err := bmservice.GetTx(req.TxId)
		if err != nil {
			checkErr(c, err)
			return
		}
		owner, err := getAccount(tx.Tx.Owner)
		if err != nil {
			c.JSON(400, gin.H{"message": "owner not registered in this service"})
			return
		}

		// query the previous owner
		bitmark, err := bmservice.GetBitmark(tx.Tx.BitmarkId)
		if err != nil {
			checkErr(c, err)
			return
		}
		prevOwner, err := getAccount(bitmark.PreviousOwner(req.TxId))
		if err != nil {
			c.JSON(400, gin.H{"message": "previous owner not registered"})
			return
		}

		nextOwnerEncrPubkey, err := bmservice.GetEncrPubkey(req.NextOnwer)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "encr public key of the previous owner not found"})
			return
		}

		log.Infof("%s | %s | %s", prevOwner.AccountNumber(), owner.AccountNumber(), req.NextOnwer)

		var txId string
		if bitmark.Owner != req.NextOnwer {
			transfer, _ := bitmarklib.NewTransfer(req.TxId, req.NextOnwer, testnet)
			err = transfer.Sign(owner.AuthKeyPair)
			if err != nil {
				c.JSON(400, gin.H{"message": "unable to sign"})
				return
			}
			txId, err = bmservice.TransferBitmark(transfer)
			if err != nil {
				checkErr(c, err)
				return
			}
		} else {
			txId = bitmark.HeadId
		}

		sessData, err := bmservice.GetSessionData(owner.AccountNumber(), tx.Tx.BitmarkId)
		if err != nil {
			checkErr(c, err)
			return
		}

		sessKey, err := bitmarklib.SessionKeyFromSessionData(sessData, prevOwner.EncrKeyPair.PublicKey, owner.EncrKeyPair.PrivateKey, prevOwner.AuthKeyPair.Account().PublicKeyBytes())
		if err != nil {
			c.JSON(400, gin.H{"message": "invalid session key"})
			return
		}

		newSessionData, err := bitmarklib.CreateSessionData(sessKey, nextOwnerEncrPubkey, owner.EncrKeyPair.PrivateKey, owner.AuthKeyPair.PrivateKeyBytes())
		if err != nil {
			c.JSON(400, gin.H{"message": "unable to create session data"})
			return
		}

		err = bmservice.UpdateSessionData(owner.AuthKeyPair, newSessionData, tx.Tx.BitmarkId, req.NextOnwer)
		if err != nil {
			checkErr(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"txid": txId})
	}
}
