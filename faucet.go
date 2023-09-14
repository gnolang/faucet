package faucet

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gnolang/faucet/config"
	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/errgroup"
)

// Faucet is a standard Gno
// native currency faucet
type Faucet struct {
	estimator Estimator // gas pricing estimations
	logger    Logger    // log feedback

	mux *chi.Mux // HTTP routing

	config      *config.Config // faucet configuration
	middlewares []Middleware   // request middlewares
}

// NewFaucet creates a new instance of the Gno faucet server
func NewFaucet(opts ...Option) *Faucet {
	f := &Faucet{
		estimator:   nil, // TODO static estimator
		logger:      nil, // TODO nil logger
		config:      nil, // TODO default config
		middlewares: nil, // TODO no middlewares

		mux: chi.NewMux(),
	}

	// Apply the options
	for _, opt := range opts {
		opt(f)
	}

	return f
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
