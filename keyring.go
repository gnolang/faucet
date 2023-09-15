package faucet

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

type keyring struct {
	addresses []crypto.Address
	keyMap    map[crypto.Address]crypto.PrivKey
}

// newKeyring initializes the keyring using the provided mnemonics
func newKeyring(mnemonic string, numAccounts uint64) *keyring {
	addresses := make([]crypto.Address, numAccounts)
	keyMap := make(map[crypto.Address]crypto.PrivKey, numAccounts)

	// Generate the seed
	seed := bip39.NewSeed(mnemonic, "")

	for i := uint32(0); i < uint32(numAccounts); i++ {
		key := generateKeyFromSeed(seed, i)
		address := key.PubKey().Address()

		addresses[i] = address
		keyMap[address] = key
	}

	return &keyring{
		addresses: addresses,
		keyMap:    keyMap,
	}
}

// getAddresses fetches the addresses in the keyring
func (k *keyring) getAddresses() []crypto.Address {
	return k.addresses
}

// getKey fetches the private key associated with the specified address
func (k *keyring) getKey(address crypto.Address) crypto.PrivKey {
	return k.keyMap[address]
}

// generateKeyFromSeed generates a private key from
// the provided seed and index
func generateKeyFromSeed(seed []byte, index uint32) crypto.PrivKey {
	pathParams := hd.NewFundraiserParams(0, crypto.CoinType, index)

	masterPriv, ch := hd.ComputeMastersFromSeed(seed)

	// This derivation can never error out, since the path params
	// are always going to be valid
	derivedPriv, _ := hd.DerivePrivateKeyForPath(masterPriv, ch, pathParams.String())

	return secp256k1.PrivKeySecp256k1(derivedPriv)
}
