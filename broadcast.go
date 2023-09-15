package faucet

import (
	"fmt"

	"github.com/gnolang/faucet/client"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// broadcastTransaction broadcasts the transaction using a COMMIT send
func broadcastTransaction(client client.Client, tx *std.Tx) error {
	// Send the transaction.
	// NOTE: Commit sends are temporary. Once
	// there is support for event indexing, this
	// call will change to a sync send
	response, err := client.SendTransactionCommit(tx)
	if err != nil {
		return fmt.Errorf("unable to send transaction, %w", err)
	}

	// Check the errors
	if response.CheckTx.IsErr() {
		return fmt.Errorf("transaction failed initial validation, %w", response.CheckTx.Error)
	}

	if response.DeliverTx.IsErr() {
		return fmt.Errorf("transaction failed during execution, %w", response.DeliverTx.Error)
	}

	return nil
}
