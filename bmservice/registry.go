package bmservice

import "fmt"

// type bitmark struct {
// 	Bitmark struct {
// 		Issuer  string `json:"issuer"`
// 		Owner   string `json:"owner"`
// 		AssetId string `json:"asset_id"`
// 		Status  string `json:"status"`
// 	} `json:"bitmark"`
// }

type tx struct {
	Tx struct {
		Owner   string `json:"owner"`
		AssetId string `json:"asset_id"`
		Status  string `json:"status"`
	} `json:"tx"`
}

// func GetBitmark(bitmarkId string) (*bitmark, error) {
// 	url := fmt.Sprintf("%s/v1/bitmarks/%s", cfg.registry, bitmarkId)
// 	req, err := newJSONRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var reply bitmark
// 	if err := submitReqWithJSONResp(req, &reply); err != nil {
// 		return nil, err
// 	}
//
// 	return &reply, nil
// }

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
