package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/bitmark-inc/bitmark-trade/bmservice"
	"github.com/bitmark-inc/logger"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/hcl"
)

var (
	testnet bool
	cfg     *config
	db      *bolt.DB
	log     *logger.L
)

type config struct {
	Chain   string `hcl:"chain"`
	Port    int    `hcl:"port"`
	DataDir string `hcl:"datadir"`
}

func init() {
	var confpath string
	flag.StringVar(&confpath, "conf", "", "Specify configuration file")
	flag.Parse()

	cfg = readConfig(confpath)

	db = openDB(fmt.Sprintf("%s/bitmark-trade.db", cfg.DataDir))

	testnet = cfg.Chain != "live"

	if err := logger.Initialise(logger.Configuration{
		Directory: cfg.DataDir,
		File:      "trade.log",
		Size:      1048576,
		Count:     10,
		Levels:    map[string]string{"DEFAULT": "info"},
	}); err != nil {
		panic(fmt.Sprintf("logger initialization failed: %s", err))
	}

	bmservice.Init(cfg.Chain)

	log = logger.New("")
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
	db, err := bolt.Open(dbpath, 0660, nil)
	if err != nil {
		panic(fmt.Sprintf("unable to init the databse: %v", err))
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(getAccountBucketName()))
		if err != nil {
			panic(fmt.Sprintf("unable to init the databse: %v", err))
		}

		return nil
	})

	return db
}

func getAccountBucketName() string {
	bucketname := "account-testnet"
	if cfg.Chain == "live" {
		bucketname = "account-livenet"
	}

	return bucketname
}

func main() {
	r := gin.Default()
	r.POST("/account", handleAccountCreation())
	r.POST("/issue", handleIssue())
	r.POST("/transfer", handleTransfer())
	r.GET("/assets/:accountNo/:bitmarkId", handleAssetDownload())
	r.Run(fmt.Sprintf(":%d", cfg.Port))
}
