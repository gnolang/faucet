package faucet

import (
	"context"
	"testing"

	"github.com/gnolang/faucet/config"
	"github.com/gnolang/faucet/spec"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFaucet_NewFaucet(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		t.Parallel()

		f, err := NewFaucet(
			&mockEstimator{},
			&mockClient{},
			WithConfig(config.DefaultConfig()),
		)

		assert.NotNil(t, f)
		assert.NoError(t, err)
	})

	t.Run("invalid config", func(t *testing.T) {
		t.Parallel()

		// Create an invalid configuration
		invalidCfg := config.DefaultConfig()
		invalidCfg.NumAccounts = 0

		// Create a faucet instance with the invalid configuration
		f, err := NewFaucet(
			&mockEstimator{},
			&mockClient{},
			WithConfig(invalidCfg),
		)

		// Make sure the error was caught
		assert.Nil(t, f)
		assert.ErrorIs(t, err, config.ErrInvalidNumAccounts)
	})

	t.Run("with CORS config", func(t *testing.T) {
		t.Parallel()

		// Example CORS config
		corsConfig := config.DefaultCORSConfig()
		corsConfig.AllowedOrigins = []string{"gno.land"}

		validCfg := config.DefaultConfig()
		validCfg.CORSConfig = corsConfig

		// Create a valid faucet instance
		// with a valid CORS configuration
		f, err := NewFaucet(
			&mockEstimator{},
			&mockClient{},
			WithConfig(validCfg),
		)

		assert.NotNil(t, f)
		assert.NoError(t, err)
	})

	t.Run("with handlers", func(t *testing.T) {
		t.Parallel()

		handlers := []Handler{
			{
				Pattern: "/hello",
				HandlerFunc: func(_ context.Context, _ *spec.BaseJSONRequest) *spec.BaseJSONResponse {
					return spec.NewJSONResponse(0, nil, nil) // empty handler
				},
			},
		}

		f, err := NewFaucet(
			&mockEstimator{},
			&mockClient{},
			WithConfig(config.DefaultConfig()),
			WithRPCHandlers(handlers),
		)

		require.NotNil(t, f)
		assert.NoError(t, err)

		// Make sure the handler was set
		routes := f.mux.Routes()
		require.Len(t, routes, len(handlers)+2) // base "/" & "/health" handlers as well

		assert.Equal(t, handlers[0].Pattern, routes[2].Pattern)
	})

	t.Run("with logger", func(t *testing.T) {
		t.Parallel()

		f, err := NewFaucet(
			&mockEstimator{},
			&mockClient{},
			WithConfig(config.DefaultConfig()),
			WithLogger(noopLogger),
		)

		assert.NotNil(t, f)
		assert.NoError(t, err)
	})

	t.Run("with prepare transaction message callback", func(t *testing.T) {
		t.Parallel()

		var (
			cfg = PrepareCfg{
				SendAmount:  std.NewCoins(std.NewCoin("ugnot", 10)),
				FromAddress: crypto.Address{1},
				ToAddress:   crypto.Address{2},
			}

			pkgPath = "gno.land/r/demo/example"
			pkgFunc = "FundPlayer"

			prepareTxMsgFn = func(cfg PrepareCfg) std.Msg {
				return vm.MsgCall{
					Caller:  cfg.FromAddress,
					PkgPath: pkgPath,
					Func:    pkgFunc,
					Args:    []string{cfg.ToAddress.String()},
					Send:    cfg.SendAmount,
				}
			}
		)

		f, err := NewFaucet(
			&mockEstimator{},
			&mockClient{},
			WithConfig(config.DefaultConfig()),
			WithPrepareTxMessageFn(prepareTxMsgFn),
		)

		require.NotNil(t, f)
		require.NoError(t, err)

		// Prepare the message
		msg := f.prepareTxMsgFn(cfg)

		// Validate the message
		msgCall, ok := msg.(vm.MsgCall)
		require.True(t, ok)

		assert.Equal(t, cfg.FromAddress, msgCall.Caller)
		assert.Equal(t, pkgPath, msgCall.PkgPath)
		assert.Equal(t, pkgFunc, msgCall.Func)
		assert.Equal(t, []string{cfg.ToAddress.String()}, msgCall.Args)
		assert.Equal(t, cfg.SendAmount, msgCall.Send)
	})
}
