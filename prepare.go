package faucet

import (
	"github.com/gnolang/faucet/estimate"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// PrepareTxMessageFn is the callback method that
// constructs the faucet fund transaction message
type PrepareTxMessageFn func(PrepareCfg) std.Msg

// PrepareCfg specifies the tx prepare configuration
type PrepareCfg struct {
	SendAmount  std.Coins      // the amount to be sent
	FromAddress crypto.Address // the faucet address
	ToAddress   crypto.Address // the beneficiary address
}

// defaultPrepareTxMessage constructs the default
// native currency transfer message
func defaultPrepareTxMessage(cfg PrepareCfg) std.Msg {
	return bank.MsgSend{
		FromAddress: cfg.FromAddress,
		ToAddress:   cfg.ToAddress,
		Amount:      cfg.SendAmount,
	}
}

// prepareTransaction prepares the transaction for signing
func prepareTransaction(
	estimator estimate.Estimator,
	msg std.Msg,
) *std.Tx {
	// Construct the transaction
	tx := &std.Tx{
		Msgs:       []std.Msg{msg},
		Signatures: nil,
	}

	// Prepare the gas fee
	gasFee := estimator.EstimateGasFee()
	gasWanted := estimator.EstimateGasWanted(tx)

	tx.Fee = std.NewFee(gasWanted, gasFee)

	return tx
}
