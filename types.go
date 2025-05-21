package faucet

import (
	"context"

	"github.com/gnolang/faucet/spec"
)

const JSONMimeType = "application/json"

// Middleware is the faucet middleware func
type Middleware func(next HandlerFunc) HandlerFunc

// Handler defines a faucet pattern handler
type Handler struct {
	HandlerFunc HandlerFunc
	Pattern     string
}

// HandlerFunc is the custom faucet request handler
type HandlerFunc func(ctx context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse
