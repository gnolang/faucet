package faucet

import (
	"errors"
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	coreTypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tm2Errors "github.com/gnolang/gno/tm2/pkg/bft/abci/example/errors"
)

func TestBroadcastTransaction(t *testing.T) {
	t.Parallel()

	t.Run("invalid broadcast", func(t *testing.T) {
		t.Parallel()

		var (
			sendErr    = errors.New("unable to send transaction")
			capturedTx *std.Tx

			mockClient = &mockClient{
				sendTransactionCommitFn: func(tx *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error) {
					capturedTx = tx

					return nil, sendErr
				},
			}
		)

		// Broadcast the transaction, and capture the error
		tx := &std.Tx{Memo: "dummy tx"}
		require.ErrorIs(t, broadcastTransaction(mockClient, tx), sendErr)

		// Make sure the correct transaction
		// broadcast was attempted
		assert.Equal(t, tx, capturedTx)
	})

	t.Run("initial tx validation error (CheckTx)", func(t *testing.T) {
		t.Parallel()

		var (
			capturedTx *std.Tx

			checkTxErr = tm2Errors.UnauthorizedError{}
			response   = &coreTypes.ResultBroadcastTxCommit{
				CheckTx: abci.ResponseCheckTx{
					ResponseBase: abci.ResponseBase{
						Error: checkTxErr,
					},
				},
			}

			mockClient = &mockClient{
				sendTransactionCommitFn: func(tx *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error) {
					capturedTx = tx

					return response, nil
				},
			}
		)

		// Broadcast the transaction, and capture the error
		tx := &std.Tx{Memo: "dummy tx"}
		require.ErrorIs(t, broadcastTransaction(mockClient, tx), checkTxErr)

		// Make sure the correct transaction
		// broadcast was attempted
		assert.Equal(t, tx, capturedTx)
	})

	t.Run("execute tx error (DeliverTx)", func(t *testing.T) {
		t.Parallel()

		var (
			capturedTx *std.Tx

			deliverTxErr = tm2Errors.BadNonceError{}
			response     = &coreTypes.ResultBroadcastTxCommit{
				DeliverTx: abci.ResponseDeliverTx{
					ResponseBase: abci.ResponseBase{
						Error: deliverTxErr,
					},
				},
			}

			mockClient = &mockClient{
				sendTransactionCommitFn: func(tx *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error) {
					capturedTx = tx

					return response, nil
				},
			}
		)

		// Broadcast the transaction, and capture the error
		tx := &std.Tx{Memo: "dummy tx"}
		require.ErrorIs(t, broadcastTransaction(mockClient, tx), deliverTxErr)

		// Make sure the correct transaction
		// broadcast was attempted
		assert.Equal(t, tx, capturedTx)
	})

	t.Run("valid broadcast", func(t *testing.T) {
		t.Parallel()

		var (
			capturedTx *std.Tx

			response = &coreTypes.ResultBroadcastTxCommit{
				DeliverTx: abci.ResponseDeliverTx{
					ResponseBase: abci.ResponseBase{
						Error: nil, // no error
					},
				},
				CheckTx: abci.ResponseCheckTx{
					ResponseBase: abci.ResponseBase{
						Error: nil, // no error
					},
				},
			}

			mockClient = &mockClient{
				sendTransactionCommitFn: func(tx *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error) {
					capturedTx = tx

					return response, nil
				},
			}
		)

		// Broadcast the transaction, and capture the error
		tx := &std.Tx{Memo: "dummy tx"}
		require.NoError(t, broadcastTransaction(mockClient, tx))

		// Make sure the correct transaction
		// broadcast was attempted
		assert.Equal(t, tx, capturedTx)
	})
}
