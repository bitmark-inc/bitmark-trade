package main

import (
	"net/http"

	"github.com/bitmark-inc/bitmark-trade/bmservice"
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

		tx, err := bmservice.GetTx(req.TxId)
		if err != nil {
			checkErr(c, err)
			return
		}

		keypair, err := getBitmarkKeypair(tx.Tx.Owner)
		if err != nil {
			c.JSON(400, gin.H{"message": "owner not registered"})
			return
		}

		txId, err := bmservice.Transfer(keypair, req.TxId, req.NextOnwer)
		if err != nil {
			checkErr(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"txid": txId})
	}
}
