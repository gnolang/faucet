package faucet

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// signCfg specifies the sign configuration
type signCfg struct {
	chainID       string // the ID of the chain
	accountNumber uint64 // the account number of the signer
	sequence      uint64 // the sequence of the signer
}

// signTransaction signs the specified transaction using
// the provided key and config
func signTransaction(tx *std.Tx, key crypto.PrivKey, cfg signCfg) error {
	// Get the sign bytes
	signBytes := tx.GetSignBytes(
		cfg.chainID,
		cfg.accountNumber,
		cfg.sequence,
	)

	// Sign the transaction
	signature, err := key.Sign(signBytes)
	if err != nil {
		return fmt.Errorf("unable to sign transaction, %w", err)
	}

	// Save the signature
	tx.Signatures = append(tx.Signatures, std.Signature{
		PubKey:    key.PubKey(),
		Signature: signature,
	})

	return nil
}
