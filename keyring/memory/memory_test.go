package memory

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/stretchr/testify/assert"
)

// generateTestMnemonic generates a new test BIP39 mnemonic using the provided entropy size
func generateTestMnemonic(t *testing.T) string {
	// Generate the entropy seed
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		t.Fatal(err)
	}

	// Generate the actual mnemonic
	mnemonic, err := bip39.NewMnemonic(entropySeed[:])
	if err != nil {
		t.Fatal(err)
	}

	return mnemonic
}

func TestKeyring_NewKeyring(t *testing.T) {
	t.Parallel()

	var (
		mnemonic    = generateTestMnemonic(t)
		numAccounts = uint64(10)
	)

	// Create the keyring
	kr := New(mnemonic, numAccounts)

	// Make sure the keyring is initialized correctly
	assert.Len(t, kr.addresses, int(numAccounts))
	assert.Len(t, kr.keyMap, int(numAccounts))
}

func TestKeyring_GetAddresses(t *testing.T) {
	t.Parallel()

	var (
		mnemonic    = generateTestMnemonic(t)
		numAccounts = uint64(10)
	)

	// Create the keyring
	kr := New(mnemonic, numAccounts)

	// Fetch the addresses
	addresses := kr.GetAddresses()

	// Make sure the addresses are valid
	assert.Len(t, addresses, int(numAccounts))

	for _, address := range addresses {
		assert.False(t, address.IsZero())
	}
}

func TestKeyring_GetKey(t *testing.T) {
	t.Parallel()

	var (
		mnemonic    = generateTestMnemonic(t)
		numAccounts = uint64(10)
	)

	// Create the keyring
	kr := New(mnemonic, numAccounts)

	// Fetch the addresses
	addresses := kr.GetAddresses()

	// Make sure the addresses are valid
	assert.Len(t, addresses, int(numAccounts))

	// Fetch the key associated with an address
	address := addresses[0]
	key := kr.GetKey(address)

	// Make sure the key matches the address
	assert.Equal(t, address, key.PubKey().Address())
}
