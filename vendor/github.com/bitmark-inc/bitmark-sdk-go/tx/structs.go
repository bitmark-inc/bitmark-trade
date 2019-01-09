package tx

import (
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
)

type Tx struct {
	Id          string `json:"id"`
	BitmarkId   string `json:"bitmark_id"`
	AssetId     string `json:"asset_id"`
	Asset       *asset.Asset
	Owner       string `json:"owner"`
	Status      string `json:"status"`
	BlockNumber int    `json:"block_number"`
	Sequence    int    `json:"offset"`
	PreviousId  string `json:"previous_id"`
}
