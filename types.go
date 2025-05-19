package faucet

import (
	"net/http"
)

const (
	jsonMimeType = "application/json"
)

// Middleware is the faucet middleware func
type Middleware func(next http.Handler) http.Handler

// Handler defines a faucet handler
type Handler struct {
	HandlerFunc http.HandlerFunc
	Pattern     string
}
