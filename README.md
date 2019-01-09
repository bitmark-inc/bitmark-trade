# Bitmark Trade


## Installation

Make sure you have a working Go environment. [See the install instructions for Go.](https://golang.org/doc/install)

```shell
$ go get github.com/bitmark-inc/bitmark-trade
```

## How to run bitmark trade?

1. Create the configuration file.

```
# select the chain of the network
#chain=live
chain = "test"

# specify the port to run Bitmark trade service
port = 8080

# all logs and db files are relative to this directory
datadir = "/var/lib/bitmark"

# provide the token for using bitmark API; please contact us for applying the token
api_token = "12345678"
```

2. Run the following command to start the service.

```shell
$ bitmark-trade -conf=<config file path>
```

## Usage

Please refer to the [API document.](https://bitmarktradeservice.docs.apiary.io/#)
