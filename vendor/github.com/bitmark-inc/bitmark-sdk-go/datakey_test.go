package bitmarksdk

import (
	"bytes"
	"encoding/hex"
	"testing"
)

type dataKeyTestCase struct {
	algorithm  string
	key        string
	plaintext  string
	ciphertext string
}

var (
	dataKeyTestCases = []dataKeyTestCase{
		dataKeyTestCase{
			"chacha20poly1305",
			"0000000000000000000000000000000000000000000000000000000000000000",
			"Hello, world!",
			"d7628bd23a7d180df7c8fb1852c2cfc31d101d6a629b2c50edf6b9751a",
		},
	}
)

func mustDecodeString(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func TestEncryption(t *testing.T) {
	for _, c := range dataKeyTestCases {
		dataKey := ChaCha20DataKey{mustDecodeString(c.key)}
		ciphertext, err := dataKey.Encrypt([]byte(c.plaintext))
		if err != nil {
			t.Error(err)
		}

		if hex.EncodeToString(ciphertext) != c.ciphertext {
			t.Fail()
		}
	}
}

func TestDecryption(t *testing.T) {
	for _, c := range dataKeyTestCases {
		dataKey := ChaCha20DataKey{mustDecodeString(c.key)}
		plaintext, err := dataKey.Decrypt(mustDecodeString(c.ciphertext))
		if err != nil {
			t.Error(err)
		}

		if string(plaintext) != c.plaintext {
			t.Fail()
		}
	}
}

func TestSessionData(t *testing.T) {
	sender, _ := AccountFromSeed("5XEECttxvRBzxzAmuV4oh6T1FcQu4mBg8eWd9wKbf8hweXsfwtJ8sfH")
	recipient, _ := AccountFromSeed("5XEECscX3EQvpqMH59Es92uE9KXuuFRQ5pmZsQtyJFiqLEEi7CqSpCo")
	outlier, _ := AccountFromSeed("5XEECrT2vfQ29k5QtLT11Pyr9bDVnMRiyDEEVSQYiiDW8MhmzVZ9g2i")

	dataKey := &ChaCha20DataKey{mustDecodeString("0000000000000000000000000000000000000000000000000000000000000000")}

	// the sender creates the session data for the recipient
	data, _ := createSessionData(sender, dataKey, recipient.EncrKey.PublicKeyBytes())

	// the recipient CAN decrypt the data key
	restoredDataKey, _ := dataKeyFromSessionData(recipient, data, sender.EncrKey.PublicKeyBytes())
	if bytes.Compare(restoredDataKey.Bytes(), dataKey.Bytes()) != 0 ||
		restoredDataKey.Algorithm() != dataKey.Algorithm() {
		t.Fail()
	}

	// the outlier CANNOT decrypt the data key
	_, err := dataKeyFromSessionData(outlier, data, sender.EncrKey.PublicKeyBytes())
	if err == nil {
		t.Fail()
	}
}
