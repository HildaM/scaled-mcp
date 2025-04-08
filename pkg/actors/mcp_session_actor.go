package actors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/traego/scaled-mcp/scaled-mcp-server/internal/utils"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/config"
	"log/slog"
	"time"

	"github.com/tochemey/goakt/v3/actor"
	"github.com/tochemey/goakt/v3/goaktpb"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/proto/mcppb"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/protocol"
)

// McpSessionActor represents an actor that handles MCP session requests
type McpSessionActor struct {
	// Session ID
	sessionId string

	serverInfo config.McpServerInfo

	// MCP protocol state
	initialized     bool
	protocolVersion string
	clientInfo      protocol.ClientInfo

	// Last activity time
	lastActivity time.Time

	// Session timeout duration
	sessionTimeout time.Duration

	clientConnectionActors map[string]*actor.PID
}

// Ensure McpSessionActor implements the Actor interface
var _ actor.Actor = (*McpSessionActor)(nil)

// NewMcpSessionActor creates a new MCP session actor
func NewMcpSessionActor(serverInfo config.McpServerInfo, sessionId string) actor.Actor {
	return &McpSessionActor{
		sessionId:              sessionId,
		serverInfo:             serverInfo,
		initialized:            false,
		sessionTimeout:         1 * time.Minute,
		lastActivity:           time.Now(),
		clientConnectionActors: make(map[string]*actor.PID),
	}
}

// Initialize sets up the session actor
func (a *McpSessionActor) Initialize() {
	a.lastActivity = time.Now()
	a.initialized = true
}

// PreStart is called when the actor is started
func (a *McpSessionActor) PreStart(ctx context.Context) error {
	slog.InfoContext(ctx, "Starting MCP session actor", "session_id", a.sessionId)
	return nil
}

// Receive handles messages sent to the actor
func (a *McpSessionActor) Receive(ctx *actor.ReceiveContext) {
	a.lastActivity = time.Now()

	// Get the message
	message := ctx.Message()

	// Handle different message types
	switch msg := message.(type) {
	case *goaktpb.PostStart:
		ctx.Logger().Debug("mcp session actor finished starting, sending cleanup message", "session_id", a.sessionId)
		err := ctx.ActorSystem().ScheduleOnce(ctx.Context(), new(mcppb.TryCleanupPreInitialized), ctx.Self(), 10*time.Second)
		if err != nil {
			slog.ErrorContext(ctx.Context(), "error scheduling pre-initialized mcp session", "err", err)
		}
	case *mcppb.RegisterConnection:
		//sender := ctx.RemoteSender()
		sender := ctx.Sender()
		a.clientConnectionActors[msg.GetConnectionId()] = sender
		ctx.Response(&mcppb.RegisterConnectionResponse{Success: true})
	//case *mcppb.JsonRpcRequest:
	case *mcppb.WrappedRequest:
		// Handle the request based on the method
		switch msg.Request.Method {
		case "initialize":
			response := a.handleInitialize(ctx.Context(), msg.Request)
			a.sendResponse(ctx, msg, response)
		case "notifications/initialized":
			a.initialized = true
		case "shutdown":
			response := a.handleShutdown(ctx.Context(), msg.Request)
			a.sendResponse(ctx, msg, response)
		default:
			// If not initialized, return an error
			if !a.initialized {
				ctx.Logger().Debug("mcp session actor got non-lifecycle message before being initialized", "session_id", a.sessionId)
				errorResp := utils.CreateErrorResponse(msg.Request, -32002, "Server not initialized", nil)
				if ctx.Sender() != nil {
					err := ctx.Sender().Tell(ctx.Context(), ctx.Sender(), errorResp)
					if err != nil {
						slog.ErrorContext(ctx.Context(), "error messaging failure", "session_id", a.sessionId)
					}
				}

				return
			}

			// Handle non-lifecycle messages
			response, err := a.handleNonLifecycleRequest(ctx.Context(), msg.Request.Id, msg.Request)
			if err != nil {
				hndlErr := protocol.NewInternalError("problem handling message", msg.Request.Id)
				errorResp := utils.CreateErrorResponseFromJsonRpcError(msg.Request, hndlErr)
				if ctx.Sender() != nil {
					err := ctx.Sender().Tell(ctx.Context(), ctx.Sender(), errorResp)
					if err != nil {
						slog.ErrorContext(ctx.Context(), "error messaging failure", "session_id", a.sessionId)
					}
				}
				slog.ErrorContext(ctx.Context(), "problem handling non-lifecycle message", "session_id", a.sessionId, "err", err)
				return
			}

			a.sendResponse(ctx, msg, response)
		}
	case *mcppb.TryCleanupPreInitialized:
		// Basically, I'm trying to limit session that start initialization but never finish in a minimal amount of time
		// If this succeeds, we should then schedule the session TTL check
		if !a.initialized {
			err := ctx.Self().Shutdown(ctx.Context())
			if err != nil {
				slog.ErrorContext(ctx.Context(), "error in shutdown", "session_id", a.sessionId)
			}
		}
	default:
		// Mark message as unhandled
		ctx.Unhandled()
		slog.WarnContext(ctx.Context(), "Received unknown message type",
			"session_id", a.sessionId,
			"message_type", fmt.Sprintf("%T", msg))
	}
}

