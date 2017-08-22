package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bitmark-inc/trade/bmservice"
	"github.com/gin-gonic/gin"
)

func checkErr(c *gin.Context, err error) {
	if err != nil {
		if serr, ok := err.(*bmservice.ServiceError); ok {
			var msg map[string]interface{}
			json.NewDecoder(strings.NewReader(err.Error())).Decode(&msg)
			c.JSON(serr.Status(), msg)
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"message": err.Error()})
		}
	}
}
