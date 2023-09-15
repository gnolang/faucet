package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_ValidateConfig(t *testing.T) {
	t.Parallel()

	t.Run("invalid listen address", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.ListenAddress = "gno.land" // doesn't follow the format

		assert.ErrorIs(t, ValidateConfig(cfg), ErrInvalidListenAddress)
	})

	t.Run("invalid chain ID", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.ChainID = "" // empty

		assert.ErrorIs(t, ValidateConfig(cfg), ErrInvalidChainID)
	})

	t.Run("invalid send amount", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.SendAmount = "1000goo" // invalid denom

		assert.ErrorIs(t, ValidateConfig(cfg), ErrInvalidSendAmount)
	})

	t.Run("invalid gas fee", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.GasFee = "1000goo" // invalid denom

		assert.ErrorIs(t, ValidateConfig(cfg), ErrInvalidGasFee)
	})

	t.Run("invalid gas wanted", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.GasWanted = "totally a number" // invalid number

		assert.ErrorIs(t, ValidateConfig(cfg), ErrInvalidGasWanted)
	})

	t.Run("invalid mnemonic", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.Mnemonic = "maybe valid mnemonic" // invalid mnemonic

		assert.ErrorIs(t, ValidateConfig(cfg), ErrInvalidMnemonic)
	})

	t.Run("invalid num accounts", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultConfig()
		cfg.NumAccounts = 0 // invalid number of accounts

		assert.ErrorIs(t, ValidateConfig(cfg), ErrInvalidNumAccounts)
	})

	t.Run("valid configuration", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, ValidateConfig(DefaultConfig()))
	})
}
