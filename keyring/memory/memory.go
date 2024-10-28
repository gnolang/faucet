package memory

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// Keyring is an in-memory keyring
type Keyring struct {
	keyMap    map[crypto.Address]crypto.PrivKey
	addresses []crypto.Address
}

// New initializes the keyring using the provided mnemonics
func New(mnemonic string, numAccounts uint64) *Keyring {
	addresses := make([]crypto.Address, numAccounts)
	keyMap := make(map[crypto.Address]crypto.PrivKey, numAccounts)

	// Generate the seed
	seed := bip39.NewSeed(mnemonic, "")

	for i := uint64(0); i < numAccounts; i++ {
		//nolint:gosec // i ranges up to numAccounts which won't overflow uint32
		key := generateKeyFromSeed(seed, uint32(i))
		address := key.PubKey().Address()

		addresses[i] = address
		keyMap[address] = key
	}

	return &Keyring{
		addresses: addresses,
		keyMap:    keyMap,
	}
}

// GetAddresses fetches the addresses in the keyring
func (k *Keyring) GetAddresses() []crypto.Address {
	return k.addresses
}

// GetKey fetches the private key associated with the specified address
func (k *Keyring) GetKey(address crypto.Address) crypto.PrivKey {
	return k.keyMap[address]
}

// generateKeyFromSeed generates a private key from
// the provided seed and index
func generateKeyFromSeed(seed []byte, index uint32) crypto.PrivKey {
	pathParams := hd.NewFundraiserParams(0, crypto.CoinType, index)

	masterPriv, ch := hd.ComputeMastersFromSeed(seed)

	//nolint:errcheck // This derivation can never error out, since the path params
	// are always going to be valid
	derivedPriv, _ := hd.DerivePrivateKeyForPath(masterPriv, ch, pathParams.String())

	return secp256k1.PrivKeySecp256k1(derivedPriv)
}
