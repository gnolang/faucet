package keyring

import "github.com/gnolang/gno/tm2/pkg/crypto"

// Keyring defines the faucet keyring functionality
type Keyring interface {
	// GetAddresses fetches the addresses in the keyring
	GetAddresses() []crypto.Address

	// GetKey fetches the private key associated with the specified address
	GetKey(address crypto.Address) crypto.PrivKey
}
