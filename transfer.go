package faucet

import (
	"errors"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// transferFunds transfers funds to the given address
func (f *Faucet) transferFunds(address crypto.Address) error {
	// Find an account that has balance to cover the transfer
	fundAccount, err := f.findFundedAccount()
	if err != nil {
		return err
	}

	sendAmount, _ := std.ParseCoins(f.config.SendAmount)

	prepareCfg := prepareCfg{
		fromAddress: fundAccount.GetAddress(),
		toAddress:   address,
		sendAmount:  sendAmount,
	}

	// Prepare the transaction
	tx := prepareTransaction(f.estimator, prepareCfg)

	// Sign the transaction
	signCfg := signCfg{
		chainID:       f.config.ChainID,
		accountNumber: fundAccount.GetAccountNumber(),
		sequence:      fundAccount.GetSequence(),
	}

	if err := signTransaction(
		tx,
		f.keyring.getKey(fundAccount.GetAddress()),
		signCfg,
	); err != nil {
		return err
	}

	// Broadcast the transaction
	return broadcastTransaction(f.client, tx)
}

// findFundedAccount finds an account
// whose balance is enough to cover the send amount
func (f *Faucet) findFundedAccount() (std.Account, error) {
	sendAmount, _ := std.ParseCoins(f.config.SendAmount)

	for _, address := range f.keyring.getAddresses() {
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
		if balance.IsAllLT(sendAmount) {
			f.logger.Error(
				"account cannot serve requests",
				"address",
				address.String(),
				"balance",
				balance.String(),
				"amount",
				sendAmount,
			)

			continue
		}

		return account, nil
	}

	return nil, errors.New("no funded account found")
}
