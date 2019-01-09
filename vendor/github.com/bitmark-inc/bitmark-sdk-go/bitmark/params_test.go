package bitmark

import (
	"encoding/json"
	"reflect"
	"testing"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
)

var (
	senderSeed   = "5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH"
	receiverSeed = "5XEECt4yuMK4xqBLr9ky5FBWpkAR6VHNZSz8fUzZDXPnN3D9MeivTSA"

	sender   account.Account
	receiver account.Account
)

func init() {
	sdk.Init(&sdk.Config{Network: sdk.Testnet})

	sender, _ = account.FromSeed(senderSeed)
	receiver, _ = account.FromSeed(receiverSeed)
}

// func TestIssunaceParams(t *testing.T) {
// 	assetId := "3c50d70e0fe78819e7755687003483523852ee6ecc59fe40a4e70e89496c4d45313c6d76141bc322ba56ad3f7cd9c906b951791208281ddba3ebb5e7ad83436c"
// 	params := NewIssuanceParams(assetId, QuantityOptions{Nonces: []uint64{1, 2, 3}})
// 	params.Sign(sender)

// 	expected := `
// 	{
// 		"issues": [
// 			{
// 				"asset_id": "3c50d70e0fe78819e7755687003483523852ee6ecc59fe40a4e70e89496c4d45313c6d76141bc322ba56ad3f7cd9c906b951791208281ddba3ebb5e7ad83436c",
// 				"owner": "e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog",
// 				"nonce": 1,
// 				"signature": "fc25264ad39a00f25bab7444923ba5322fcf9cb2e358ca2a4274d0d21a15fb3f1c0ead8381d49a1e3aecfc3f5dea5d3cc775b3ee8e77357d9f17c31143cd9705"
// 			},
// 			{
// 				"asset_id": "3c50d70e0fe78819e7755687003483523852ee6ecc59fe40a4e70e89496c4d45313c6d76141bc322ba56ad3f7cd9c906b951791208281ddba3ebb5e7ad83436c",
// 				"owner": "e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog",
// 				"nonce": 2,
// 				"signature": "34c703be52c94312549f3111e0004006fb722814d06cd1e1b961a95e344ff2f78439fc53a477e3460b9338b537f04dd2f1cabd7bb17fd9a655f4557772db5200"
// 			},
// 			{
// 				"asset_id": "3c50d70e0fe78819e7755687003483523852ee6ecc59fe40a4e70e89496c4d45313c6d76141bc322ba56ad3f7cd9c906b951791208281ddba3ebb5e7ad83436c",
// 				"owner": "e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog",
// 				"nonce": 3,
// 				"signature": "4c85a7c045b2f68f816d478709f3bde7aebde081d079fe905606758e79d4e9906e296836efbb8f04fa8e3ff5a0dda3794f3e4d7777ed0049a94ff6c1d4517600"
// 			}
// 		]
// 	}`

// 	if actual := toJSONString(params); !equalJSONString(actual, expected) {
// 		t.Fatalf("incorrect offer params: actual = %+v, expected = %+v", actual, expected)
// 	}
// }

func TestTransferParams(t *testing.T) {
	params := NewTransferParams(receiver.AccountNumber())
	params.FromLatestTx("67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80")
	params.Sign(sender)

	expected := `{
		"transfer": {
			"link":"67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80",
			"owner":"eZpG6Wi9SQvpDatEP7QGrx6nvzwd6s6R8DgMKgDbDY1R5bjzb9",
			"signature": "97dfcddd1f987bb313939b3f5a6ab7a893e1fcc45f111e0b893a075e42bab0f5db2b36a6de01511a0dfcd034397eb1d6c4a50194e67957dc9538766076bc2405"
		}
	}`

	if actual := toJSONString(params); !equalJSONString(actual, expected) {
		t.Fatalf("incorrect offer params: actual = %+v, expected = %+v", actual, expected)
	}
}

func TestOfferParams(t *testing.T) {
	params := NewOfferParams(receiver.AccountNumber(), nil)
	params.FromLatestTx("fa9bb80247dd0f6b3e3f21153f49fbb297b9568e67e298c96dbd75d3a348efeb")
	params.Sign(sender)

	expected := `{
		"offer": {
			"record": {
				"link":"fa9bb80247dd0f6b3e3f21153f49fbb297b9568e67e298c96dbd75d3a348efeb",
				"owner":"eZpG6Wi9SQvpDatEP7QGrx6nvzwd6s6R8DgMKgDbDY1R5bjzb9",
				"signature": "23fa2e96f534f57fe8bf088e4b0d88b60b6afa7df5cb7aca1d2bf6dd2ed68ee59e8d3281154417cc0993c19fa3dd8e982371bcc0709b2b00fb7d18566bf56603"
			},
			"extra_info": null
		}
	}`

	if actual := toJSONString(params); !equalJSONString(actual, expected) {
		t.Fatalf("incorrect offer params: actual = %+v, expected = %+v", actual, expected)
	}
}

func TestResponseParams(t *testing.T) {
	params := NewTransferResponseParams(&Bitmark{
		Offer: &TransferOffer{
			Id: "d205ed72-792f-43ca-885a-737949be6501",
			Record: &CountersignedTransferRequest{
				Link:      "d4abf04785b9a32ba1de395607e760aa3f59315ae9def35eed7b0c93e0e357f3",
				Owner:     "eZpG6Wi9SQvpDatEP7QGrx6nvzwd6s6R8DgMKgDbDY1R5bjzb9",
				Signature: "7aeeb5070e85a3701b271feaa9aff37e93eefb4081624605fc7da061c2c2008edce888404ed079c5d3d7e6d57044b3a68ad1a8c8ff57ed0558c51dbd715d2d09",
			},
		},
	}, "accept")
	params.Sign(receiver)

	expected := `{
		"id": "d205ed72-792f-43ca-885a-737949be6501",
		"action": "accept",
		"countersignature": "8e8d70d97afb92024fe3997a3b331515461931b52f515b90fb801b562eacfb158f60292976b6a1cf96e2b097fe4eebe133f64aedadaf2b986d312d9b5646190b"
	}`

	if actual := toJSONString(params); !equalJSONString(actual, expected) {
		t.Fatalf("incorrect response params: actual = %+v, expected = %+v", actual, expected)
	}
}

func toJSONString(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func equalJSONString(s1, s2 string) bool {
	r1 := make(map[string]interface{})
	json.Unmarshal([]byte(s1), &r1)

	r2 := make(map[string]interface{})
	json.Unmarshal([]byte(s2), &r2)

	return reflect.DeepEqual(r1, r2)
}
