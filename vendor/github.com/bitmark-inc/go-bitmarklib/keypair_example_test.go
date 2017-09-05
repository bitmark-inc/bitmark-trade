package bitmarklib_test

import (
	"fmt"
	"github.com/bitmark-inc/go-bitmarklib"
	"log"
)

func ExampleNewKeyPairFromBase58Seed() {
	seed := "8VNLU6LSMjnCfMNHG9YftLV1TVWzAphfCSwJsf351974"
	kp, err := bitmarklib.NewKeyPairFromBase58Seed(seed, true, bitmarklib.ED25519)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(kp)
	// Output: 3E2z4v1HjJUtxzh8CQeDN3NJ1THPs2bcuPiqJrEFNZLhw58gcWBuM1ZouCwwZZxBmZztkdEaRJrmLLoBpyb7dEoz
}

func ExampleNewKeyPairFromKIF() {
	kif := "cYK2SzQnYLG55yiRCSryymEw3EaNYnCD2mtCwkVXdFLSzQ4ReV"
	kp, err := bitmarklib.NewKeyPairFromKIF(kif)
	if err != nil {
		log.Fatal(err)
	}
	if !kp.PrivateKey.IsTesting() {
		log.Fatal(err)
	}
	fmt.Println(kp)
	// Output: 2T2LWi7M7Qz9vx3vaMiNCwzreSGEkBMkDcbMhq2Ss5LZTtWWAXUxVjk7N5Gg1guTW1XMHu4wrTXBik8EmVkPvZcK
}

func ExampleNewPubKeyFromAccount() {
	account := "fqN6WnjUaekfrqBvvmsjVskoqXnhJ632xJPHzdSgReC6bhZGuP"
	pubKey, err := bitmarklib.NewPubKeyFromAccount(account)
	if err != nil {
		log.Fatal(err)
	}
	if !pubKey.IsTesting() {
		log.Fatal(err)
	}
	fmt.Println(pubKey)
	// Output: fqN6WnjUaekfrqBvvmsjVskoqXnhJ632xJPHzdSgReC6bhZGuP
}
