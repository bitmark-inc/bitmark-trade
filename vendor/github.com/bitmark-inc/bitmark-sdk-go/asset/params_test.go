package asset

import (
	"encoding/json"
	"reflect"
	"testing"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"github.com/bitmark-inc/bitmark-sdk-go/account"
)

var (
	seed = "5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH"

	registrant account.Account
)

func init() {
	sdk.Init(&sdk.Config{Network: sdk.Testnet})

	registrant, _ = account.FromSeed(seed)
}

func TestRegistrantionParams(t *testing.T) {
	params, _ := NewRegistrationParams(
		"name",
		map[string]string{
			"k1": "v1",
			"k2": "v2",
		})
	params.SetFingerprint([]byte("hello world"))
	params.Sign(registrant)

	expected := `
	{
		"name": "name",
		"metadata": "k1\u0000v1\u0000k2\u0000v2",
		"fingerprint": "01840006653e9ac9e95117a15c915caab81662918e925de9e004f774ff82d7079a40d4d27b1b372657c61d46d470304c88c788b3a4527ad074d1dccbee5dbaa99a",
		"registrant": "e1pFRPqPhY2gpgJTpCiwXDnVeouY9EjHY6STtKwdN6Z4bp4sog",
		"signature": "dc9ad2f4948d5f5defaf9043098cd2f3c245b092f0d0c2fc9744fab1835cfb1ad533ee0ff2a72d1cdd7a69f8ba6e95013fc517d5d4a16ca1b0036b1f3055270c"
	}`

	if actual := toJSONString(params); !equalJSONString(actual, expected) {
		t.Fatalf("incorrect offer params: actual = %+v, expected = %+v", actual, expected)
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
