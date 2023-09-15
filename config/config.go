package config

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
)

const (
	DefaultListenAddress = "0.0.0.0:8545"
	DefaultRemote        = "http://127.0.0.1:26657"
	DefaultChainID       = "dev"
	DefaultSendAmount    = "1000000ugnot"
	DefaultGasFee        = "1000000ugnot"
	DefaultGasWanted     = "100000"
	DefaultMnemonic      = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"
	DefaultNumAccounts   = uint64(1)
)

var (
	listenAddressRegex = regexp.MustCompile(`^\d{1,3}(\.\d{1,3}){3}:\d+$`)
	remoteRegex        = regexp.MustCompile(`^https?:\/\/(?:w{1,3}\.)?[^\s.]+(?:\.[a-z]+)*(?::\d+)(?![^<]*(?:<\/\w+>|\/?>))$`)
	amountRegex        = regexp.MustCompile(`^\d+ugnot$`)
	numberRegex        = regexp.MustCompile(`^\d+$`)
)

// Config defines the base-level Faucet configuration
type Config struct {
	// The address at which the faucet will be served.
	// Format should be: <IP>:<PORT>
	ListenAddress string `toml:"listen_address"`

	// The JSON-RPC URL of the Gno chain
	Remote string `toml:"remote"`

	// The chain ID associated with the remote Gno chain
	ChainID string `toml:"chain_id"`

	// The mnemonic for the faucet
	Mnemonic string `toml:"mnemonic"`

	// The number of faucet accounts,
	// based on the mnemonic (account 0, index x)
	NumAccounts uint64 `toml:"num_accounts"`

	// The static send amount (native currency).
	// Format should be: <AMOUNT>ugnot
	SendAmount string `toml:"send_amount"`

	// The static gas fee for the transaction.
	// Format should be: <AMOUNT>ugnot
	GasFee string `toml:"gas_fee"`

	// The static gas wanted for the transaction.
	// Format should be: <AMOUNT>
	GasWanted string `toml:"gas_wanted"`

	// The associated CORS config, if any
	CORSConfig *CORS `toml:"cors_config"`
}

// DefaultConfig returns the default faucet configuration
func DefaultConfig() *Config {
	return &Config{
		ListenAddress: DefaultListenAddress,
		Remote:        DefaultRemote,
		ChainID:       DefaultChainID,
		SendAmount:    DefaultSendAmount,
		GasFee:        DefaultGasFee,
		GasWanted:     DefaultGasWanted,
		Mnemonic:      DefaultMnemonic,
		NumAccounts:   DefaultNumAccounts,
		CORSConfig:    DefaultCORSConfig(),
	}
}

// ValidateConfig validates the faucet configuration
func ValidateConfig(config *Config) error {
	// validate the listen address
	if !listenAddressRegex.Match([]byte(config.ListenAddress)) {
		return errors.New("invalid listen address")
	}

	// validate the remote address
	if !remoteRegex.Match([]byte(config.Remote)) {
		return errors.New("invalid remote address")
	}

	// validate the chain ID
	if config.ChainID == "" {
		return errors.New("invalid chain ID")
	}

	// validate the send amount
	if !amountRegex.Match([]byte(config.SendAmount)) {
		return errors.New("invalid send amount")
	}

	// validate the gas fee
	if !amountRegex.Match([]byte(config.GasFee)) {
		return errors.New("invalid gas fee")
	}

	// validate the gas wanted
	if !numberRegex.Match([]byte(config.GasWanted)) {
		return errors.New("invalid gas wanted")
	}

	// validate the mnemonic is bip39-compliant
	if !bip39.IsMnemonicValid(config.Mnemonic) {
		return fmt.Errorf("invalid mnemonic, %s", config.Mnemonic)
	}

	// validate at least one faucet account is set
	if config.NumAccounts < 1 {
		return errors.New("invalid number of faucet accounts")
	}

	return nil
}
