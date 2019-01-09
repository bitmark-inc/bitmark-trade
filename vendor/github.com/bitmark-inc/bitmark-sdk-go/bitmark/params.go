package bitmark

import (
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/bitmark-inc/bitmark-sdk-go/account"
	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/utils"
)

type OfferResponseAction string

const Accept OfferResponseAction = "accept"
const Reject OfferResponseAction = "reject"
const Cancel OfferResponseAction = "cancel"

var nonceIndex uint64

type QuantityOptions struct {
	Nonces   []uint64
	Quantity int
}

type IssuanceParams struct {
	Issuances []*IssueRequest `json:"issues"`
}

type IssueRequest struct {
	AssetId   string `json:"asset_id" pack:"hex64"`
	Owner     string `json:"owner" pack:"account"`
	Nonce     uint64 `json:"nonce" pack:"uint64"`
	Signature string `json:"signature"`
}

func NewIssuanceParams(assetId string, quantity int) *IssuanceParams {
	ip := &IssuanceParams{
		Issuances: make([]*IssueRequest, 0),
	}

	a, _ := asset.Get(assetId)
	if a == nil || (a != nil && a.Status != "confirmed") {
		issuance := &IssueRequest{
			AssetId: assetId,
			Nonce:   0,
		}
		ip.Issuances = append(ip.Issuances, issuance)

		quantity -= 1
	}

	for i := 0; i < quantity; i++ {
		atomic.AddUint64(&nonceIndex, 1)
		nonce := uint64(time.Now().UTC().Unix())*1000 + nonceIndex%1000
		issuance := &IssueRequest{
			AssetId: assetId,
			Nonce:   nonce,
		}
		ip.Issuances = append(ip.Issuances, issuance)
	}

	return ip
}

// Sign all issunaces in a batch
func (p *IssuanceParams) Sign(issuer account.Account) error {
	for _, issuance := range p.Issuances {
		issuance.Owner = issuer.AccountNumber()
		message, err := utils.Pack(issuance)
		if err != nil {
			return err
		}
		issuance.Signature = hex.EncodeToString(issuer.Sign(message))
	}

	return nil
}

type TransferParams struct {
	Transfer *TransferRequest `json:"transfer"`
}

type TransferRequest struct {
	Link                    string   `json:"link" pack:"hex32"`
	Escrow                  *payment `json:"-" pack:"payment"` // optional escrow payment address
	Owner                   string   `json:"owner" pack:"account"`
	Signature               string   `json:"signature"`
	requireCountersignature bool
}

type payment struct {
	Currency string `json:"currency"`
	Address  string `json:"address"`
	Amount   uint64 `json:"amount,string"`
}

func NewTransferParams(receiver string) *TransferParams {
	return &TransferParams{
		Transfer: &TransferRequest{
			Owner: receiver,
			requireCountersignature: false,
		},
	}
}

// FromBitmark sets link asynchronously
func (t *TransferParams) FromBitmark(bitmarkId string) error {
	bitmark, err := Get(bitmarkId, false)
	if err != nil {
		return err
	}

	t.Transfer.Link = bitmark.LatestTxId
	return nil
}

// FromLatestTx sets link synchronously
func (t *TransferParams) FromLatestTx(txId string) {
	t.Transfer.Link = txId
}

func (t *TransferParams) Sign(sender account.Account) error {
	message, err := utils.Pack(t.Transfer)
	if err != nil {
		return err
	}
	t.Transfer.Signature = hex.EncodeToString(sender.Sign(message))
	return nil
}

type OfferParams struct {
	Offer struct {
		Transfer  *TransferRequest       `json:"record"`
		ExtraInfo map[string]interface{} `json:"extra_info"`
	} `json:"offer"`
}

func NewOfferParams(receiver string, info map[string]interface{}) *OfferParams {
	return &OfferParams{
		Offer: struct {
			Transfer  *TransferRequest       `json:"record"`
			ExtraInfo map[string]interface{} `json:"extra_info"`
		}{
			Transfer: &TransferRequest{
				Owner: receiver,
				requireCountersignature: true,
			},
			ExtraInfo: info,
		},
	}
}

// FromBitmark sets link asynchronously
func (o *OfferParams) FromBitmark(bitmarkId string) error {
	bitmark, err := Get(bitmarkId, false)
	if err != nil {
		return err
	}

	o.Offer.Transfer.Link = bitmark.LatestTxId
	return nil
}

// FromLatestTx sets link synchronously
func (o *OfferParams) FromLatestTx(txId string) {
	o.Offer.Transfer.Link = txId
}

func (o *OfferParams) Sign(sender account.Account) error {
	message, err := utils.Pack(o.Offer.Transfer)
	if err != nil {
		return err
	}
	o.Offer.Transfer.Signature = hex.EncodeToString(sender.Sign(message))
	return nil
}

type CountersignedTransferRequest struct {
	Link             string   `json:"link" pack:"hex32"`
	Escrow           *payment `json:"-" pack:"payment"` // optional escrow payment address
	Owner            string   `json:"owner" pack:"account"`
	Signature        string   `json:"signature" pack:"hex64"`
	Countersignature string   `json:"countersignature"`
}

type ResponseParams struct {
	Id               string              `json:"id"`
	Action           OfferResponseAction `json:"action"`
	Countersignature string              `json:"countersignature"`
	auth             http.Header
	record           *CountersignedTransferRequest
}

func NewTransferResponseParams(bitmark *Bitmark, action OfferResponseAction) *ResponseParams {
	return &ResponseParams{
		Id:     bitmark.Offer.Id,
		Action: action,
		auth:   make(http.Header),
		record: bitmark.Offer.Record,
	}
}

func (r *ResponseParams) Sign(acct account.Account) error {
	ts := strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
	parts := []string{
		"updateOffer",
		r.Id,
		acct.AccountNumber(),
		ts,
	}
	message := strings.Join(parts, "|")
	sig := hex.EncodeToString(acct.Sign([]byte(message)))

	r.auth.Add("requester", acct.AccountNumber())
	r.auth.Add("timestamp", ts)
	r.auth.Add("signature", sig)

	if r.Action == Accept {
		message, err := utils.Pack(r.record)
		if err != nil {
			return err
		}
		r.Countersignature = hex.EncodeToString(acct.Sign(message))
	}
	return nil
}
