package bmservice

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	bitmarklib "github.com/bitmark-inc/go-bitmarklib"
	"golang.org/x/crypto/ed25519"
)

func obtainToken(kp *bitmarklib.KeyPair) (string, error) {
	ts := strconv.FormatInt(time.Now().UnixNano(), 10)
	sig := hex.EncodeToString(ed25519.Sign(kp.PrivateKeyBytes(), []byte(ts)))

	url := fmt.Sprintf("%s/s/api/token", cfg.storage)
	body := map[string]string{
		"account":   kp.Account().String(),
		"timestamp": ts,
		"signature": sig,
	}

	req, err := newJSONRequest("POST", url, body)
	if err != nil {
		return "", err
	}

	var reply map[string]string
	if err := submitReqWithJSONResp(req, &reply); err != nil {
		return "", err
	}

	return reply["token"], nil
}

// Upload asset
func Upload(kp *bitmarklib.KeyPair, bitmarkId, filename string, filecontent []byte) error {
	token, err := obtainToken(kp)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/s/assets/%s?token=%s", cfg.storage, bitmarkId, token)
	req, err := newFileUploadRequest(url, "file", filename, filecontent)
	if err != nil {
		return err
	}

	err = submitReqWithJSONResp(req, nil)
	if err != nil {
		return err
	}

	return nil
}

// Download asset
func Download(kp *bitmarklib.KeyPair, bitmarkId string) (string, []byte, error) {
	token, err := obtainToken(kp)
	if err != nil {
		return "", nil, err
	}

	url := fmt.Sprintf("%s/s/assets/%s?token=%s", cfg.storage, bitmarkId, token)
	return submitReqWithFileResp(url)
}
