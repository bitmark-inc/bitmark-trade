package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/crypto/sha3"

	"github.com/bitmark-inc/bitmark-trade/bmservice"
	"github.com/gin-gonic/gin"
)

type issueRequest struct {
	AssetURL   string            `json:"asset_url"`
	Registrant string            `json:"registrant"`
	Name       string            `json:"name"`
	Metadata   map[string]string `json:"metadata"`
	Quantity   int               `json:"quantity"`
}

// TODO:
func readAsset(u string) ([]byte, string, string, error) {
	result, err := url.Parse(u)
	if err != nil {
		fmt.Println(err.Error())
		return nil, "", "", err
	}

	switch result.Scheme {
	case "file":
		dat, err := ioutil.ReadFile(result.Path)
		if err != nil {
			fmt.Println(err.Error())
			return nil, "", "", err
		}
		return dat, result.Path, computeFingerprint(dat), err
	default:
		return nil, "", "", errors.New("scheme not supported for asset_url")
	}
}

func computeFingerprint(content []byte) string {
	digest := sha3.Sum512(content)
	return hex.EncodeToString(digest[:])
}

func handleIssue() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req issueRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"message": "invalid request body"})
			return
		}

		keypair, err := getBitmarkKeypair(req.Registrant)
		if err != nil {
			c.JSON(400, gin.H{"message": "user not registered"})
			return
		}

		content, path, fingerprint, err := readAsset(req.AssetURL)
		if err != nil {
			c.JSON(400, gin.H{"message": "unable to read asset"})
			return
		}

		bitmarkIds, err := bmservice.Issue(keypair, req.Name, fingerprint, req.Quantity)
		if err != nil {
			checkErr(c, err)
			return
		}

		for _, id := range bitmarkIds {
			err := bmservice.Upload(keypair, id, path, content)
			if err != nil {
				checkErr(c, err)
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"bitmark_ids": bitmarkIds})
	}
}
