package bmservice

import (
	"fmt"

	bitmarklib "github.com/bitmark-inc/go-bitmarklib"
)

type issueRequest struct {
	Assets []bitmarklib.Asset
	Issues []bitmarklib.Issue
}

type issueResponse []struct {
	TxId string `json:"txId"`
}

func Issue(kp *bitmarklib.KeyPair, name, fingerprint string, metadata map[string]string, quantity int) ([]string, error) {
	asset := bitmarklib.NewAsset(name, fingerprint)
	asset.SetMeta(metadata)
	if err := asset.Sign(kp); err != nil {
		return nil, err
	}

	issues := make([]bitmarklib.Issue, quantity)
	for i := 0; i < quantity; i++ {
		issue := bitmarklib.NewIssue(asset.AssetIndex())
		if err := issue.Sign(kp); err != nil {
			// TODO: could we ignore this?
			continue
		}
		issues[i] = issue
	}

	url := fmt.Sprintf("%s/v1/issue", cfg.gateway)
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

func Transfer(kp *bitmarklib.KeyPair, txId, owner string) (string, error) {
	transfer, err := bitmarklib.NewTransfer(txId, owner, isTestChain)
	if err != nil {
		return "", err
	}

	if err = transfer.Sign(kp); err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v1/transfer", cfg.gateway)
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
