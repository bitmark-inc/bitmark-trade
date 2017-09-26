package main

import (
	"net/http"
	"strconv"

	"github.com/bitmark-inc/bitmark-trade/bmservice"
	bitmarklib "github.com/bitmark-inc/go-bitmarklib"
	"github.com/gin-gonic/gin"
)

func handleAssetDownload() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountNo := c.Param("accountNo")
		bitmarkId := c.Param("bitmarkId")

		bitmark, err := bmservice.GetBitmark(bitmarkId)
		if err != nil {
			checkErr(c, err)
			return
		}

		owner, err := getAccount(accountNo)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "owner account not found"})
			return
		}

		prevOwner, err := getAccount(bitmark.PreviousOwner(""))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "previous owner not registered in this service"})
			return
		}

		issuer, err := getAccount(bitmark.Issuer)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "issuer not registered in this service"})
			return
		}

		filename, ciphertext, err := bmservice.Download(owner.AuthKeyPair, bitmarkId)
		if err != nil {
			checkErr(c, err)
			return
		}

		sessData, err := bmservice.GetSessionData(owner.AccountNumber(), bitmarkId)
		if err != nil {
			checkErr(c, err)
			return
		}

		sessKey, err := bitmarklib.SessionKeyFromSessionData(sessData, prevOwner.EncrKeyPair.PublicKey, owner.EncrKeyPair.PrivateKey, prevOwner.AuthKeyPair.Account().PublicKeyBytes())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "wrong encryption from the previous owner",
				"error":   err.Error(),
			})
			return
		}

		plaintext, err := bitmarklib.DecryptAssetFile(ciphertext, sessKey, issuer.AuthKeyPair.Account().PublicKeyBytes())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "wrong encryption from the previous owner",
				"error":   err.Error(),
			})
			return
		}

		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Length", strconv.Itoa(len(plaintext)))
		c.Data(http.StatusOK, http.DetectContentType(plaintext), plaintext)
		return
	}
}
