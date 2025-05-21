package faucet

import (
	"log/slog"
	"net/http"

	"github.com/gnolang/faucet/config"
)

type Option func(f *Faucet)

// WithLogger specifies the logger for the faucet
func WithLogger(l *slog.Logger) Option {
	return func(f *Faucet) {
		f.logger = l
	}
}

// WithConfig specifies the config for the faucet
func WithConfig(c *config.Config) Option {
	return func(f *Faucet) {
		f.config = c
	}
}

// WithMiddlewares specifies the rpc request middlewares for the faucet.
// Middlewares are applied for each endpoint
func WithMiddlewares(middlewares []Middleware) Option {
	return func(f *Faucet) {
		f.rpcMiddlewares = middlewares
	}
}

// WithHTTPMiddlewares specifies the http request middlewares for the faucet.
// Middlewares are applied for each endpoint
func WithHTTPMiddlewares(middlewares []func(http.Handler) http.Handler) Option {
	return func(f *Faucet) {
		f.httpMiddlewares = middlewares
	}
}

// WithRPCHandlers specifies the HTTP handlers for the faucet
func WithRPCHandlers(handlers []Handler) Option {
	return func(f *Faucet) {
		f.rpcHandlers = append(f.rpcHandlers, handlers...)
	}
}

// WithPrepareTxMessageFn specifies the faucet
// transaction message constructor
func WithPrepareTxMessageFn(prepareTxMsgFn PrepareTxMessageFn) Option {
	return func(f *Faucet) {
		f.prepareTxMsgFn = prepareTxMsgFn
	}
}
