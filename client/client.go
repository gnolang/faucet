package client

import (
	coreTypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Client defines the TM2 client functionality
type Client interface {
	// Account methods //

	// GetAccount fetches the account if it has been initialized
	GetAccount(address crypto.Address) (std.Account, error)

	// Transaction methods //

	// SendTransactionSync sends the specified transaction to the network,
	// and does not wait for it to be committed to the chain
	SendTransactionSync(tx *std.Tx) (*coreTypes.ResultBroadcastTx, error)

	// SendTransactionCommit sends the specified transaction to the network,
	// and wait for it to be committed to the chain
	SendTransactionCommit(tx *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error)

	// Ping calls the simplest element on the implementation that is able
	// to tell if it is still alive. If there is any problem, it returns an error.
	Ping() error
}
