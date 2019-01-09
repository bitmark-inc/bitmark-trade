package account

import (
	"strings"
	"testing"

	"golang.org/x/crypto/ed25519"

	sdk "github.com/bitmark-inc/bitmark-sdk-go"
	"golang.org/x/text/language"
)

type valid struct {
	seed          string
	phrases       []string
	accountNumber string
	network       sdk.Network
}

var (
	testnetAccount = valid{
		"9J87CAsHdFdoEu6N1unZk3sqhVBkVL8Z8",
		[]string{
			"name gaze apart lamp lift zone believe steak session laptop crowd hill",
			"箱 阻 起 歸 徹 矮 問 栽 瓜 鼓 支 樂",
		},
		"eMCcmw1SKoohNUf3LeioTFKaYNYfp2bzFYpjm3EddwxBSWYVCb",
		sdk.Testnet,
	}

	livenetAccount = valid{
		"9J87GaPq7FR9Uacdi3FUoWpP6LbEpo1Ax",
		[]string{
			"surprise mesh walk inject height join sound minor margin over jewel venue",
			"薯 托 劍 景 擔 額 牢 痛 亦 軟 凱 誼",
		},
		"aiKFA9dKkNHPys3nSZrLTPusoocPqXSFp5EexsgQ1hbYUrJVne",
		sdk.Livenet,
	}

	testnetDeprecatedAccount = valid{
		"5XEECt18HGBGNET1PpxLhy5CsCLG9jnmM6Q8QGF4U2yGb1DABXZsVeD",
		[]string{
			"accident syrup inquiry you clutch liquid fame upset joke glow best school repeat birth library combine access camera organ trial crazy jeans lizard science",
		},
		"ec6yMcJATX6gjNwvqp8rbc4jNEasoUgbfBBGGyV5NvoJ54NXva",
		sdk.Testnet,
	}

	livenetDeprecatedAccount = valid{
		"5XEECqWqA47qWg86DR5HJ29HhbVqwigHUAhgiBMqFSBycbiwnbY639s",
		[]string{
			"ability panel leave spike mixture token voice certain today market grief crater cruise smart camera palm wheat rib swamp labor bid rifle piano glass",
		},
		"bDnC8nCaupb1AQtNjBoLVrGmobdALpBewkyYRG7kk2euMG93Bf",
		sdk.Livenet,
	}

	langCheckSequence = []language.Tag{language.AmericanEnglish, language.TraditionalChinese}
)

func check(t *testing.T, a Account, data valid) {
	if a.Seed() != data.seed {
		t.Fatalf("invalid seed: expected = %s, actual = %s", testnetAccount.seed, a.Seed())
	}

	for i, p := range data.phrases {
		phrase, _ := a.RecoveryPhrase(langCheckSequence[i])
		if strings.Join(phrase, " ") != data.phrases[i] {
			t.Fatalf("invalid recovery phrase: expected = %s, actual = %s", p, phrase)
		}
	}

	if a.AccountNumber() != data.accountNumber {
		t.Fatalf("invalid account number: expected = %s, actual = %s", data.accountNumber, a.AccountNumber())
	}

	if a.Network() != data.network {
		t.Fatalf("invalid network: expected = %s, actual = %s", data.network, a.Network())
	}
}

func TestTestnetAccount(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Testnet})

	acctFromSeed, err := FromSeed(testnetAccount.seed)
	if err != nil {
		t.Fatalf("failed to recover from seed: %s", err)
	}
	check(t, acctFromSeed, testnetAccount)

	for i, lang := range langCheckSequence {
		phrase := strings.Split(testnetAccount.phrases[i], " ")
		acctFromPhrase, err := FromRecoveryPhrase(phrase, lang)
		if err != nil {
			t.Fatalf("failed to recover from phrase: %s", err)
		}
		check(t, acctFromPhrase, testnetAccount)
	}
}

func TestLivenetAccount(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Livenet})

	acctFromSeed, err := FromSeed(livenetAccount.seed)
	if err != nil {
		t.Fatalf("failed to recover from seed: %s", err)
	}
	check(t, acctFromSeed, livenetAccount)

	for i, lang := range langCheckSequence {
		phrase := strings.Split(livenetAccount.phrases[i], " ")
		acctFromPhrase, err := FromRecoveryPhrase(phrase, lang)
		if err != nil {
			t.Fatalf("failed to recover from phrase: %s", err)
		}
		check(t, acctFromPhrase, livenetAccount)
	}
}

func TestTestnetDeprecatedtTestnetAccount(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Testnet})

	acctFromSeed, err := FromSeed(testnetDeprecatedAccount.seed)
	if err != nil {
		t.Fatalf("failed to recover from seed: %s", err)
	}
	check(t, acctFromSeed, testnetDeprecatedAccount)

	for i, lang := range langCheckSequence {
		if i >= len(testnetDeprecatedAccount.phrases) {
			break
		}
		phrase := strings.Split(testnetDeprecatedAccount.phrases[i], " ")
		acctFromPhrase, err := FromRecoveryPhrase(phrase, lang)
		if err != nil {
			t.Fatalf("failed to recover from phrase: %s", err)
		}
		check(t, acctFromPhrase, testnetDeprecatedAccount)
	}
}

func TestTestnetDeprecatedtLivenetAccount(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Livenet})

	acctFromSeed, err := FromSeed(livenetDeprecatedAccount.seed)
	if err != nil {
		t.Fatalf("failed to recover from seed: %s", err)
	}
	check(t, acctFromSeed, livenetDeprecatedAccount)

	for i, lang := range langCheckSequence {
		if i >= len(livenetDeprecatedAccount.phrases) {
			break
		}
		phrase := strings.Split(livenetDeprecatedAccount.phrases[i], " ")
		acctFromPhrase, err := FromRecoveryPhrase(phrase, lang)
		if err != nil {
			t.Fatalf("failed to recover from phrase: %s", err)
		}
		check(t, acctFromPhrase, livenetDeprecatedAccount)
	}
}

func TestParseTestnetAccountNumber(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Testnet})

	acctFromSeed, err := FromSeed(testnetAccount.seed)
	if err != nil {
		t.Fatalf("failed to recover from seed: %s", err)
	}

	network, pubkey, err := ParseAccountNumber(acctFromSeed.AccountNumber())
	if err != nil {
		t.Fatal(err)
	}

	if network != sdk.Testnet {
		t.Fatal("wrong network")
	}

	sig := acctFromSeed.Sign([]byte("Hello, world!"))

	if !ed25519.Verify(pubkey, []byte("Hello, world!"), sig) {
		t.Fatal("wrong public key")
	}
}

func TestParseLivenetAccountNumber(t *testing.T) {
	sdk.Init(&sdk.Config{Network: sdk.Livenet})

	acctFromSeed, err := FromSeed(livenetAccount.seed)
	if err != nil {
		t.Fatalf("failed to recover from seed: %s", err)
	}

	network, pubkey, err := ParseAccountNumber(acctFromSeed.AccountNumber())
	if err != nil {
		t.Fatal(err)
	}

	if network != sdk.Livenet {
		t.Fatal("wrong network")
	}

	sig := acctFromSeed.Sign([]byte("Hello, world!"))

	if !ed25519.Verify(pubkey, []byte("Hello, world!"), sig) {
		t.Fatal("wrong public key")
	}
}
