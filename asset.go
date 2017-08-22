package main

import (
	"net/http"
	"strconv"

	"github.com/bitmark-inc/trade/bmservice"
	"github.com/gin-gonic/gin"
)

func handleAssetDownload() gin.HandlerFunc {
	return func(c *gin.Context) {
		account := c.Param("account")
		bitmarkId := c.Param("bitmarkId")

		kp, err := getBitmarkKeypair(account)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "user not exists"})
			return
		}

		filename, content, err := bmservice.Download(kp, bitmarkId)
		if err != nil {
			checkErr(c, err)
			return
		}
		filesize := strconv.Itoa(len(content))

		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Length", filesize)
		c.Data(http.StatusOK, http.DetectContentType(content), content)
		return
	}
}
