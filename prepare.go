package faucet

import (
	"github.com/gnolang/faucet/estimate"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// prepareCfg specifies the tx prepare configuration
type prepareCfg struct {
	sendAmount  std.Coins      // the amount to be sent
	fromAddress crypto.Address // the faucet address
	toAddress   crypto.Address // the beneficiary address
}

// prepareTransaction prepares the transaction for signing
func prepareTransaction(
	estimator estimate.Estimator,
	cfg prepareCfg,
) *std.Tx {
	// Construct the transaction
	msg := vm.MsgCall{
		FromAddress: cfg.fromAddress,
		ToAddress:   cfg.toAddress,
		PkgAddr: "gno.land/r/demo/chess/register",
		Func: "RegisterPlayer",
		Args: []string{"calleraddr", "token"},
		Send:      cfg.sendAmount,
	}

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
