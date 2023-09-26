package faucet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gnolang/faucet/writer"
	httpWriter "github.com/gnolang/faucet/writer/http"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

const (
	unableToHandleRequest = "unable to handle faucet request"
	faucetSuccess         = "successfully executed faucet transfer"
)

var errInvalidBeneficiary = errors.New("invalid beneficiary address")

// defaultHTTPHandler is the default faucet transfer handler
func (f *Faucet) defaultHTTPHandler(w http.ResponseWriter, r *http.Request) {
	// Load the requests
	requestBody, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		http.Error(w, "unable to read request", http.StatusBadRequest)

		return
	}

	// Extract the requests
	requests, err := extractRequests(requestBody)
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
func (f *Faucet) handleRequest(writer writer.ResponseWriter, requests Requests) {
	// Parse all JSON-RPC requests
	responses := make(Responses, len(requests))

	for i, baseRequest := range requests {
		// Log the request
		f.logger.Debug(
			"incoming request",
			"request",
			baseRequest,
		)

		// Extract the beneficiary
		beneficiary, err := extractBeneficiary(baseRequest)
		if err != nil {
			// Save the error response
			responses[i] = Response{
				Result: unableToHandleRequest,
				Error:  err.Error(),
			}

			continue
		}

		// Run the method handler
		if err := f.transferFunds(beneficiary); err != nil {
			f.logger.Debug(
				unableToHandleRequest,
				"request",
				baseRequest,
				"error",
				err,
			)

			responses[i] = Response{
				Result: unableToHandleRequest,
				Error:  err.Error(),
			}

			continue
		}

		responses[i] = Response{
			Result: faucetSuccess,
		}
	}

	if len(responses) == 1 {
		// Write the JSON response as a single response
		writer.WriteResponse(responses[0])

		return
	}

	// Write the JSON response as a batch
	writer.WriteResponse(responses)
}

// extractRequests extracts the base JSON requests from the request body
func extractRequests(requestBody []byte) (Requests, error) {
	// Extract the request
	var requests Requests

	// Check if the request is a batch request
	if err := json.Unmarshal(requestBody, &requests); err != nil {
		// Try to get a single JSON request, since this is not a batch
		var baseRequest Request
		if err := json.Unmarshal(requestBody, &baseRequest); err != nil {
			return nil, err
		}

		requests = Requests{
			baseRequest,
		}
	}

	return requests, nil
}

// extractBeneficiary extracts the beneficiary from the base faucet request
func extractBeneficiary(request Request) (crypto.Address, error) {
	// Validate the beneficiary address is set
	if request.To == "" {
		return crypto.Address{}, errInvalidBeneficiary
	}

	// Validate the beneficiary address is valid
	beneficiary, err := crypto.AddressFromBech32(request.To)
	if err != nil {
		return crypto.Address{}, fmt.Errorf("%w, %w", errInvalidBeneficiary, err)
	}

	return beneficiary, nil
}
