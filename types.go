package faucet

import "net/http"

type Middleware func(next http.Handler) http.Handler

type Handler struct {
	Pattern     string
	HandlerFunc http.HandlerFunc
}
