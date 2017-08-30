package bitmarklib_test

import (
	"encoding/json"
	"fmt"
	"github.com/bitmark-inc/go-bitmarklib"
	"log"
)

func Example_createAsset() {
	seed := "8VNLU6LSMjnCfMNHG9YftLV1TVWzAphfCSwJsf351974"
	kp, err := bitmarklib.NewKeyPairFromBase58Seed(seed, true, bitmarklib.ED25519)
	if err != nil {
		log.Fatal(err)
	}

	asset := bitmarklib.NewAsset("testcase", "test_fingerprint")
	err = asset.Sign(kp)
	if err != nil {
		log.Fatal(err)
	}
	b, err := json.Marshal(asset)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	// Output: {"name":"testcase","fingerprint":"test_fingerprint","metadata":"","registrant":"e1m7c2amjuTYrRf18LyDHContyYo27Vw2PpeKdryWZmasZBnWU","signature":"68cbccfd732e494b827aadfa7546de73d4f0c3e69ce1d5ff81d100e2fdc05696969f193a024c9f77ae79900252304c84f0ebdffee2f01b117713a58afb006d06"}
}

func Example_createIssue() {
	seed := "8VNLU6LSMjnCfMNHG9YftLV1TVWzAphfCSwJsf351974"
	kp, err := bitmarklib.NewKeyPairFromBase58Seed(seed, true, bitmarklib.ED25519)
	if err != nil {
		log.Fatal(err)
	}

	asset := bitmarklib.NewAsset("testcase", "test_fingerprint")
	err = asset.Sign(kp)
	if err != nil {
		log.Fatal(err)
	}

	issue := bitmarklib.NewIssue(asset.AssetIndex())
	err = issue.Sign(kp)
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.Marshal(issue)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	// Output: {"asset":"e3edef121b7dd0e30f389716cae89a1589335d7d8aa9fe3802f8b19f19505077ff13a21f4c3374c1f188dd9fe13fcab39f2690e1fdae54473a850e8dc519e882","owner":"e1m7c2amjuTYrRf18LyDHContyYo27Vw2PpeKdryWZmasZBnWU","nonce":1499245158002,"signature":"d7cfb3f008000bffae32b98df6bacf8b5a068d4d29ebc114666fbf354be1c81725bcfb1124613e4dcfdc9674019b0723c4aebe32d4f3ed246dde19176c111308"}
}
