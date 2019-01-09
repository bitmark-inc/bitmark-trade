package tx

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
)

func Get(txId string, loadAsset bool) (*Tx, error) {
	client := sdk.GetAPIClient()

	vals := url.Values{}
	vals.Set("pending", "true")
	if loadAsset {
		vals.Set("asset", "true")
	}

	req, err := client.NewRequest("GET", fmt.Sprintf("/v3/txs/%s?%s", txId, vals.Encode()), nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Tx    *Tx          `json:"tx"`
		Asset *asset.Asset `json:"asset"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, err
	}

	result.Tx.Asset = result.Asset

	return result.Tx, nil
}

func List(builder *QueryParamsBuilder) ([]*Tx, error) {
	txs := make([]*Tx, 0)
	it := NewIterator(builder)
	for it.Before() {
		for _, t := range it.Values() {
			txs = append(txs, t)
		}
	}
	if it.Err() != nil {
		return nil, it.Err()
	}

	return txs, nil
}

type APIResultIterator struct {
	builder *QueryParamsBuilder
	current int
	data    []*Tx
	err     error
}

type Iterator interface {
	Before() bool
	After() bool
	Values() []*Tx
	Err() error
}

func NewIterator(builder *QueryParamsBuilder) Iterator {
	return &APIResultIterator{
		builder: builder,
	}
}

func (i *APIResultIterator) Before() bool {
	i.builder.params.Set("to", "earlier")
	return i.next()
}

func (i *APIResultIterator) After() bool {
	i.builder.params.Set("to", "later")
	return i.next()
}

func (i *APIResultIterator) next() bool {
	if i.current > 0 {
		i.builder.params.Set("at", strconv.Itoa(i.current))
	}

	params, err := i.builder.Build()
	if err != nil {
		i.err = err
		return false
	}

	client := sdk.GetAPIClient()
	req, err := client.NewRequest("GET", "/v3/txs?"+params, nil)
	if err != nil {
		i.err = err
		return false
	}

	var result struct {
		Txs    []*Tx          `json:"txs"`
		Assets []*asset.Asset `json:"assets"`
	}
	if err := client.Do(req, &result); err != nil {
		i.err = err
		return false
	}

	if len(result.Txs) > 0 {
		if len(result.Assets) > 0 { // load assets
			assets := make(map[string]*asset.Asset)
			for _, a := range result.Assets {
				assets[a.Id] = a
			}

			for _, b := range result.Txs {
				b.Asset = assets[b.AssetId]
			}
		}

		i.current = result.Txs[len(result.Txs)-1].Sequence
		i.data = result.Txs
		return true
	}

	return false
}

func (i *APIResultIterator) Values() []*Tx {
	return i.data
}

func (i *APIResultIterator) Err() error {
	return i.err
}

type QueryParamsBuilder struct {
	params url.Values
	err    error
}

func NewQueryParamsBuilder() *QueryParamsBuilder {
	return &QueryParamsBuilder{params: url.Values{}}
}

func (ub *QueryParamsBuilder) OwnedBy(owner string, transient bool) *QueryParamsBuilder {
	ub.params.Set("owner", owner)
	ub.params.Set("sent", strconv.FormatBool(transient))
	return ub
}

func (ub *QueryParamsBuilder) ReferencedBitmark(bitmarkId string) *QueryParamsBuilder {
	ub.params.Set("bitmark_id", bitmarkId)
	return ub
}

func (ub *QueryParamsBuilder) ReferencedBlockNumber(blockNumber int64) *QueryParamsBuilder {
	ub.params.Set("block_number", fmt.Sprintf("%d", blockNumber))
	return ub
}

func (ub *QueryParamsBuilder) ReferencedAsset(assetId string) *QueryParamsBuilder {
	ub.params.Set("asset_id", assetId)
	return ub
}

func (ub *QueryParamsBuilder) LoadAsset(load bool) *QueryParamsBuilder {
	ub.params.Set("asset", strconv.FormatBool(load))
	return ub
}

func (ub *QueryParamsBuilder) Limit(size int) *QueryParamsBuilder {
	if size > 100 {
		ub.err = errors.New("invalid size: max = 100")
	}
	ub.params.Set("limit", strconv.Itoa(size))
	return ub
}

func (ub *QueryParamsBuilder) Build() (string, error) {
	if ub.err != nil {
		return "", ub.err
	}

	return ub.params.Encode(), nil
}
