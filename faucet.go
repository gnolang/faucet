package faucet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/gnolang/faucet/client"
	"github.com/gnolang/faucet/config"
	"github.com/gnolang/faucet/estimate"
	"github.com/gnolang/faucet/keyring"
	"github.com/gnolang/faucet/keyring/memory"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"
)

// Faucet is a standard Gno faucet
type Faucet struct {
	estimator estimate.Estimator // gas pricing estimations
	logger    *slog.Logger       // log feedback
	client    client.Client      // TM2 client
	keyring   keyring.Keyring    // the faucet keyring

	mux *chi.Mux // HTTP routing

	config         *config.Config     // faucet configuration
	middlewares    []Middleware       // request middlewares
	handlers       []Handler          // request handlers
	prepareTxMsgFn PrepareTxMessageFn // transaction message creator

	maxSendAmount std.Coins // the max send amount per drip
}

var noopLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// NewFaucet creates a new instance of the Gno faucet server
func NewFaucet(
	estimator estimate.Estimator,
	client client.Client,
	opts ...Option,
) (*Faucet, error) {
	f := &Faucet{
		estimator:      estimator,
		client:         client,
		logger:         noopLogger,
		config:         config.DefaultConfig(),
		prepareTxMsgFn: defaultPrepareTxMessage,
		middlewares:    nil, // no middlewares by default

		mux: chi.NewMux(),
	}

	// Set the single default HTTP handler
	f.handlers = []Handler{
		{
			f.defaultHTTPHandler,
			"/",
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
	//nolint:errcheck // MaxSendAmount is validated beforehand
	f.maxSendAmount, _ = std.ParseCoins(f.config.MaxSendAmount)

	// Generate the in-memory keyring
	f.keyring = memory.New(f.config.Mnemonic, f.config.NumAccounts)

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

	// Register the health check handler
	f.mux.Get("/healthcheck", f.healthCheckHandler)

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
		defer f.logger.Info("faucet shut down")

		ln, err := net.Listen("tcp", faucet.Addr)
		if err != nil {
			return err
		}

		f.logger.Info(
			fmt.Sprintf(
				"faucet started at %s",
				ln.Addr().String(),
			),
		)

		if err := faucet.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
