package main

import (
	"net/http"

	bmksdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/gin-gonic/gin"
)

func checkErr(c *gin.Context, err error) {
	if err != nil {
		if serr, ok := err.(*bmksdk.ServiceError); ok {
			c.JSON(http.StatusServiceUnavailable, serr)
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": err.Error()})
		}
	}
}
