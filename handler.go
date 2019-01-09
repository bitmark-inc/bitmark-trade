package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/bitmark-inc/bitmark-sdk-go/tx"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/sha3"
)

type issueRequest struct {
	AssetURL   string            `json:"asset_url"`
	Registrant string            `json:"registrant"`
	Name       string            `json:"name"`
	Metadata   map[string]string `json:"metadata"`
	Quantity   int               `json:"quantity"`
}

type transferRequest struct {
	TxId      string `json:"txid"`
	NextOnwer string `json:"owner"`
}

func createAccount() gin.HandlerFunc {
	return func(c *gin.Context) {
		acct, err := account.New()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := service.registerEncPubkey(acct); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := addAccount(acct); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"account": acct.AccountNumber()})
	}
}

func issueBitmarks() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req issueRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request body"})
			return
		}

		issuer, err := getAccount(req.Registrant)
		if err != nil {
			c.JSON(400, gin.H{"error": "owner not registered"})
			return
		}

		assaetURL, _ := url.Parse(req.AssetURL)
		fileName := filepath.Base(assaetURL.Path)
		fileContent, err := ioutil.ReadFile(assaetURL.Path)
		if err != nil {
			c.JSON(400, gin.H{"error": "unable to read asset file"})
			return
		}
		digest := sha3.Sum512(fileContent)
		fingerprint := "01" + hex.EncodeToString(digest[:])
		assetIndex := sha3.Sum512([]byte(fingerprint))
		assetId := hex.EncodeToString(assetIndex[:])

		if err := service.uploadAsset(issuer, assetId, fileName, fileContent); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		a, _ := asset.Get(assetId)
		if a == nil || (a != nil && a.Status != "confirmed") {
			rp, _ := asset.NewRegistrationParams(req.Name, req.Metadata)
			rp.SetFingerprint(fileContent)
			rp.Sign(issuer)
			if _, err := asset.Register(rp); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to register the asset: %s", err.Error())})
				return
			}
		}

		ip := bitmark.NewIssuanceParams(assetId, req.Quantity)
		ip.Sign(issuer)
		bitmarkIds, err := bitmark.Issue(ip)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to issue bitmarks: %s", err.Error())})
			return
		}

		c.JSON(http.StatusOK, gin.H{"bitmark_ids": bitmarkIds})
	}
}

func transferBitmark() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req transferRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request body"})
			return
		}

		// query the current owner
		tx, err := tx.Get(req.TxId, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		currentOwner, err := getAccount(tx.Owner)
		if err != nil {
			c.JSON(400, gin.H{"error": "owner not registered in this service"})
			return
		}

		// handle session data
		access, err := service.getAssetAccess(currentOwner, tx.BitmarkId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		senderPublicKey, err := service.getEncPubkey(access.Sender)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		dataKey, err := dataKeyFromSessionData(currentOwner, access.SessData, senderPublicKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		recipientEncrPubkey, err := service.getEncPubkey(req.NextOnwer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		data, err := createSessionData(currentOwner, dataKey, recipientEncrPubkey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = service.addSessionData(currentOwner, tx.BitmarkId, req.NextOnwer, data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		params := bitmark.NewTransferParams(req.NextOnwer)
		params.FromBitmark(tx.BitmarkId)
		params.Sign(currentOwner)
		txId, err := bitmark.Transfer(params)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"txid": txId})
	}
}

func downloadAsset() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountNo := c.Param("accountNo")
		bitmarkId := c.Param("bitmarkId")

		owner, err := getAccount(accountNo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		access, err := service.getAssetAccess(owner, bitmarkId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fileName, encryptedFileContent, err := service.getAssetContent(access.URL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		senderEncrPubkey, err := service.getEncPubkey(access.Sender)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		dataKey, err := dataKeyFromSessionData(owner, access.SessData, senderEncrPubkey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		plaintext, err := dataKey.Decrypt(encryptedFileContent)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Disposition", "attachment; filename="+fileName)
		c.Header("Content-Length", strconv.Itoa(len(plaintext)))
		c.Data(http.StatusOK, http.DetectContentType(plaintext), plaintext)
		return
	}
}
