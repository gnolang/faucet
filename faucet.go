package faucet

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gnolang/faucet/client"
	"github.com/gnolang/faucet/config"
	"github.com/gnolang/faucet/estimate"
	"github.com/gnolang/faucet/log"
	"github.com/gnolang/faucet/log/noop"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"
)

// Faucet is a standard Gno faucet
type Faucet struct {
	estimator estimate.Estimator // gas pricing estimations
	logger    log.Logger         // log feedback
	client    client.Client      // TM2 client

	mux *chi.Mux // HTTP routing

	config      *config.Config // faucet configuration
	middlewares []Middleware   // request middlewares
	handlers    []Handler      // request handlers

	keyring    *keyring  // the faucet keyring
	sendAmount std.Coins // for fast lookup
}

// NewFaucet creates a new instance of the Gno faucet server
func NewFaucet(
	estimator estimate.Estimator,
	client client.Client,
	opts ...Option,
) (*Faucet, error) {
	f := &Faucet{
		estimator:   estimator,
		client:      client,
		logger:      noop.New(),
		config:      config.DefaultConfig(),
		middlewares: nil, // no middlewares by default

		mux: chi.NewMux(),
	}

	// Set the single default HTTP handler
	f.handlers = []Handler{
		{
			"/",
			f.defaultHTTPHandler,
		},
	}

	// Apply the options
	for _, opt := range opts {
		opt(f)
	}

	// Validate the configuration
	if err := config.ValidateConfig(f.config); err != nil {
		return nil, fmt.Errorf("invalid configuration, %w", err)
	}

	// Set the send amount
	f.sendAmount, _ = std.ParseCoins(f.config.SendAmount)

	// Generate the keyring
	f.keyring = newKeyring(f.config.Mnemonic, f.config.NumAccounts)

	// Set up the CORS middleware
	if f.config.CORSConfig != nil {
		corsMiddleware := cors.New(cors.Options{
			AllowedOrigins: f.config.CORSConfig.AllowedOrigins,
			AllowedMethods: f.config.CORSConfig.AllowedMethods,
			AllowedHeaders: f.config.CORSConfig.AllowedHeaders,
		})

		f.mux.Use(corsMiddleware.Handler)
	}

	// Set up additional middlewares
	for _, middleware := range f.middlewares {
		f.mux.Use(middleware)
	}

	// Set up the request handlers
	for _, handler := range f.handlers {
		f.mux.Post(handler.Pattern, handler.HandlerFunc)
	}

	return f, nil
}

// Serve serves the Gno faucet [BLOCKING]
func (f *Faucet) Serve(ctx context.Context) error {
	faucet := &http.Server{
		Addr:              f.config.ListenAddress,
		Handler:           f.mux,
		ReadHeaderTimeout: 60 * time.Second,
	}

	group, gCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		f.logger.Info(
			fmt.Sprintf(
				"faucet started at %s",
				f.config.ListenAddress,
			),
		)
		defer f.logger.Info("faucet shut down")

		if err := faucet.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	group.Go(func() error {
		<-gCtx.Done()

		f.logger.Info("faucet to be shutdown")

		wsCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		return faucet.Shutdown(wsCtx)
	})

	return group.Wait()
}
