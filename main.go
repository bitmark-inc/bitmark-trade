package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/bitmark-inc/trade/bmservice"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/hcl"
)

var (
	testnet bool
	db      *bolt.DB
)

type config struct {
	Chain   string `hcl:"chain"`
	DataDir string `hcl:"datadir"`
}

func init() {
	var confpath string
	flag.StringVar(&confpath, "conf", "", "Specify configuration file")
	flag.Parse()

	cfg := readConfig(confpath)

	db = openDB(fmt.Sprintf("%s/bitmark-trade.db", cfg.DataDir))

	testnet = cfg.Chain != "production"

	bmservice.Init(cfg.Chain)
}

func readConfig(confpath string) *config {
	var cfg config

	dat, err := ioutil.ReadFile(confpath)
	if err != nil {
		panic(fmt.Sprintf("unable to read the configuration: %v", err))
	}

	if err = hcl.Unmarshal(dat, &cfg); nil != err {
		panic(fmt.Sprintf("unable to parse the configuration: %v", err))
	}

	return &cfg
}

func openDB(dbpath string) *bolt.DB {
	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		panic(fmt.Sprintf("unable to init the databse: %v", err))
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("account"))
		panic(fmt.Sprintf("unable to init the databse: %v", err))
	})

	return db
}

func main() {
	r := gin.Default()
	r.POST("/account", handleAccountCreation())
	r.POST("/issue", handleIssue())
	r.POST("/transfer", handleTransfer())
	r.GET("/assets/:account/:bitmarkId", handleAssetDownload())
	r.Run()
}
