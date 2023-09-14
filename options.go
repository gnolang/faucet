package faucet

import (
	"github.com/gnolang/faucet/config"
)

type Option func(f *Faucet)

func WithLogger(l Logger) Option {
	return func(f *Faucet) {
		f.logger = l
	}
}

func WithEstimator(e Estimator) Option {
	return func(f *Faucet) {
		f.estimator = e
	}
}

func WithConfig(c *config.Config) Option {
	return func(f *Faucet) {
		f.config = c
	}
}

func WithMiddlewares(middlewares []Middleware) Option {
	return func(f *Faucet) {
		f.middlewares = middlewares
	}
}
