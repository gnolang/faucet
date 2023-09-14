package config

import (
	"errors"
	"regexp"
)

const (
	DefaultListenAddress = "0.0.0.0:8545"
	DefaultRemote        = "http://127.0.0.1:26657"
	DefaultChainID       = "dev"
	DefaultSendAmount    = "1000000ugnot"
	DefaultGasFee        = "1000000ugnot"
	DefaultGasWanted     = "100000"
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

	return nil
}
