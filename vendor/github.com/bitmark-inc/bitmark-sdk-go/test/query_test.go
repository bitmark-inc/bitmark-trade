package test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/bitmark-inc/bitmark-sdk-go/asset"
	"github.com/bitmark-inc/bitmark-sdk-go/bitmark"
	"github.com/bitmark-inc/bitmark-sdk-go/tx"
)

func TestGetAsset(t *testing.T) {
	actual, err := asset.Get("2bc5189e77b55f8f671c62cb46650c3b0fa9f6219509427ea3f146de30d79d5598cdfab4ef754e19d1d8a0e4033d1e48adb92c0d83b74d00094c354f4948dc22")
	if err != nil {
		t.Error(err)
	}
	createdAt, _ := time.Parse("2006-01-02T15:04:05.000000Z", "2018-09-07T07:46:25.000000Z")
	expected := &asset.Asset{
		Id:   "2bc5189e77b55f8f671c62cb46650c3b0fa9f6219509427ea3f146de30d79d5598cdfab4ef754e19d1d8a0e4033d1e48adb92c0d83b74d00094c354f4948dc22",
		Name: "HA25124377",
		Metadata: map[string]string{
			"Source":     "Bitmark Health",
			"Saved Time": "2018-09-07T07:45:41.948Z",
		},
		Fingerprint: "016ef802c0f912ed69a5afc0e6c08fbe96de3284e7cc6e685111d5f1705049f20b695443bc2d7bae7fe2091d9e7a880a50a51c2d0be1963a99b9914f60f2462040",
		Registrant:  "eTicVBQqmGzxNMGiZGtKzDdufXZsiFKH3SR8FcVYM7MQTZ47k3",
		Status:      "confirmed",
		BlockNumber: 8696,
		Sequence:    8631,
		CreatedAt:   createdAt,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("incorrect asset record:\nactual=%+v\nexpected=%+v", actual, expected)
	}
}

func TestGetNonExsitingAsset(t *testing.T) {
	_, err := asset.Get("2bc5189e77b55f8f671c62cb46650c3b0fa9f6219509427ea3f146de30d79d55")
	if err.Error() != "[4000] asset not found" {
		t.Fatalf("incorrect error message")
	}
}

func TestListAsset(t *testing.T) {
	builder := asset.NewQueryParamsBuilder().RegisteredBy("epX3bZVM3g87BfNvbK5r4cizPX6Mkyvod4vLQFdDemZvWsxiGr").Limit(10)

	assets, err := asset.List(builder)
	if err != nil {
		t.Error(err)
	}

	for _, a := range assets {
		printBeautifulJSON(t, a)
	}
}

func TestListNonExsitingAssets(t *testing.T) {
	builder := asset.NewQueryParamsBuilder().RegisteredBy("epX3bZVM3g87BfNvbK5r4cizP").Limit(10)
	assets, _ := asset.List(builder)
	if len(assets) > 0 {
		t.Errorf("should return empty assets")
	}
}

func TestGetBitmark(t *testing.T) {
	bitmark, err := bitmark.Get("5b00a0395e1fa2ff4771f43d986efdae7847500bbe2736ca1823f7aa97ce8fef", true)
	if err != nil {
		t.Error(err)
	}
	printBeautifulJSON(t, bitmark)
}

func TestGetNonExsitingBitmark(t *testing.T) {
	_, err := bitmark.Get("2bc5189e77b55f8f671c62cb46650c3b", true)
	if err.Error() != "[4000] bitmark not found" {
		t.Fatalf("incorrect error message")
	}
}
func TestListBitmark(t *testing.T) {
	builder := bitmark.NewQueryParamsBuilder().
		IssuedBy("e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog").
		OwnedBy("eZpG6Wi9SQvpDatEP7QGrx6nvzwd6s6R8DgMKgDbDY1R5bjzb9", true).
		ReferencedAsset("1f21148a273b5e63773ceee976a84bcd014d88ac2c18a29cac4442120b430e158386b0ad90515c69e7d1fd6df8f3d523e3550741e88d0d04798627a57b0006c9").
		LoadAsset(true).
		Limit(10)

	bitmarks, err := bitmark.List(builder)
	if err != nil {
		t.Error(err)
	}

	for _, b := range bitmarks {
		printBeautifulJSON(t, b)
	}
}

func TestListNonExsitingBitmarks(t *testing.T) {
	builder := bitmark.NewQueryParamsBuilder().ReferencedAsset("1f21148a273b5e63773ceee976a84bcd014d88ac2c18a29cac4442120b430e15").Limit(10)

	bitmarks, err := bitmark.List(builder)
	if err != nil {
		t.Error(err)
	}

	if len(bitmarks) > 0 {
		t.Errorf("should return empty bitmarks")
	}
}
func TestGetTx(t *testing.T) {
	actual, err := tx.Get("67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80", false)
	if err != nil {
		t.Error(err)
	}

	expected := &tx.Tx{
		Id:          "67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80",
		BitmarkId:   "67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80",
		AssetId:     "3c50d70e0fe78819e7755687003483523852ee6ecc59fe40a4e70e89496c4d45313c6d76141bc322ba56ad3f7cd9c906b951791208281ddba3ebb5e7ad83436c",
		Asset:       nil,
		Owner:       "e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog",
		Status:      "confirmed",
		BlockNumber: 8668,
		Sequence:    728970,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("incorrect asset record:\nactual=%+v\nexpected=%+v", actual, expected)
	}
	printBeautifulJSON(t, actual)
}

func TestGetNonExsitingTx(t *testing.T) {
	_, err := tx.Get("67ef8bfee0ef7b8c33eda34ba21c8b2b", true)
	if err.Error() != "[4000] tx not found" {
		t.Fatalf("incorrect error message")
	}
}

func TestListProvenance(t *testing.T) {
	builder := tx.NewQueryParamsBuilder().
		ReferencedBitmark("67ef8bfee0ef7b8c33eda34ba21c8b2b0fbff601a7021984b2e27985251a0a80").
		LoadAsset(true).
		Limit(10)

	txs, err := tx.List(builder)
	if err != nil {
		t.Error(err)
	}

	for _, tx := range txs {
		printBeautifulJSON(t, tx)
	}
}

func TestListNonExsitingTxs(t *testing.T) {
	builder := tx.NewQueryParamsBuilder().ReferencedBitmark("2bc5189e77b55f8f671c62cb46650c3b").Limit(10)

	txs, err := tx.List(builder)
	if err != nil {
		t.Error(err)
	}

	if len(txs) > 0 {
		t.Errorf("should return empty txs")
	}
}

func printBeautifulJSON(t *testing.T, v interface{}) {
	item, _ := json.MarshalIndent(v, "", "\t")
	t.Log("\n" + string(item))
}
