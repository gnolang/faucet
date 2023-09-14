package config

// Config defines the base-level Faucet configuration
type Config struct {
	// The address at which the faucet will be served.
	// Format should be: <IP>:<PORT>
	ListenAddress string `toml:"listen_address"`

	// The JSON-RPC URL of the Gno chain
	Remote string `toml:"remote"`

	// The chain ID associated with the remote Gno chain
	ChainID string `toml:"chain_id"`

	// The static gas fee for the transaction.
	// Format should be: <AMOUNT>ugnot
	GasFee string `toml:"gas_fee"`

	// The static gas wanted for the transaction.
	// Format should be: <AMOUNT>ugnot
	GasWanted string `toml:"gas_wanted"`

	// The associated CORS config, if any
	CORSConfig *CORS `toml:"cors_config"`
}
