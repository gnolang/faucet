package faucet

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/gnolang/faucet/spec"
	"github.com/gnolang/faucet/writer"
	httpWriter "github.com/gnolang/faucet/writer/http"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	unableToHandleRequest = "unable to handle faucet request"
	faucetSuccess         = "successfully executed faucet transfer"
)

const DefaultDripMethod = "drip" // the default JSON-RPC method for a faucet drip

var (
	errInvalidBeneficiary = errors.New("invalid beneficiary address")
	errInvalidSendAmount  = errors.New("invalid send amount")
	errInvalidMethod      = errors.New("unknown RPC method call")
)

// drip is a single Faucet transfer request
type drip struct {
	amount std.Coins
	to     crypto.Address
}

var amountRegex = regexp.MustCompile(`^\d+ugnot$`)

// defaultHTTPHandler is the default faucet transfer handler
func (f *Faucet) defaultHTTPHandler(w http.ResponseWriter, r *http.Request) {
	// Load the requests
	requestBody, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		http.Error(
			w,
			"unable to read request",
			http.StatusBadRequest,
		)

		return
	}

	// Extract the requests
	requests, err := spec.ExtractBaseRequests(requestBody)
	if err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)

		return
	}

	// Handle the requests
	w.Header().Set("Content-Type", jsonMimeType)
	f.handleRequest(
		httpWriter.New(f.logger, w),
		requests,
	)
}

// handleRequest is the common default faucet handler
func (f *Faucet) handleRequest(writer writer.ResponseWriter, requests spec.BaseJSONRequests) {
	// Parse all JSON-RPC requests
	responses := make(spec.BaseJSONResponses, len(requests))

	for i, req := range requests {
		f.logger.Debug("incoming request", "request", req)

		responses[i] = f.handleSingleRequest(req)
	}

	if len(responses) == 1 {
		// Write the JSON response as a single response
		writer.WriteResponse(responses[0])

		return
	}

	// Write the JSON response as a batch
	writer.WriteResponse(responses)
}

// handleSingleRequest validates and executes one drip request
func (f *Faucet) handleSingleRequest(req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
	// Make sure it's a valid base request
	if !spec.IsValidBaseRequest(req) {
		return spec.NewJSONResponse(
			req.ID,
			nil,
			spec.NewJSONError("invalid JSON-RPC 2.0 request", spec.InvalidRequestErrorCode),
		)
	}

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
