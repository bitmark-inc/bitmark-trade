package bitmarklib_test

import (
	"encoding/json"
	"fmt"
	"github.com/bitmark-inc/go-bitmarklib"
	"log"
)

func Example_createTransfer() {
	seed := "8VNLU6LSMjnCfMNHG9YftLV1TVWzAphfCSwJsf351974"
	kp, err := bitmarklib.NewKeyPairFromBase58Seed(seed, true, bitmarklib.ED25519)
	if err != nil {
		log.Fatal(err)
	}

	transfer, err := bitmarklib.NewTransfer("6776599a5fd4f2ade1ca87ee5fffd0295bb69b1969ffab1ec042a5f71ef74209", "fqN6WnjUaekfrqBvvmsjVskoqXnhJ632xJPHzdSgReC6bhZGuP", true)
	if err != nil {
		log.Fatal(err)
	}
	err = transfer.Sign(kp)
	if err != nil {
		log.Fatal(err)
	}

	b, err := json.Marshal(transfer)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
	// Output: {"link":"6776599a5fd4f2ade1ca87ee5fffd0295bb69b1969ffab1ec042a5f71ef74209","payment":null,"owner":"fqN6WnjUaekfrqBvvmsjVskoqXnhJ632xJPHzdSgReC6bhZGuP","signature":"e14642c9b9f0e1409ad2a0b9a248bf356ee9a9854a4c4f602d3de8daccf5cd11ee270c945c7d3bb58a3fc2858c109bdb9ad31914d2e1a1867383fdcd1f300f0e"}
}
