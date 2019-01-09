package main

import (
	"github.com/bitmark-inc/bitmark-sdk-go/account"
)

func getEncrKey(acct account.Account) account.EncrKey {
	switch v := acct.(type) {
	case *account.AccountV1:
		return v.EncrKey
	case *account.AccountV2:
		return v.EncrKey
	default:
		return nil
	}
}
