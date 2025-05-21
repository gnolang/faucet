package spec

import "encoding/json"

const JSONRPCVersion = "2.0"

// BaseJSON defines the base JSON fields
// all JSON-RPC requests and responses need to have
type BaseJSON struct {
	JSONRPC string `json:"jsonrpc"`
	ID      uint   `json:"id"`
}

// BaseJSONRequest defines the base JSON request format
type BaseJSONRequest struct {
	BaseJSON

	Method string          `json:"method"`
	Params []any           `json:"params"`
	Meta   json.RawMessage `json:"meta"`
}

// BaseJSONRequests represents a batch of JSON-RPC requests
type BaseJSONRequests []*BaseJSONRequest

// BaseJSONResponses represents a batch of JSON-RPC responses
type BaseJSONResponses []*BaseJSONResponse

// BaseJSONResponse defines the base JSON response format
type BaseJSONResponse struct {
	Result any            `json:"result,omitempty"`
	Error  *BaseJSONError `json:"error,omitempty"`
	BaseJSON
}

// BaseJSONError defines the base JSON response error format
type BaseJSONError struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// NewJSONResponse creates a new JSON-RPC response
func NewJSONResponse(
	id uint,
	result any,
	err *BaseJSONError,
) *BaseJSONResponse {
	return &BaseJSONResponse{
		BaseJSON: BaseJSON{
			ID:      id,
			JSONRPC: JSONRPCVersion,
		},
		Result: result,
		Error:  err,
	}
}

// NewJSONError creates a new JSON-RPC error
func NewJSONError(message string, code int) *BaseJSONError {
	return &BaseJSONError{
		Code:    code,
		Message: message,
	}
}

// NewJSONRequest creates a new JSON-RPC request
func NewJSONRequest(
	id uint,
	method string,
	params []any,
) *BaseJSONRequest {
	return &BaseJSONRequest{
		BaseJSON: BaseJSON{
			ID:      id,
			JSONRPC: JSONRPCVersion,
		},
		Method: method,
		Params: params,
	}
}

// GenerateResponseError generates the JSON-RPC server error response
func GenerateResponseError(err error) *BaseJSONError {
	return NewJSONError(err.Error(), ServerErrorCode)
}

// ExtractBaseRequests extracts the base JSON-RPC request from the request body
func ExtractBaseRequests(requestBody []byte) (BaseJSONRequests, error) {
	// Extract the request
	var requests BaseJSONRequests

	// Check if the request is a batch request
	if err := json.Unmarshal(requestBody, &requests); err != nil {
		// Try to get a single JSON-RPC request, since this is not a batch
		var baseRequest *BaseJSONRequest
		if err := json.Unmarshal(requestBody, &baseRequest); err != nil {
			return nil, err
		}

		requests = BaseJSONRequests{
			baseRequest,
		}
	}

	return requests, nil
}

// IsValidBaseRequest validates that the base JSON request is valid
func IsValidBaseRequest(baseRequest *BaseJSONRequest) bool {
	if baseRequest.Method == "" {
		return false
	}

	return baseRequest.JSONRPC == JSONRPCVersion
}
