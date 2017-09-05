package bmservice

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	bitmarklib "github.com/bitmark-inc/go-bitmarklib"
	"golang.org/x/crypto/ed25519"
)

type issueRequest struct {
	Assets []bitmarklib.Asset
	Issues []bitmarklib.Issue
}

type issueResponse []struct {
	TxId string `json:"txId"`
}

func IssueBitmark(asset bitmarklib.Asset, issues []bitmarklib.Issue) ([]string, error) {
	url := fmt.Sprintf("%s/v1/issue", cfg.core)
	body := issueRequest{
		Assets: []bitmarklib.Asset{asset},
		Issues: issues,
	}
	req, err := newJSONRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	var reply issueResponse
	if err := submitReqWithJSONResp(req, &reply); err != nil {
		return nil, err
	}

	bitmarkIds := make([]string, 0)
	for _, tx := range reply {
		bitmarkIds = append(bitmarkIds, tx.TxId)
	}
	return bitmarkIds, nil
}

func TransferBitmark(transfer *bitmarklib.Transfer) (string, error) {
	url := fmt.Sprintf("%s/v1/transfer", cfg.core)
	body := map[string]interface{}{
		"transfer": transfer,
	}
	req, err := newJSONRequest("POST", url, body)
	if err != nil {
		return "", err
	}

	var reply issueResponse
	if err := submitReqWithJSONResp(req, &reply); err != nil {
		return "", err
	}

	return reply[0].TxId, nil
}

func UpdateSessionData(k *bitmarklib.KeyPair, sessionData *bitmarklib.SessionData, bitmarkId, accountNo string) error {
	r, _ := json.Marshal(sessionData)
	data := string(r)
	sig := hex.EncodeToString(ed25519.Sign(k.PrivateKeyBytes(), []byte(data)))

	url := fmt.Sprintf("%s/v1/session/%s?account_no=%s", cfg.core, bitmarkId, accountNo)
	body := map[string]interface{}{
		"data":      data,
		"signature": sig,
	}
	fmt.Printf("%v", body)
	req, err := newJSONRequest("PUT", url, body)
	if err != nil {
		return err
	}

	if err := submitReqWithJSONResp(req, nil); err != nil {
		return err
	}

	return nil
}

func GetSessionData(accountNo string, bitmarkId string) (*bitmarklib.SessionData, error) {
	url := fmt.Sprintf("%s/v1/session/%s?account_no=%s", cfg.core, bitmarkId, accountNo)
	req, err := newJSONRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var reply bitmarklib.SessionData
	if err := submitReqWithJSONResp(req, &reply); err != nil {
		return nil, err
	}

	return &reply, nil
}
