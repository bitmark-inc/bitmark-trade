package main

import (
	"net/http"
	"net/url"

	bmksdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/gin-gonic/gin"
)

type issueRequest struct {
	AssetURL   string            `json:"asset_url"`
	Registrant string            `json:"registrant"`
	Name       string            `json:"name"`
	Metadata   map[string]string `json:"metadata"`
	Quantity   int               `json:"quantity"`
}

func handleIssue() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req issueRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"message": "invalid request body"})
			return
		}

		issuer, err := getAccount(req.Registrant)
		if err != nil {
			c.JSON(400, gin.H{"message": "owner not registered"})
			return
		}

		assaetURL, _ := url.Parse(req.AssetURL)
		af, err := bmksdk.NewAssetFile(assaetURL.Path, bmksdk.Private)
		if err != nil {
			c.JSON(400, gin.H{"message": "unable to read asset file"})
			return
		}
		af.Describe(req.Name, req.Metadata)
		log.Infof("asset_id: %s", af.Id())

		bitmarkIds, err := bmk.IssueByAssetFile(issuer, af, req.Quantity)
		if err != nil {
			checkErr(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"bitmark_ids": bitmarkIds})
	}
}
