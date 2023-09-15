package faucet

import "net/http"

type Middleware func(next http.Handler) http.Handler

type Handler struct {
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Requests []Request

type Request struct {
	To string `json:"to"`
}

type Responses []Response

type Response struct {
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

const (
	jsonMimeType = "application/json"
)
