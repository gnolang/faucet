package faucet

import (
	"errors"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignTransaction(t *testing.T) {
	t.Parallel()

	t.Run("valid signature", func(t *testing.T) {
		t.Parallel()

		var (
			chainID       = "gno"
			accountNumber = uint64(1)
			sequence      = uint64(0)

			capturedSignData []byte
			signature        = []byte("signature")

			mockPubKey = &mockPubKey{
				stringFn: func() string {
					return "public key"
				},
			}
			mockPrivKey = &mockPrivKey{
				signFn: func(signData []byte) ([]byte, error) {
					capturedSignData = signData

					return signature, nil
				},
				pubKeyFn: func() crypto.PubKey {
					return mockPubKey
				},
			}
		)

		// Create a dummy tx
		tx := &std.Tx{}
		expectedSignBytes, err := tx.GetSignBytes(chainID, accountNumber, sequence)
		require.NoError(t, err)

		cfg := signCfg{
			chainID:       chainID,
			accountNumber: accountNumber,
			sequence:      sequence,
		}

		// Sign the transaction
		require.NoError(t, signTransaction(tx, mockPrivKey, cfg))

		// Make sure the correct bytes were signed
		assert.Equal(t, expectedSignBytes, capturedSignData)

		// Make sure the signature was appended
		require.Len(t, tx.Signatures, 1)

		// Make sure the signature is valid
		sig := tx.Signatures[0]

		assert.Equal(t, signature, sig.Signature)
		assert.Equal(t, mockPubKey.String(), sig.PubKey.String())
	})

	t.Run("invalid signature", func(t *testing.T) {
		t.Parallel()

		var (
			chainID       = "gno"
			accountNumber = uint64(1)
			sequence      = uint64(0)

			capturedSignData []byte
			signErr          = errors.New("invalid sign data")

			mockPrivKey = &mockPrivKey{
				signFn: func(signData []byte) ([]byte, error) {
					capturedSignData = signData

					return nil, signErr
				},
			}
		)

		// Create a dummy tx
		tx := &std.Tx{}
		expectedSignBytes, err := tx.GetSignBytes(chainID, accountNumber, sequence)
		require.NoError(t, err)

		cfg := signCfg{
			chainID:       chainID,
			accountNumber: accountNumber,
			sequence:      sequence,
		}

		// Sign the transaction
		require.ErrorIs(t, signTransaction(tx, mockPrivKey, cfg), signErr)

		// Make sure the appropriate bytes were attempted
		// to be signed
		assert.Equal(t, expectedSignBytes, capturedSignData)

		// Make sure no signatures were appended
		assert.Len(t, tx.Signatures, 0)
	})
}
