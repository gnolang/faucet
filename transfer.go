package faucet

import (
	"errors"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var errNoFundedAccount = errors.New("no funded account found")

// transferFunds transfers funds to the given address
func (f *Faucet) transferFunds(address crypto.Address) error {
	// Find an account that has balance to cover the transfer
	fundAccount, err := f.findFundedAccount()
	if err != nil {
		return err
	}

	// Prepare the transaction
	pCfg := PrepareCfg{
		FromAddress: fundAccount.GetAddress(),
		ToAddress:   address,
		SendAmount:  f.sendAmount,
	}
	tx := prepareTransaction(f.estimator, f.prepareTxMsgFn(pCfg))

	// Sign the transaction
	sCfg := signCfg{
		chainID:       f.config.ChainID,
		accountNumber: fundAccount.GetAccountNumber(),
		sequence:      fundAccount.GetSequence(),
	}

	if err := signTransaction(
		tx,
		f.keyring.GetKey(fundAccount.GetAddress()),
		sCfg,
	); err != nil {
		return err
	}

	// Broadcast the transaction
	return broadcastTransaction(f.client, tx)
}

// findFundedAccount finds an account
// whose balance is enough to cover the send amount
func (f *Faucet) findFundedAccount() (std.Account, error) {
	// A funded account is an account that can
	// cover the initial transfer fee, as well
	// as the send amount
	estimatedFee := f.estimator.EstimateGasFee()
	requiredFunds := f.sendAmount.Add(std.NewCoins(estimatedFee))

	for _, address := range f.keyring.GetAddresses() {
		// Fetch the account
		account, err := f.client.GetAccount(address)
		if err != nil {
			f.logger.Error(
				"unable to fetch account",
				"address",
				address.String(),
				"error",
				err,
			)

			continue
		}

		// Fetch the balance
		balance := account.GetCoins()

		// Make sure there are enough funds
		if balance.IsAllLT(requiredFunds) {
			f.logger.Error(
				"account cannot serve requests",
				"address",
				address.String(),
				"balance",
				balance.String(),
				"amount",
				requiredFunds,
			)

			continue
		}

		return account, nil
	}

	return nil, errNoFundedAccount
}
