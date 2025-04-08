package msghandlers

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"log/slog"
//
//	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/protocol"
//	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/resources"
//)
//
//// ToolHandler handles tool-related MCP requests
//type ToolHandler struct {
//	registry resources.ToolRegistry
//}
//
//// NewToolHandler creates a new tool handler
//func NewToolHandler(registry resources.ToolRegistry) *ToolHandler {
//	return &ToolHandler{
//		registry: registry,
//	}
//}
//
//// HandleListTools handles a request to list resources
//func (h *ToolHandler) HandleListTools(ctx context.Context, params json.RawMessage) (interface{}, error) {
//	var listOpts struct {
//		Cursor string `json:"cursor,omitempty"`
//		Limit  int    `json:"limit,omitempty"`
//	}
//
//	if err := json.Unmarshal(params, &listOpts); err != nil {
//		slog.Error("Failed to unmarshal list resources params", "error", err)
//		return nil, protocol.NewInvalidParamsError("Invalid parameters for resources/list", nil)
//	}
//
//	toolListOpts := resources.ToolListOptions{
//		Cursor: listOpts.Cursor,
//		Limit:  listOpts.Limit,
//	}
//
//	result := h.registry.ListTools(ctx, toolListOpts)
//	return result, nil
//}
//
//// HandleGetTool handles a request to get a specific tool
//func (h *ToolHandler) HandleGetTool(ctx context.Context, params json.RawMessage) (interface{}, error) {
//	var getOpts struct {
//		Name string `json:"name"`
//	}
//
//	if err := json.Unmarshal(params, &getOpts); err != nil {
//		slog.Error("Failed to unmarshal get tool params", "error", err)
//		return nil, protocol.NewInvalidParamsError("Invalid parameters for resources/get", nil)
//	}
//
//	if getOpts.Name == "" {
//		return nil, protocol.NewInvalidParamsError("Tool name is required", nil)
//	}
//
//	tool, found := h.registry.GetTool(ctx, getOpts.Name)
//	if !found {
//		return nil, protocol.NewError(protocol.ErrMethodNotFound, fmt.Sprintf("Tool '%s' not found", getOpts.Name), nil, nil)
//	}
//
//	return tool, nil
//}
//
//// HandleCallTool handles a request to invoke a tool
//func (h *ToolHandler) HandleCallTool(ctx context.Context, params json.RawMessage) (interface{}, error) {
//	var invokeOpts struct {
//		Name       string                 `json:"name"`
//		Parameters map[string]interface{} `json:"parameters"`
//	}
//
//	if err := json.Unmarshal(params, &invokeOpts); err != nil {
//		slog.Error("Failed to unmarshal invoke tool params", "error", err)
//		return nil, protocol.NewInvalidParamsError("Invalid parameters for resources/invoke", nil)
//	}
//
//	if invokeOpts.Name == "" {
//		return nil, protocol.NewInvalidParamsError("Tool name is required", nil)
//	}
//
//	// Ensure parameters is not nil
//	if invokeOpts.Parameters == nil {
//		invokeOpts.Parameters = make(map[string]interface{})
//	}
//
//	result, err := h.registry.CallTool(ctx, invokeOpts.Name, invokeOpts.Parameters)
//	if err != nil {
//		if err == resources.ErrToolNotFound {
//			return nil, protocol.NewError(protocol.ErrMethodNotFound, fmt.Sprintf("Tool '%s' not found", invokeOpts.Name), nil, nil)
//		}
//		if err == resources.ErrInvalidParams {
//			return nil, protocol.NewInvalidParamsError(err.Error(), nil)
//		}
//		slog.Error("Failed to invoke tool", "name", invokeOpts.Name, "error", err)
//		return nil, protocol.NewInternalError(fmt.Sprintf("Failed to invoke tool '%s'", invokeOpts.Name), nil)
//	}
//
//	return result, nil
//}
