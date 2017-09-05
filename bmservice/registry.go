package bmservice

import "fmt"

type holder struct {
	TxId   string `json:"tx_id"`
	Owner  string `json:"owner"`
	Status string `json:"status"`
}

type bitmark struct {
	HeadId     string   `json:"head_id"`
	Issuer     string   `json:"issuer"`
	Id         string   `json:"id"`
	Owner      string   `json:"owner"`
	AssetId    string   `json:"asset_id"`
	Status     string   `json:"status"`
	Provenance []holder `json:"provenance"`
}

func (b *bitmark) PreviousOwner() string {
	if b.HeadId == b.Id {
		return b.Provenance[0].Owner
	}

	return b.Provenance[1].Owner
}

type tx struct {
	Tx struct {
		Owner     string `json:"owner"`
		AssetId   string `json:"asset_id"`
		Status    string `json:"status"`
		BitmarkId string `json:"bitmark_id"`
	} `json:"tx"`
}

func GetBitmark(bitmarkId string) (*bitmark, error) {
	url := fmt.Sprintf("%s/v1/bitmarks/%s?provenance=true&pending=true", cfg.core, bitmarkId)
	req, err := newJSONRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var reply struct {
		Bitmark bitmark `json:"bitmark"`
	}
	if err := submitReqWithJSONResp(req, &reply); err != nil {
		return nil, err
	}

	return &reply.Bitmark, nil
}

func GetTx(txId string) (*tx, error) {
	url := fmt.Sprintf("%s/v1/txs/%s", cfg.core, txId)
	req, err := newJSONRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	var reply tx
	if err := submitReqWithJSONResp(req, &reply); err != nil {
		return nil, err
	}

	return &reply, nil
}
