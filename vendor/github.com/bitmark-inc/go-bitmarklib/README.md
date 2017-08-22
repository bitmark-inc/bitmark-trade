# Bitmark Library in Go

 [![GoDoc](https://godoc.org/github.com/bitmark-inc/go-bitmarklib?status.svg)](https://godoc.org/github.com/bitmark-inc/go-bitmarklib)

This is a library for generating structs which is used to perform bitmark issuing and transfering in bitmark blockchain.

## Start using it

1. Download and install it:

    ```sh
    $ go get github.com/bitmark-inc/go-bitmarklib
    ```

2. Import it in your code:

    ```go
    import "github.com/bitmark-inc/go-bitmarklib"
    ```

## API Examples

### Keypair

Keypair is an identity. You can use it to sign assets, issues, and tranfer.

```go
func main() {
	kp, err := bitmarklib.NewKeyPair(true, bitmarklib.ED25519)
	if err != nil {
		log.Fatal(err)
	}
}
```

### Issue

Issue is to claim the ownership to a specific asset.

```go
func main() {
	asset := bitmarklib.NewAsset("test", "someFingerprint")

	err := asset.SetMeta(map[string]string{
		"description": "test stuff",
	})
	if err != nil {
		log.Fatal(err)
	}

	if err = asset.Sign(keypair); err != nil {
		log.Fatal(err)
	}

	issue := bitmarklib.NewIssue(asset.AssetIndex())
	if err := issue.Sign(keypair); err != nil {
		log.Fatal(err)
	}
}
```

### Transfer

Transfer is to transfer issues to another owner.

```go
func main() {
	keypair, err := bitmarklib.NewKeyPairFromBase58Seed(kkk)
	if err != nil {
		log.Fatal(err)
	}

	txId := "95eae4952002b80251a0e0793177161fe67fee26e6c81df06649540dad3999f2"
	newOwnerKey := "fa447039da1cb03c0e48ab48dec69769d3affce01a5565a4b64a5d920f3c21a9"

	transfer, err := bitmarklib.NewTransfer(txId, newOwnerKey)
	if err != nil {
		log.Fatal(err)
	}
	err = transfer.Sign(keypair.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
}
```
