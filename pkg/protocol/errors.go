package protocol

import (
	"fmt"
)

// Standard JSON-RPC error codes
const (
	// Parse error
	ErrParse = -32700
	// Invalid request
	ErrInvalidRequest = -32600
	// Method not found
	ErrMethodNotFound = -32601
	// Invalid params
	ErrInvalidParams = -32602
	// Internal error
	ErrInternal = -32603
	// Server error (reserved for implementation-defined server errors)
	ErrServer = -32000
)

// JsonRpcError represents a JSON-RPC error
type JsonRpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	ID      interface{} // The ID of the request that caused this error
}

// Error implements the error interface
func (e *JsonRpcError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("JSON-RPC error %d: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// ToResponse converts the error to a JSON-RPC response message
func (e *JsonRpcError) ToResponse() JSONRPCMessage {
	errorObj := map[string]interface{}{
		"code":    e.Code,
		"message": e.Message,
	}

	if e.Data != nil {
		errorObj["data"] = e.Data
	}

	return JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      e.ID,
		Error:   errorObj,
	}
}

// Is implements the errors.Is interface for JsonRpcError
// This allows errors.Is(err, &JsonRpcError{}) to work properly
func (e *JsonRpcError) Is(target error) bool {
	targetErr, ok := target.(*JsonRpcError)
	if !ok {
		return false
	}

	// If the target error has a specific code, check if it matches
	if targetErr.Code != 0 && e.Code != targetErr.Code {
		return false
	}

	// If the target error has a specific message, check if it matches
	if targetErr.Message != "" && e.Message != targetErr.Message {
		return false
	}

	return true
}

// NewError creates a new JSON-RPC error
func NewError(code int, message string, data interface{}, id interface{}) *JsonRpcError {
	return &JsonRpcError{
		Code:    code,
		Message: message,
		Data:    data,
		ID:      id,
	}
}

// NewParseError creates a new parse error
func NewParseError(details string, id interface{}) *JsonRpcError {
	message := "Parse error"
	if details != "" {
		message += ": " + details
	}
	return NewError(ErrParse, message, nil, id)
}

// NewInvalidRequestError creates a new invalid request error
func NewInvalidRequestError(details string, id interface{}) *JsonRpcError {
	message := "Invalid request"
	if details != "" {
		message += ": " + details
	}
	return NewError(ErrInvalidRequest, message, nil, id)
}

// NewMethodNotFoundError creates a new method not found error
func NewMethodNotFoundError(method string, id interface{}) *JsonRpcError {
	return NewError(ErrMethodNotFound, "Method not found: "+method, nil, id)
}

// NewInvalidParamsError creates a new invalid params error
func NewInvalidParamsError(details string, id interface{}) *JsonRpcError {
	message := "Invalid params"
	if details != "" {
		message += ": " + details
	}
	return NewError(ErrInvalidParams, message, nil, id)
}

// NewInternalError creates a new internal error
func NewInternalError(details string, id interface{}) *JsonRpcError {
	message := "Internal error"
	if details != "" {
		message += ": " + details
	}
	return NewError(ErrInternal, message, nil, id)
}

// NewServerError creates a new server error
func NewServerError(code int, message string, data interface{}, id interface{}) *JsonRpcError {
	if code >= -31999 && code <= -32000 {
		return NewError(code, message, data, id)
	}
	// If the code is not in the server error range, use the standard server error code
	return NewError(ErrServer, message, data, id)
}

// CreateErrorResponse creates a JSON-RPC error response message
// This is kept for backward compatibility
func CreateErrorResponse(id interface{}, err *JsonRpcError) JSONRPCMessage {
	// Update the ID in the error if it's not already set
	if err.ID == nil {
		err.ID = id
	}
	return err.ToResponse()
}

// IsJsonRpcError checks if an error is a JSON-RPC error
func IsJsonRpcError(err error) (*JsonRpcError, bool) {
	if err == nil {
		return nil, false
	}

	rpcErr, ok := err.(*JsonRpcError)
	return rpcErr, ok
}
