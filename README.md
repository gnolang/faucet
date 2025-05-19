<h2 align="center">⚛️ Tendermint2 Faucet ⚛️</h2>

## Overview

`faucet` is a versatile command-line interface (CLI) tool and library designed to effortlessly deploy a faucet server
for Gno Tendermint 2 networks.

### Default endpoint (root)

The `faucet` adopts the [JSON-RPC 2.0 standard](https://www.jsonrpc.org/specification) for requests / responses.

By default, the `/` endpoint is the home of the `drip` method, to handle faucet drips. The first parameter is the
beneficiary address, and the second one is the string representation of the drip amount (`std.Coins`).

This can of course be overwritten with custom handling logic by the faucet creator (see below).

```json
{
  "jsonrpc": "2.0",
  "id": 0,
  "method": "drip",
  "params": [
    "g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2",
    "1000ugnot"
  ]
}
```

## Key Features

### Customizability

This faucet is highly customizable, allowing you to tailor it to the specific needs of your Tendermint 2 network. You
can easily configure token types, distribution limits, and other parameters to suit your requirements.

### Batch JSON Request Support

The faucet supports batch JSON requests, making it efficient for mass token distribution. You can submit multiple
requests in a single batch, reducing the overhead of individual requests.

### Extensibility

The faucet is designed with extensibility in mind. You can extend its functionality through middleware and custom
handlers. This means you can integrate additional features or connect it to external systems seamlessly, like a
rate limiting system.

## Getting Started

### As a binary

To get started with the Tendermint 2 Faucet, follow these steps:

1. Clone this repository to your local machine:

```bash
git clone github.com/gnolang/faucet
```

2. Build out the binary

To build out the binary, run the following command:

```bash
make build
```

3. Run the faucet

To run the faucet, start the built binary:

```bash
./build/faucet --mnemonic "<faucet_account_mnemonic>"
```

The provided mnemonic will be used to derive accounts which will be used to serve
funds to users. Make sure that the accounts derived from it are well funded.

It should print something like the following. (Note the port number at the end.)

```
2024-01-11T12:47:27.826+0100	INFO	cmd/logger.go:17	faucet started at [::]:8545
```

4. To send coins to a single account, in a new terminal enter the following (change to the correct recipient address):

```bash
curl --location --request POST 'http://localhost:8545' --header 'Content-Type: application/json' --data '{ "jsonrpc": "2.0", "id": 0, "method": "drip", "params": [ "g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2", "1000ugnot" ] }'
```

5. To ensure the faucet is listening to requests, you can ping the health endpoint (returns a simple status 200):

```bash
curl --location --request GET 'http://localhost:8545/health'
```

### As a library

To add `faucet` to your Go project, simply run:

```bash
go get github.com/gnolang/faucet
```

To use the faucet, you can set it up as such:

```go
package main

import (
	// ...
	"context"

	"github.com/gnolang/faucet/client/http"
	"github.com/gnolang/faucet/estimate/static"
)

func main() {
	// Create the faucet
	f, err := NewFaucet(
		static.New(...), // gas estimator
		http.NewClient(...), // remote address 
	)

	// The faucet is controlled through a top-level context
	ctx, cancelFn := context.WithCancel(context.Background())

	// Start the faucet
	go f.Serve(ctx)

	// Close the faucet
	cancelFn()
}

```

## What kind of extensibility?

### Middleware

Middleware functions can be added to extend the faucet's functionality. For example, you can add middleware to
authenticate users, log requests, or implement rate limiting.

### Custom Handlers

Custom request handlers can be created to handle specific actions or integrate with external systems. For instance, you
can create a custom handler to trigger additional actions when tokens are distributed, such as sending notifications.