func (a *McpSessionActor) sendResponse(ctx *actor.ReceiveContext, wrappedMsg *mcppb.WrappedRequest, response *mcppb.JsonRpcResponse) {
	if wrappedMsg.IsAsk {
		ctx.Response(response)
	} else {
		rc, ok := a.clientConnectionActors[wrappedMsg.RespondToConnectionId]
		if !ok {
			slog.ErrorContext(ctx.Context(), "could not find actor to respond for connection to", "connectionId", wrappedMsg.RespondToConnectionId)
		}
		ctx.Tell(rc, response)
	}
}

// PostStop is called when the actor is stopped
func (a *McpSessionActor) PostStop(ctx context.Context) error {
	slog.InfoContext(ctx, "Stopping MCP session actor", "session_id", a.sessionId)
	return nil
}

// handleInitialize processes an initialize request
func (a *McpSessionActor) handleInitialize(ctx context.Context, req *mcppb.JsonRpcRequest) *mcppb.JsonRpcResponse {
	slog.InfoContext(ctx, "Handling initialize request", "session_id", a.sessionId)

	// Create base response
	response := &mcppb.JsonRpcResponse{
		Jsonrpc: "2.0",
	}

	// Copy the ID from the request
	switch id := req.Id.(type) {
	case *mcppb.JsonRpcRequest_IntId:
		response.Id = &mcppb.JsonRpcResponse_IntId{IntId: id.IntId}
	case *mcppb.JsonRpcRequest_StringId:
		response.Id = &mcppb.JsonRpcResponse_StringId{StringId: id.StringId}
	}

	// Parse the parameters
	if req.ParamsJson == "" {
		return utils.CreateErrorResponse(req, -32602, "Invalid params: missing parameters", nil)
	}

	// Parse into our type
	var params protocol.InitializeParams
	if err := json.Unmarshal([]byte(req.ParamsJson), &params); err != nil {
		return utils.CreateErrorResponse(req, -32602, "Invalid params: "+err.Error(), nil)
	}

	// Check protocol version
	supportedVersions := []string{"2024-11-05", "2025-03"}
	versionSupported := false
	for _, v := range supportedVersions {
		if params.ProtocolVersion == v {
			versionSupported = true
			break
		}
	}

	if !versionSupported {
		errorData := map[string]interface{}{
			"supportedVersions": supportedVersions,
		}
		return utils.CreateErrorResponse(req, -32602, "Unsupported protocol version", errorData)
	}

	// Store client info and capabilities
	a.initialized = true
	a.protocolVersion = params.ProtocolVersion
	a.clientInfo = params.ClientInfo

	// Create the result
	result := protocol.InitializeResult{
		ProtocolVersion: params.ProtocolVersion,
		ServerInfo: protocol.ServerInfo{
			Name:    "scaled-mcp-server",
			Version: "1.0.0", // TODO: Get from config
		},
		Capabilities: a.serverInfo.GetServerCapabilities(),
		SessionID:    a.sessionId,
	}

	// Convert result to JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return utils.CreateErrorResponse(req, -32603, "Internal error: "+err.Error(), nil)
	}

	response.Response = &mcppb.JsonRpcResponse_ResultJson{
		ResultJson: string(resultJSON),
	}
	return response
}

// handleShutdown processes a shutdown request
func (a *McpSessionActor) handleShutdown(ctx context.Context, req *mcppb.JsonRpcRequest) *mcppb.JsonRpcResponse {
	// TODO need to comm out to all the connection listeners that I'm shutting down

	// Mark the session as not initialized
	a.initialized = false

	// Create base response
	response := &mcppb.JsonRpcResponse{
		Jsonrpc: "2.0",
	}

	// Copy the ID from the request
	switch id := req.Id.(type) {
	case *mcppb.JsonRpcRequest_IntId:
		response.Id = &mcppb.JsonRpcResponse_IntId{IntId: id.IntId}
	case *mcppb.JsonRpcRequest_StringId:
		response.Id = &mcppb.JsonRpcResponse_StringId{StringId: id.StringId}
	}

	// Create empty result
	response.Response = &mcppb.JsonRpcResponse_ResultJson{
		ResultJson: "{}",
	}

	return response
}

// handleNonLifecycleRequest processes other MCP requests
func (a *McpSessionActor) handleNonLifecycleRequest(ctx context.Context, messageId interface{}, req *mcppb.JsonRpcRequest) (*mcppb.JsonRpcResponse, error) {
	// Create base response
	response := &mcppb.JsonRpcResponse{
		Jsonrpc: "2.0",
	}

	// Copy the ID from the request
	switch id := req.Id.(type) {
	case *mcppb.JsonRpcRequest_IntId:
		response.Id = &mcppb.JsonRpcResponse_IntId{IntId: id.IntId}
	case *mcppb.JsonRpcRequest_StringId:
		response.Id = &mcppb.JsonRpcResponse_StringId{StringId: id.StringId}
	}

	// Check if this is a tool-related method
	exc := a.serverInfo.GetExecutors()
	if a.serverInfo.GetExecutors().CanHandleMethod(req.Method) {
		resp, err := exc.HandleMethod(ctx, req.Method, req)
		if err != nil {
			return nil, actor.NewInternalError(err)
		}
		return resp, nil
	} else {
		return nil, protocol.NewMethodNotFoundError(req.Method, messageId)
	}
}
