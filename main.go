package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	bmksdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/logger"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/hcl"
)

var (
	service *Service
	db      *bolt.DB
	log     *logger.L
)

type config struct {
	Chain    string `hcl:"chain"`
	Port     int    `hcl:"port"`
	DataDir  string `hcl:"datadir"`
	APIToken string `hcl:"api_token"`
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
		_, err := tx.CreateBucketIfNotExists(getAccountBucketName())
		if err != nil {
			panic(fmt.Sprintf("unable to init the databse: %v", err))
		}

		return nil
	})

	return db
}

func main() {
	var confpath string
	flag.StringVar(&confpath, "conf", "", "Specify configuration file")
	flag.Parse()

	cfg := readConfig(confpath)

	var sdkcfg bmksdk.Config
	switch cfg.Chain {
	case "test":
		sdkcfg = bmksdk.Config{
			HTTPClient: http.DefaultClient,
			Network:    bmksdk.Testnet,
			APIToken:   cfg.APIToken,
		}
		service = &Service{
			http.DefaultClient,
			"https://api.test.bitmark.com",
			"https://key.test.bitmarkaccountassets.com",
		}
	case "live":
		sdkcfg = bmksdk.Config{
			HTTPClient: http.DefaultClient,
			Network:    bmksdk.Livenet,
			APIToken:   cfg.APIToken,
		}
		service = &Service{
			http.DefaultClient,
			"https://api.bitmark.com",
			"https://key.bitmarkaccountassets.com",
		}
	}

	bmksdk.Init(&sdkcfg)

	db = openDB(fmt.Sprintf("%s/bitmark-trade.db", cfg.DataDir))

	if err := logger.Initialise(logger.Configuration{
		Directory: cfg.DataDir,
		File:      "trade.log",
		Size:      1048576,
		Count:     10,
		Levels:    map[string]string{"DEFAULT": "info"},
	}); err != nil {
		panic(fmt.Sprintf("logger initialization failed: %s", err))
	}

	log = logger.New("")

	r := gin.Default()
	r.POST("/account", createAccount())
	r.POST("/issue", issueBitmarks())
	r.POST("/transfer", transferBitmark())
	r.GET("/assets/:accountNo/:bitmarkId", downloadAsset())
	r.Run(fmt.Sprintf(":%d", cfg.Port))
}
