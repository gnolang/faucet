package faucet

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

type keyring []crypto.PrivKey

// newKeyring initializes the keyring using the provided mnemonics
func newKeyring(mnemonic string, numAccounts uint64) (keyring, error) {
	keys := make(keyring, numAccounts)

	// Generate the seed
	seed := bip39.NewSeed(mnemonic, "")

	for i := uint32(0); i < uint32(numAccounts); i++ {
		key, err := generateKeyFromSeed(seed, i)
		if err != nil {
			return nil, err
		}

		keys[i] = key
	}

	return keys, nil
}

// getAddresses fetches the addresses in the keyring
func (k keyring) getAddresses() []crypto.Address {
	addresses := make([]crypto.Address, len(k))

	for index, key := range k {
		addresses[index] = key.PubKey().Address()
	}

	return addresses
}

func (k keyring) getKey(address crypto.Address) crypto.PrivKey {
	for _, key := range k {
		if key.PubKey().Address() == address {
			return key
		}
	}

	return nil
}

// generateKeyFromSeed generates a private key from
// the provided seed and index
func generateKeyFromSeed(seed []byte, index uint32) (crypto.PrivKey, error) {
	pathParams := hd.NewFundraiserParams(0, crypto.CoinType, index)

	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, pathParams.String())
	if err != nil {
		return nil, err
	}

	return secp256k1.PrivKeySecp256k1(derivedPriv), nil
}
