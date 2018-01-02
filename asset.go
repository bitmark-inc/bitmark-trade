package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func handleAssetDownload() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountNo := c.Param("accountNo")
		bitmarkId := c.Param("bitmarkId")

		acct, err := getAccount(accountNo)
		if err != nil {
			checkErr(c, err)
			return
		}

		name, content, err := bmk.DownloadAsset(acct, bitmarkId)
		if err != nil {
			checkErr(c, err)
			return
		}

		c.Header("Content-Disposition", "attachment; filename="+name)
		c.Header("Content-Length", strconv.Itoa(len(content)))
		c.Data(http.StatusOK, http.DetectContentType(content), content)
		return
	}
}
