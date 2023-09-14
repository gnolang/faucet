package faucet

import "net/http"

type Estimator interface {
	// TODO define
}

type Logger interface {
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type Middleware func(next http.Handler) http.Handler
