package faucet

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareTransaction(t *testing.T) {
	t.Parallel()

	var (
		fromAddress = crypto.Address{0}
		toAddress   = crypto.Address{1}
		sendAmount  = std.NewCoins(std.NewCoin("ugnot", 10))

		expectedGasFee    = std.NewCoin("gnot", 1)
		expectedGasWanted = int64(100)
		capturedTx        *std.Tx

		mockEstimator = &mockEstimator{
			estimateGasFeeFn: func() std.Coin {
				return expectedGasFee
			},

			estimateGasWantedFn: func(tx *std.Tx) int64 {
				capturedTx = tx

				return expectedGasWanted
			},
		}
	)

	// Prepare the transaction
	cfg := PrepareCfg{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		SendAmount:  sendAmount,
	}

	tx := prepareTransaction(mockEstimator, defaultPrepareTxMessage(cfg))

	// Make sure the transaction was created
	require.NotNil(t, tx)

	// Make sure the transaction is unsigned
	assert.Len(t, tx.Signatures, 0)

	// Make sure the transaction fee is correct
	expectedFee := std.NewFee(expectedGasWanted, expectedGasFee)
	assert.Equal(t, expectedFee, tx.Fee)

	// Make sure the correct transaction was estimated
	assert.Equal(t, tx, capturedTx)

	// Make sure there is one message
	require.Len(t, tx.Msgs, 1)

	msg := tx.Msgs[0]

	// Make sure the message is a send message
	msgSend, ok := msg.(bank.MsgSend)
	require.True(t, ok)

	// Make sure the message has valid fields
	assert.Equal(t, fromAddress, msgSend.FromAddress)
	assert.Equal(t, toAddress, msgSend.ToAddress)
	assert.Equal(t, sendAmount, msgSend.Amount)
}
