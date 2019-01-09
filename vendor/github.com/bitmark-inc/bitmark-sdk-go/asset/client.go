package asset

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
)

type registrationRequest struct {
	Assets []*RegistrationParams `json:"assets"`
}

type registeredItem struct {
	Id        string `json:"id"`
	Duplocate bool   `json:"duplicate"`
}

func Register(params *RegistrationParams) (string, error) {
	r := registrationRequest{
		Assets: []*RegistrationParams{params},
	}
	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(r)

	client := sdk.GetAPIClient()
	req, _ := client.NewRequest("POST", "/v3/register-asset", body)

	var result struct {
		Assets []registeredItem `json:"assets"`
	}
	if err := client.Do(req, &result); err != nil {
		return "", err
	}
	return result.Assets[0].Id, nil
}

func Get(assetId string) (*Asset, error) {
	client := sdk.GetAPIClient()

	req, err := client.NewRequest("GET", "/v3/assets/"+assetId+"?pending=true", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Asset *Asset `json:"asset"`
	}
	if err := client.Do(req, &result); err != nil {
		return nil, err
	}

	return result.Asset, nil
}

func List(builder *QueryParamsBuilder) ([]*Asset, error) {
	assets := make([]*Asset, 0)
	it := NewIterator(builder)
	for it.Before() {
		for _, a := range it.Values() {
			assets = append(assets, a)
		}
	}
	if it.Err() != nil {
		return nil, it.Err()
	}

	return assets, nil
}

type Iterator interface {
	Before() bool
	After() bool
	Values() []*Asset
	Err() error
}

type APIResultIterator struct {
	builder *QueryParamsBuilder
	current int
	data    []*Asset
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
	req, err := client.NewRequest("GET", "/v3/assets?"+params, nil)
	if err != nil {
		i.err = err
		return false
	}

	var result struct {
		Assets []*Asset `json:"assets"`
	}
	if err := client.Do(req, &result); err != nil {
		i.err = err
		return false
	}

	if len(result.Assets) > 0 {
		i.current = result.Assets[len(result.Assets)-1].Sequence
		i.data = result.Assets
		return true
	}

	return false
}

func (i *APIResultIterator) Values() []*Asset {
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

func (qb *QueryParamsBuilder) RegisteredBy(registrant string) *QueryParamsBuilder {
	qb.params.Set("registrant", registrant)
	return qb
}

func (qb *QueryParamsBuilder) Limit(size int) *QueryParamsBuilder {
	if size > 100 {
		qb.err = errors.New("invalid size: max = 100")
	}
	qb.params.Set("limit", strconv.Itoa(size))
	return qb
}

func (qb *QueryParamsBuilder) Build() (string, error) {
	if qb.err != nil {
		return "", qb.err
	}

	return qb.params.Encode(), nil
}
