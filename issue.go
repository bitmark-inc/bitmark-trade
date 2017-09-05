package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/crypto/sha3"

	"github.com/bitmark-inc/bitmark-trade/bmservice"
	bitmarklib "github.com/bitmark-inc/go-bitmarklib"
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

		registrant, err := getAccount(req.Registrant)
		if err != nil {
			c.JSON(400, gin.H{"message": "user not registered"})
			return
		}
		issuer := registrant

		content, path, fingerprint, err := readAsset(req.AssetURL)
		if err != nil {
			c.JSON(400, gin.H{"message": "unable to read asset"})
			return
		}

		asset := bitmarklib.NewAsset(req.Name, fingerprint)
		asset.SetMeta(req.Metadata)
		if sigerr := asset.Sign(registrant.AuthKeyPair); sigerr != nil {
			c.JSON(400, gin.H{"message": "unable to sign the asset"})
			return
		}

		issues := make([]bitmarklib.Issue, req.Quantity)
		for i := 0; i < req.Quantity; i++ {
			issue := bitmarklib.NewIssue(asset.AssetIndex())
			if sigerr := issue.Sign(registrant.AuthKeyPair); sigerr != nil {
				c.JSON(400, gin.H{"message": "unable to sign the issue"})
				return
			}
			issues[i] = issue
		}

		bitmarkIds, err := bmservice.IssueBitmark(asset, issues)
		if err != nil {
			checkErr(c, err)
			return
		}

		err = encryptAndUploadAsset(issuer, bitmarkIds, path, content)
		if err != nil {
			checkErr(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"bitmark_ids": bitmarkIds})
	}
}

func encryptAndUploadAsset(issuer *Account, bmIds []string, name string, content []byte) error {
	sessKey, err := bitmarklib.NewChaCha20SessionKey()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	errors := make(chan error)

	go func() {
		defer wg.Done()

		// encrypt the asset file
		encryptedAsset, err := bitmarklib.EncryptAssetFile(content, sessKey, issuer.AuthKeyPair.PrivateKeyBytes())
		if err != nil {
			errors <- err
			return
		}

		// upload the encrypted asset file
		if err := bmservice.Upload(issuer.AuthKeyPair, bmIds[0], name, encryptedAsset); err != nil {
			errors <- err
			return
		}
	}()

	go func() {
		defer wg.Done()

		// encrypt the session key and computes signatures
		sessData, err := bitmarklib.CreateSessionData(sessKey, issuer.EncrKeyPair.PublicKey, issuer.EncrKeyPair.PrivateKey, issuer.AuthKeyPair.PrivateKeyBytes())
		if err != nil {
			errors <- err
			return
		}

		// upload the session data
		for _, id := range bmIds {
			if err := bmservice.UpdateSessionData(issuer.AuthKeyPair, sessData, id, issuer.AccountNumber()); err != nil {
				errors <- err
				return
			}
		}
	}()

	wg.Wait()
	close(errors)

	var e string
	if len(errors) > 0 {
		for err := range errors {
			e += "; " + err.Error()
		}
		return fmt.Errorf(e)
	}

	return nil
}
