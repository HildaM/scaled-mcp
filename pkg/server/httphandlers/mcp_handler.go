package httphandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/tochemey/goakt/v3/actor"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/config"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/protocol"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/session/store"
)

// MCPHandler handles MCP protocol requests
type MCPHandler struct {
	config       *config.ServerConfig
	actorSystem  actor.ActorSystem
	sessionStore store.SessionStore
	serverInfo   config.McpServerInfo
}

// NewMCPHandler creates a new MCP handler
func NewMCPHandler(config *config.ServerConfig, actorSystem actor.ActorSystem, sessionStore store.SessionStore, serverInfo config.McpServerInfo) *MCPHandler {
	return &MCPHandler{
		config:       config,
		actorSystem:  actorSystem,
		sessionStore: sessionStore,
		serverInfo:   serverInfo,
	}
}

type McpRequest struct {
	IsBatch  bool
	Message  protocol.JSONRPCMessage
	Messages []protocol.JSONRPCMessage
}

func parseMessageRequest(r *http.Request) (McpRequest, error) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return McpRequest{}, fmt.Errorf("failed to read body: %w", err)
	}

	// Parse the request
	var message protocol.JSONRPCMessage
	var messages []protocol.JSONRPCMessage
	var isBatch bool

	// Try to parse as a single message first
	if err := json.Unmarshal(body, &message); err != nil {
		// If that fails, try to parse as a batch
		if err := json.Unmarshal(body, &messages); err != nil {
			return McpRequest{}, fmt.Errorf("failed to parse body: %w", err)
		}
		isBatch = true
	}

	return McpRequest{
		IsBatch:  isBatch,
		Message:  message,
		Messages: messages,
	}, nil
}

func writeMessage(w http.ResponseWriter, messageId interface{}, msg protocol.JSONRPCMessage) {
	responseJSON, err := json.Marshal(msg)
	if err != nil {
		handleError(w, err, msg.ID)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(responseJSON)
	return
}

// handleError processes errors from request handling
// It distinguishes between JSON-RPC errors and other errors
func handleError(w http.ResponseWriter, err error, id interface{}) {
	w.Header().Set("Content-Type", "application/json")

	// Check if it's a JSON-RPC error
	var jsonRpcError *protocol.JsonRpcError
	if errors.As(err, &jsonRpcError) {
		// It's a JSON-RPC error, so we can use its ToResponse method
		// Make sure the ID is set
		if jsonRpcError.ID == nil {
			jsonRpcError.ID = id
		}

		response := jsonRpcError.ToResponse()
		responseJSON, marshalErr := json.Marshal(response)
		if marshalErr != nil {
			// If we can't marshal the error response, fall back to a generic JSON-RPC server error
			slog.Error("Failed to marshal JSON-RPC error response", "error", marshalErr)
			fallbackError := protocol.NewServerError(protocol.ErrServer, "Internal server error", nil, id)
			fallbackJSON, _ := json.Marshal(fallbackError.ToResponse())
			w.WriteHeader(http.StatusOK) // JSON-RPC errors use 200 OK with error in body
			_, _ = w.Write(fallbackJSON)
			return
		}

		w.WriteHeader(http.StatusOK) // JSON-RPC errors use 200 OK with error in body
		_, writeErr := w.Write(responseJSON)
		if writeErr != nil {
			slog.Error("Failed to write JSON-RPC error response", "error", writeErr)
		}
		return
	}

	// It's not a JSON-RPC error, so return a generic 500 error
	slog.Error("Internal server error", "error", err)

	// Create a standard JSON-RPC internal error
	internalError := protocol.NewInternalError(err.Error(), id)
	response := internalError.ToResponse()

	responseJSON, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		// If we can't marshal the error response, fall back to a generic JSON-RPC server error
		slog.Error("Failed to marshal internal error response", "error", marshalErr)
		fallbackError := protocol.NewServerError(protocol.ErrServer, "Internal server error", nil, id)
		fallbackJSON, _ := json.Marshal(fallbackError.ToResponse())
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(fallbackJSON)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	_, writeErr := w.Write(responseJSON)
	if writeErr != nil {
		slog.Error("Failed to write error response", "error", writeErr)
	}
}

// handleBatchError handles errors from batch request processing
func handleBatchError(w http.ResponseWriter, err error) {
	// For batch errors, we typically return a single error response
	// since the batch as a whole failed
	slog.Error("Batch processing error", "error", err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	// Create a generic batch error response
	// Note: For batch errors, we don't have a specific ID
	internalError := protocol.NewInternalError(err.Error(), nil)
	response := internalError.ToResponse()

	responseJSON, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		// If we can't marshal the error response, fall back to a generic JSON-RPC server error
		slog.Error("Failed to marshal batch error response", "error", marshalErr)
		fallbackError := protocol.NewServerError(protocol.ErrServer, "Internal server error", nil, nil)
		fallbackJSON, _ := json.Marshal(fallbackError.ToResponse())
		_, _ = w.Write(fallbackJSON)
		return
	}

	_, writeErr := w.Write(responseJSON)
	if writeErr != nil {
		slog.Error("Failed to write batch error response", "error", writeErr)
	}
}
