package faucet

import (
	"log/slog"

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

// WithMiddlewares specifies the request middlewares for the faucet
func WithMiddlewares(middlewares []Middleware) Option {
	return func(f *Faucet) {
		f.middlewares = middlewares
	}
}

// WithHandlers specifies the HTTP handlers for the faucet
func WithHandlers(handlers []Handler) Option {
	return func(f *Faucet) {
		f.handlers = append(f.handlers, handlers...)
	}
}

// WithPrepareTxMessageFn specifies the faucet
// transaction message constructor
func WithPrepareTxMessageFn(prepareTxMsgFn PrepareTxMessageFn) Option {
	return func(f *Faucet) {
		f.prepareTxMsgFn = prepareTxMsgFn
	}
}

// WithHealthcheck specifies whether to enable the health check
func WithHealthcheck(enable bool) Option {
	return func(f *Faucet) {
		f.healthcheck = enable
	}
}
