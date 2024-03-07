package faucet

import (
	"net/http"
	"testing"

	"github.com/gnolang/faucet/config"
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

	t.Run("with middlewares", func(t *testing.T) {
		t.Parallel()

		middlewares := []Middleware{
			func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Example empty middleware
					next.ServeHTTP(w, r)
				})
			},
		}

		cfg := config.DefaultConfig()
		cfg.CORSConfig = nil // disable CORS middleware

		f, err := NewFaucet(
			&mockEstimator{},
			&mockClient{},
			WithConfig(cfg),
			WithMiddlewares(middlewares),
		)

		require.NotNil(t, f)
		assert.NoError(t, err)

		// Make sure the middleware was set
		assert.Len(t, f.mux.Middlewares(), len(middlewares))
	})

	t.Run("with handlers", func(t *testing.T) {
		t.Parallel()

		handlers := []Handler{
			{
				Pattern: "/hello",
				HandlerFunc: func(_ http.ResponseWriter, _ *http.Request) {
					// Empty handler
				},
			},
		}

		f, err := NewFaucet(
			&mockEstimator{},
			&mockClient{},
			WithConfig(config.DefaultConfig()),
			WithHandlers(handlers),
		)

		require.NotNil(t, f)
		assert.NoError(t, err)

		// Make sure the handler was set
		routes := f.mux.Routes()
		require.Len(t, routes, len(handlers)+1) // base "/" handler as well

		assert.Equal(t, handlers[0].Pattern, routes[1].Pattern)
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
