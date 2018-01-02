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

		txId, err := bmk.Transfer(owner, tx.Tx.BitmarkId, req.NextOnwer)
		if err != nil {
			checkErr(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"txid": txId})
	}
}
