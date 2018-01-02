package bitmarksdk

import (
	"strings"
	"testing"
)

type valid struct {
	seed   string
	phrase string
}

var (
	testnetData = valid{
		"5XEECt18HGBGNET1PpxLhy5CsCLG9jnmM6Q8QGF4U2yGb1DABXZsVeD",
		"accident syrup inquiry you clutch liquid fame upset joke glow best school repeat birth library combine access camera organ trial crazy jeans lizard science",
	}

	livenetData = valid{
		"5XEECqWqA47qWg86DR5HJ29HhbVqwigHUAhgiBMqFSBycbiwnbY639s",
		"ability panel leave spike mixture token voice certain today market grief crater cruise smart camera palm wheat rib swamp labor bid rifle piano glass",
	}
)

func TestTestnetAccount(t *testing.T) {
	acct1, _ := AccountFromSeed(testnetData.seed)
	if acct1.Network() != Testnet {
		t.Fail()
	}

	acct2, _ := AccountFromRecoveryPhrase(testnetData.phrase)
	if acct2.Network() != Testnet {
		t.Fail()
	}

	if strings.Join(acct1.RecoveryPhrase(), " ") != testnetData.phrase {
		t.Fail()
	}

	if acct2.Seed() != testnetData.seed {
		t.Fail()
	}

	if acct1.AccountNumber() != acct2.AccountNumber() {
		t.Fail()
	}
}

func TestLivenetAccount(t *testing.T) {
	acct1, _ := AccountFromSeed(livenetData.seed)
	if acct1.Network() != Livenet {
		t.Fail()
	}

	acct2, _ := AccountFromRecoveryPhrase(livenetData.phrase)
	if acct2.Network() != Livenet {
		t.Fail()
	}

	if strings.Join(acct1.RecoveryPhrase(), " ") != livenetData.phrase {
		t.Fail()
	}

	if acct2.Seed() != livenetData.seed {
		t.Fail()
	}

	if acct1.AccountNumber() != acct2.AccountNumber() {
		t.Fail()
	}
}
