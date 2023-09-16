package faucet

import "net/http"

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

type Requests []Request

// Request is a single Faucet transfer request
type Request struct {
	To string `json:"to"`
}

type Responses []Response

// Response is a single Faucet transfer response
type Response struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}
