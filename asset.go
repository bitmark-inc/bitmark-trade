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

		prevOwnerEncrPubkey, err := bmservice.GetEncrPubkey(bitmark.PreviousOwner(""))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "encr public key of the previous owner not found"})
			return
		}
		prevOwnerAuthPubkey, err := authPublicKeyFromAccountNumber(bitmark.PreviousOwner(""))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "auth public key of the previous owner not found"})
			return
		}

		issuerAuthPubkey, err := authPublicKeyFromAccountNumber(bitmark.Issuer)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "auth public key of the issuer not found"})
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

		sessKey, err := bitmarklib.SessionKeyFromSessionData(sessData, prevOwnerEncrPubkey, owner.EncrKeyPair.PrivateKey, prevOwnerAuthPubkey.PublicKeyBytes())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "wrong encryption from the previous owner",
				"error":   err.Error(),
			})
			return
		}

		plaintext, err := bitmarklib.DecryptAssetFile(ciphertext, sessKey, issuerAuthPubkey.PublicKeyBytes())
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
