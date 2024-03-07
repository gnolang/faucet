package faucet

import (
	"errors"
	"testing"

	"github.com/gnolang/faucet/config"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	coreTypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFaucet_TransferFunds(t *testing.T) {
	t.Parallel()

	t.Run("unable to fetch accounts", func(t *testing.T) {
		t.Parallel()

		var (
			fetchErr = errors.New("unable to fetch account")
			amount   = std.NewCoins(std.NewCoin("ugnot", 1))

			mockClient = &mockClient{
				getAccountFn: func(_ crypto.Address) (std.Account, error) {
					return nil, fetchErr
				},
			}
			mockEstimator = &mockEstimator{
				estimateGasFeeFn: func() std.Coin {
					return std.NewCoin("ugnot", 0)
				},
			}
		)

		// Create faucet
		f, err := NewFaucet(
			mockEstimator,
			mockClient,
		)

		require.NoError(t, err)
		require.NotNil(t, f)

		// Attempt the transfer
		assert.ErrorIs(t, f.transferFunds(crypto.Address{}, amount), errNoFundedAccount)
	})

	t.Run("no funded accounts", func(t *testing.T) {
		t.Parallel()

		var (
			sendAmount     = std.NewCoins(std.NewCoin("ugnot", 10))
			accountBalance = std.NewCoins(std.NewCoin("ugnot", 5))

			mockClient = &mockClient{
				getAccountFn: func(_ crypto.Address) (std.Account, error) {
					return &mockAccount{
						getCoinsFn: func() std.Coins {
							return accountBalance // less than the send amount
						},
					}, nil
				},
			}
			mockEstimator = &mockEstimator{
				estimateGasFeeFn: func() std.Coin {
					return std.NewCoin("ugnot", 0)
				},
			}
		)

		// Create faucet
		cfg := config.DefaultConfig()
		cfg.MaxSendAmount = sendAmount.String()

		f, err := NewFaucet(
			mockEstimator,
			mockClient,
			WithConfig(cfg),
		)

		require.NoError(t, err)
		require.NotNil(t, f)

		// Attempt the transfer
		assert.ErrorIs(t, f.transferFunds(crypto.Address{}, sendAmount), errNoFundedAccount)
	})

	t.Run("unable to sign transaction", func(t *testing.T) {
		t.Parallel()

		var (
			sendAmount = std.NewCoins(std.NewCoin("ugnot", 10))

			signErr = errors.New("unable to sign transaction")

			mockClient = &mockClient{
				getAccountFn: func(_ crypto.Address) (std.Account, error) {
					return &mockAccount{
						getCoinsFn: func() std.Coins {
							return sendAmount // can cover the send amount
						},
					}, nil
				},
			}
			mockEstimator = &mockEstimator{
				estimateGasFeeFn: func() std.Coin {
					return std.NewCoin("ugnot", 0)
				},
			}
			mockPrivKey = &mockPrivKey{
				signFn: func(_ []byte) ([]byte, error) {
					return nil, signErr
				},
			}
			mockKeyring = &mockKeyring{
				getKeyFn: func(_ crypto.Address) crypto.PrivKey {
					return mockPrivKey
				},
				getAddressesFn: func() []crypto.Address {
					return []crypto.Address{
						{0}, // 1 account
					}
				},
			}
		)

		// Create faucet
		cfg := config.DefaultConfig()
		cfg.MaxSendAmount = sendAmount.String()

		f, err := NewFaucet(
			mockEstimator,
			mockClient,
			WithConfig(cfg),
		)

		// Set the keyring
		f.keyring = mockKeyring

		require.NoError(t, err)
		require.NotNil(t, f)

		// Attempt the transfer
		assert.ErrorIs(t, f.transferFunds(crypto.Address{}, sendAmount), signErr)
	})

	t.Run("valid asset transfer", func(t *testing.T) {
		t.Parallel()

		var (
			sendAmount = std.NewCoins(std.NewCoin("ugnot", 10))

			response = &coreTypes.ResultBroadcastTxCommit{
				CheckTx: abci.ResponseCheckTx{
					ResponseBase: abci.ResponseBase{
						Error: nil, // no error
					},
				},
				DeliverTx: abci.ResponseDeliverTx{
					ResponseBase: abci.ResponseBase{
						Error: nil, // no error
					},
				},
			}

			mockClient = &mockClient{
				getAccountFn: func(_ crypto.Address) (std.Account, error) {
					return &mockAccount{
						getCoinsFn: func() std.Coins {
							return sendAmount
						},
					}, nil
				},
				sendTransactionCommitFn: func(_ *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error) {
					return response, nil
				},
			}
			mockEstimator = &mockEstimator{
				estimateGasFeeFn: func() std.Coin {
					return std.NewCoin("ugnot", 0)
				},
			}
			mockPrivKey = &mockPrivKey{
				signFn: func(_ []byte) ([]byte, error) {
					return []byte("signature"), nil
				},
			}
			mockKeyring = &mockKeyring{
				getKeyFn: func(_ crypto.Address) crypto.PrivKey {
					return mockPrivKey
				},
				getAddressesFn: func() []crypto.Address {
					return []crypto.Address{
						{0}, // 1 account
					}
				},
			}
		)

		// Create faucet
		cfg := config.DefaultConfig()
		cfg.MaxSendAmount = sendAmount.String()

		f, err := NewFaucet(
			mockEstimator,
			mockClient,
			WithConfig(cfg),
		)

		f.keyring = mockKeyring

		require.NoError(t, err)
		require.NotNil(t, f)

		// Attempt the transfer
		assert.NoError(t, f.transferFunds(crypto.Address{}, sendAmount))
	})
}
