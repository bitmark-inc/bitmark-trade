package bitmark

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
)

type txItem struct {
	TxId string `json:"txId"`
}

func Issue(params *IssuanceParams) ([]string, error) {
	client := sdk.GetAPIClient()

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return nil, err
	}

	req, err := client.NewRequest("POST", "/v3/issue", body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Bitmarks []struct {
			Id string `json:"id"`
		} `json:"bitmarks"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, err
	}

	bitmarkIds := make([]string, 0)
	for _, item := range result.Bitmarks {
		bitmarkIds = append(bitmarkIds, item.Id)
	}

	return bitmarkIds, nil
}

func Transfer(params *TransferParams) (string, error) {
	client := sdk.GetAPIClient()

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return "", err
	}

	req, err := client.NewRequest("POST", "/v3/transfer", body)
	if err != nil {
		return "", err
	}

	var result txItem
	if err := client.Do(req, &result); err != nil {
		return "", err
	}

	return result.TxId, nil
}

func Offer(params *OfferParams) error {
	client := sdk.GetAPIClient()

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return err
	}

	req, err := client.NewRequest("POST", "/v3/transfer", body)
	if err != nil {
		return err
	}

	err = client.Do(req, nil)
	return err
}

func Respond(params *ResponseParams) error {
	client := sdk.GetAPIClient()

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(params); err != nil {
		return err
	}

	req, err := client.NewRequest("PATCH", "/v3/transfer", body)
	if err != nil {
		return err
	}
	// TODO: set signaure beautifully
	for k, v := range params.auth {
		req.Header.Add(k, v[0])
	}

	err = client.Do(req, nil)
	return err
}

func Get(bitmarkId string, loadAsset bool) (*Bitmark, error) {
	client := sdk.GetAPIClient()

	vals := url.Values{}
	vals.Set("pending", "true")
	if loadAsset {
		vals.Set("asset", "true")
	}

	req, err := client.NewRequest("GET", fmt.Sprintf("/v3/bitmarks/%s?%s", bitmarkId, vals.Encode()), nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Bitmark *Bitmark     `json:"bitmark"`
		Asset   *asset.Asset `json:"asset"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, err
	}

	result.Bitmark.Asset = result.Asset

	return result.Bitmark, nil
}

func List(builder *QueryParamsBuilder) ([]*Bitmark, error) {
	bitmarks := make([]*Bitmark, 0)
	it := NewIterator(builder)
	for it.Before() {
		for _, b := range it.Values() {
			bitmarks = append(bitmarks, b)
		}
	}
	if it.Err() != nil {
		return nil, it.Err()
	}

	return bitmarks, nil
}

type Iterator interface {
	Before() bool
	After() bool
	Values() []*Bitmark
	Err() error
}

type APIResultIterator struct {
	builder *QueryParamsBuilder
	current int
	data    []*Bitmark
	err     error
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

	req, err := client.NewRequest("GET", "/v3/bitmarks?"+params, nil)
	if err != nil {
		i.err = err
		return false
	}

	var result struct {
		Bitmarks []*Bitmark     `json:"bitmarks"`
		Assets   []*asset.Asset `json:"assets"`
	}

	if err := client.Do(req, &result); err != nil {
		i.err = err
		return false
	}

	if len(result.Bitmarks) > 0 {
		if len(result.Assets) > 0 { // load assets
			assets := make(map[string]*asset.Asset)
			for _, a := range result.Assets {
				assets[a.Id] = a
			}

			for _, b := range result.Bitmarks {
				b.Asset = assets[b.AssetId]
			}
		}

		i.current = result.Bitmarks[len(result.Bitmarks)-1].Commit
		i.data = result.Bitmarks
		return true
	}

	return false
}

func (i *APIResultIterator) Values() []*Bitmark {
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

func (ub *QueryParamsBuilder) IssuedBy(issuer string) *QueryParamsBuilder {
	ub.params.Set("issuer", issuer)
	return ub
}

func (ub *QueryParamsBuilder) OwnedBy(owner string, transient bool) *QueryParamsBuilder {
	ub.params.Set("owner", owner)
	ub.params.Set("sent", strconv.FormatBool(transient))
	return ub
}

func (ub *QueryParamsBuilder) OfferFrom(sender string) *QueryParamsBuilder {
	ub.params.Set("offer_from", sender)
	return ub
}

func (ub *QueryParamsBuilder) OfferTo(receiver string) *QueryParamsBuilder {
	ub.params.Set("offer_to", receiver)
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
