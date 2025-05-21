package faucet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/gnolang/faucet/spec"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const faucetSuccess = "successfully executed faucet transfer"

const DefaultDripMethod = "drip" // the default JSON-RPC method for a faucet drip

var (
	errInvalidBeneficiary = errors.New("invalid beneficiary address")
	errInvalidSendAmount  = errors.New("invalid send amount")
	errInvalidMethod      = errors.New("unknown RPC method call")
)

// wrapJSONRPC wraps the given handler and rpcMiddlewares into a JSON-RPC 2.0 pipeline
func wrapJSONRPC(handlerFn HandlerFunc, mws ...Middleware) http.HandlerFunc {
	callChain := chainMiddlewares(mws...)(handlerFn)

	return func(w http.ResponseWriter, r *http.Request) {
		// Grab the request(s)
		requests, err := parseRequests(r.Body)
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("unable to read request: %s", err.Error()),
				http.StatusBadRequest,
			)

			return
		}

		var (
			ctx = r.Context()

			responses = make(spec.BaseJSONResponses, 0)
		)

		for _, req := range requests {
			// Make sure it's a valid base request
			if !spec.IsValidBaseRequest(req) {
				responses = append(responses, spec.NewJSONResponse(
					req.ID,
					nil,
					spec.NewJSONError("invalid JSON-RPC 2.0 request", spec.InvalidRequestErrorCode),
				))

				continue
			}

			// Parse the request.
			// This executes all the rpcMiddlewares, and
			// finally the base handler for the endpoint
			resp := callChain(ctx, req)

			responses = append(responses, resp)
		}

		w.Header().Set("Content-Type", JSONMimeType)

		// Create the encoder
		enc := json.NewEncoder(w)

		if len(responses) == 1 {
			// Write the JSON response as a single response
			_ = enc.Encode(responses[0]) //nolint:errcheck // Fine to leave unchecked

			return
		}

		// Write the JSON response as a batch
		_ = enc.Encode(responses) //nolint:errcheck // Fine to leave unchecked
	}
}

// chainMiddlewares combines the given rpcMiddlewares
func chainMiddlewares(mw ...Middleware) Middleware {
	return func(final HandlerFunc) HandlerFunc {
		h := final

		for i := len(mw) - 1; i >= 0; i-- {
			h = mw[i](h)
		}

		return h
	}
}

// parseRequests parses the JSON-RPC requests from the request body
func parseRequests(body io.Reader) (spec.BaseJSONRequests, error) {
	// Load the requests
	requestBody, readErr := io.ReadAll(body)
	if readErr != nil {
		return nil, fmt.Errorf("unable to read request: %w", readErr)
	}

	// Extract the requests
	requests, err := spec.ExtractBaseRequests(requestBody)
	if err != nil {
		return nil, fmt.Errorf("invalid request body: %w", err)
	}

	return requests, nil
}

// drip is a single Faucet transfer request
type drip struct {
	amount std.Coins
	to     crypto.Address
}

var amountRegex = regexp.MustCompile(`^\d+ugnot$`)

// defaultHTTPHandler is the default faucet transfer handler
func (f *Faucet) defaultHTTPHandler(_ context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
	// Make sure the method call on "/" is "drip"
	if req.Method != DefaultDripMethod {
		return spec.NewJSONResponse(
			req.ID,
			nil,
			spec.NewJSONError(errInvalidMethod.Error(), spec.MethodNotFoundErrorCode),
		)
	}

	// Parse params into a drip request
	dripRequest, err := extractDripRequest(req.Params)
	if err != nil {
		return spec.NewJSONResponse(
			req.ID,
			nil,
			spec.NewJSONError(err.Error(), spec.InvalidParamsErrorCode),
		)
	}

	// Check if the amount is set
	if dripRequest.amount.IsZero() {
		// drip amount is not set, use
		// the max faucet drip amount
		dripRequest.amount = f.maxSendAmount
	}

	// Check if the amount exceeds the max
	// drip amount for the faucet
	if dripRequest.amount.IsAllGT(f.maxSendAmount) {
		return spec.NewJSONResponse(
			req.ID,
			nil,
			spec.NewJSONError(errInvalidSendAmount.Error(), spec.InvalidRequestErrorCode),
		)
	}

	// Attempt fund transfer
	if err := f.transferFunds(dripRequest.to, dripRequest.amount); err != nil {
		f.logger.Debug("unable to handle drip", "req", req, "err", err)

		return spec.NewJSONResponse(req.ID, nil, spec.GenerateResponseError(err))
	}

	return spec.NewJSONResponse(req.ID, faucetSuccess, nil)
}

// extractDripRequest extracts the base drip params from the request
func extractDripRequest(params []any) (*drip, error) {
	// Extract the drip params
	if len(params) < 1 {
		return nil, errInvalidBeneficiary
	}

	addrStr, ok := params[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: beneficiary must be a string", errInvalidBeneficiary)
	}

	// Validate the beneficiary address is valid
	beneficiary, err := crypto.AddressFromBech32(addrStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidBeneficiary, err)
	}

	// Validate the send amount is valid
	if len(params) == 1 {
		// No amount specified
		return &drip{
			to:     beneficiary,
			amount: std.Coins{},
		}, nil
	}

	amountStr, ok := params[1].(string)
	if !ok || !amountRegex.MatchString(amountStr) {
		return nil, errInvalidSendAmount
	}

	amount, err := std.ParseCoins(amountStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidSendAmount, err)
	}

	return &drip{
		to:     beneficiary,
		amount: amount,
	}, nil
}

// healthcheckHandler is the default health check handler for the faucet
func (f *Faucet) healthcheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
