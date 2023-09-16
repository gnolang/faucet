package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gnolang/faucet/cmd/root"
)

func main() {
	if err := root.New().ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)

		os.Exit(1)
	}
}
