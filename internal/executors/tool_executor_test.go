package executors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traego/scaled-mcp/pkg/config"
	"github.com/traego/scaled-mcp/pkg/proto/mcppb"
	"github.com/traego/scaled-mcp/pkg/protocol"
	"github.com/traego/scaled-mcp/pkg/resources"
)

// TestToolRegistry is an in-memory implementation of resources.ToolRegistry for testing
type TestToolRegistry struct {
	Tools map[string]protocol.Tool
	Calls map[string]interface{}
}

func NewTestToolRegistry() *TestToolRegistry {
	return &TestToolRegistry{
		Tools: make(map[string]protocol.Tool),
		Calls: make(map[string]interface{}),
	}
}

func (r *TestToolRegistry) GetTool(ctx context.Context, name string) (protocol.Tool, error) {
	tool, ok := r.Tools[name]
	if !ok {
		return protocol.Tool{}, resources.ErrToolNotFound
	}
	return tool, nil
}

func (r *TestToolRegistry) ListTools(ctx context.Context, opts protocol.ToolListOptions) (protocol.ToolListResult, error) {
	var tools []protocol.Tool
	for _, tool := range r.Tools {
		tools = append(tools, tool)
	}
	return protocol.ToolListResult{
		Tools:      tools,
		NextCursor: "",
	}, nil
}

func (r *TestToolRegistry) CallTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	_, ok := r.Tools[name]
	if !ok {
		return nil, resources.ErrToolNotFound
	}

	// Store the call for verification
	r.Calls[name] = params

	// Return a ToolCallResult with text content
	textContent := protocol.NewTextContent("Tool execution successful")
	return protocol.NewToolCallResult([]protocol.ToolCallContent{textContent}, false), nil
}

// TestServerInfo is an in-memory implementation of config.McpServerInfo for testing
type TestServerInfo struct {
	FeatureRegistry resources.FeatureRegistry
	ServerCaps      protocol.ServerCapabilities
	ServerConfig    *config.ServerConfig
}

func NewTestServerInfo() *TestServerInfo {
	toolRegistry := NewTestToolRegistry()

	return &TestServerInfo{
		FeatureRegistry: resources.FeatureRegistry{
			ToolRegistry: toolRegistry,
		},
		ServerCaps: protocol.ServerCapabilities{
			Tools: &protocol.ToolsServerCapability{
				ListChanged: true,
			},
			Experimental: map[string]interface{}{
				"version": "1.0.0",
			},
		},
		ServerConfig: &config.ServerConfig{
			ProtocolVersion: protocol.ProtocolVersion20250326,
		},
	}
}

func (s *TestServerInfo) GetFeatureRegistry() resources.FeatureRegistry {
	return s.FeatureRegistry
}

func (s *TestServerInfo) GetServerCapabilities() protocol.ServerCapabilities {
	return s.ServerCaps
}

func (s *TestServerInfo) GetServerConfig() *config.ServerConfig {
	return s.ServerConfig
}

func (s *TestServerInfo) GetExecutors() config.MethodHandler {
	return nil // Not needed for these tests
}

func (s *TestServerInfo) GetAuthHandler() config.AuthHandler {
	return nil
}

func (s *TestServerInfo) GetTraceHandler() config.TraceHandler {
	return nil
}

func TestToolExecutor_CanHandleMethod(t *testing.T) {
	// Create a test server info
	serverInfo := NewTestServerInfo()

	// Create a tool executor
	executor := NewToolExecutor(serverInfo)

	// Test that the executor can handle known methods
	assert.True(t, executor.CanHandleMethod("tools/list"))
	assert.True(t, executor.CanHandleMethod("tools/get"))
	assert.True(t, executor.CanHandleMethod("tools/call"))

	// Test that the executor cannot handle unknown methods
	assert.False(t, executor.CanHandleMethod("unknown/method"))
	assert.False(t, executor.CanHandleMethod("tools/unknown"))
}

func TestToolExecutor_HandleMethod_List(t *testing.T) {
	// Create a test server info with tools
	serverInfo := NewTestServerInfo()

	// Get the tool registry
	toolRegistry, ok := serverInfo.FeatureRegistry.ToolRegistry.(*TestToolRegistry)
	require.True(t, ok)

	// Add some test tools
	toolRegistry.Tools["test-tool"] = protocol.Tool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: protocol.InputSchema{
			Type: "object",
			Properties: map[string]protocol.SchemaProperty{
				"param1": {
					Type:        "string",
					Description: "A test parameter",
				},
			},
			Required: []string{"param1"},
		},
	}

	// Create a tool executor
	executor := NewToolExecutor(serverInfo)

	// Test the list method
	ctx := context.Background()
	req := &mcppb.JsonRpcRequest{
		Jsonrpc: "2.0",
		Id: &mcppb.JsonRpcRequest_StringId{
			StringId: "1",
		},
		Method: "tools/list",
	}

	resp, err := executor.HandleMethod(ctx, "tools/list", req)
	require.NoError(t, err)
	assert.Equal(t, "1", resp.GetStringId())
	assert.Equal(t, "2.0", resp.Jsonrpc)

	// Parse the result
	var result map[string]interface{}
	err = json.Unmarshal([]byte(resp.GetResultJson()), &result)
	require.NoError(t, err)

	// Verify the tools are in the result
	tools, ok := result["tools"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, tools, 1)

	// Verify the tool properties
	tool := tools[0].(map[string]interface{})
	assert.Equal(t, "test-tool", tool["name"])
	assert.Equal(t, "A test tool", tool["description"])
}

func TestToolExecutor_HandleMethod_Get(t *testing.T) {
	// Create a test server info with tools
	serverInfo := NewTestServerInfo()

	// Get the tool registry
	toolRegistry, ok := serverInfo.FeatureRegistry.ToolRegistry.(*TestToolRegistry)
	require.True(t, ok)

	// Add a test tool
	toolRegistry.Tools["test-tool"] = protocol.Tool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: protocol.InputSchema{
			Type: "object",
			Properties: map[string]protocol.SchemaProperty{
				"param1": {
					Type:        "string",
					Description: "A test parameter",
				},
			},
			Required: []string{"param1"},
		},
	}

	// Create a tool executor
	executor := NewToolExecutor(serverInfo)

	// Test the get method with a valid tool
	ctx := context.Background()
	paramsBytes, _ := json.Marshal(map[string]interface{}{
		"name": "test-tool",
	})
	req := &mcppb.JsonRpcRequest{
		Jsonrpc: "2.0",
		Id: &mcppb.JsonRpcRequest_StringId{
			StringId: "1",
		},
		Method:     "tools/get",
		ParamsJson: string(paramsBytes),
	}

	resp, err := executor.HandleMethod(ctx, "tools/get", req)
	require.NoError(t, err)
	assert.Equal(t, "1", resp.GetStringId())
	assert.Equal(t, "2.0", resp.Jsonrpc)

	// Parse the result
	var result map[string]interface{}
	err = json.Unmarshal([]byte(resp.GetResultJson()), &result)
	require.NoError(t, err)

	// Verify the tool properties
	assert.Equal(t, "test-tool", result["name"])
	assert.Equal(t, "A test tool", result["description"])

	// Test the get method with an invalid tool
	paramsBytes, _ = json.Marshal(map[string]interface{}{
		"name": "non-existent-tool",
	})
	req = &mcppb.JsonRpcRequest{
		Jsonrpc: "2.0",
		Id: &mcppb.JsonRpcRequest_StringId{
			StringId: "2",
		},
		Method:     "tools/get",
		ParamsJson: string(paramsBytes),
	}

	resp, err = executor.HandleMethod(ctx, "tools/get", req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Test the get method with missing name parameter
	paramsBytes, _ = json.Marshal(map[string]interface{}{})
	req = &mcppb.JsonRpcRequest{
		Jsonrpc: "2.0",
		Id: &mcppb.JsonRpcRequest_StringId{
			StringId: "3",
		},
		Method:     "tools/get",
		ParamsJson: string(paramsBytes),
	}

	resp, err = executor.HandleMethod(ctx, "tools/get", req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	// Test the get method with empty name parameter
	paramsBytes, _ = json.Marshal(map[string]interface{}{
		"name": "",
	})
	req = &mcppb.JsonRpcRequest{
		Jsonrpc: "2.0",
		Id: &mcppb.JsonRpcRequest_StringId{
			StringId: "4",
		},
		Method:     "tools/get",
		ParamsJson: string(paramsBytes),
	}

	resp, err = executor.HandleMethod(ctx, "tools/get", req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestToolExecutor_HandleMethod_Call(t *testing.T) {
	// Create a test server info with tools
	serverInfo := NewTestServerInfo()

	// Get the tool registry
	toolRegistry, ok := serverInfo.FeatureRegistry.ToolRegistry.(*TestToolRegistry)
	require.True(t, ok)

	// Add some test tools
	toolRegistry.Tools["test-tool"] = protocol.Tool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: protocol.InputSchema{
			Type: "object",
			Properties: map[string]protocol.SchemaProperty{
				"param1": {
					Type:        "string",
					Description: "A test parameter",
				},
			},
			Required: []string{"param1"},
		},
	}

	// Create a tool executor
	executor := NewToolExecutor(serverInfo)

	// Test context
	ctx := context.Background()

	t.Run("Call with valid parameters", func(t *testing.T) {
		paramsBytes, _ := json.Marshal(map[string]interface{}{
			"name": "test-tool",
			"arguments": map[string]interface{}{
				"param1": "test-value",
			},
		})
		req := &mcppb.JsonRpcRequest{
			Jsonrpc: "2.0",
			Id: &mcppb.JsonRpcRequest_StringId{
				StringId: "1",
			},
			Method:     "tools/call",
			ParamsJson: string(paramsBytes),
		}

		resp, err := executor.HandleMethod(ctx, "tools/call", req)
		require.NoError(t, err)
		assert.Equal(t, "1", resp.GetStringId())
		assert.Equal(t, "2.0", resp.Jsonrpc)

		// Parse the result
		var result map[string]interface{}
		err = json.Unmarshal([]byte(resp.GetResultJson()), &result)
		require.NoError(t, err)

		// Verify the result structure matches ToolCallResult
		content, ok := result["content"].([]interface{})
		require.True(t, ok, "Result should contain a 'content' array")
		require.Equal(t, 1, len(content), "Should have 1 content item")

		// Check the content item
		contentItem, ok := content[0].(map[string]interface{})
		require.True(t, ok, "Content item should be a map")
		assert.Equal(t, "text", contentItem["type"], "Content type should be 'text'")
		assert.Equal(t, "Tool execution successful", contentItem["text"], "Content text should match")

		// Check isError field
		isError, ok := result["isError"].(bool)
		require.True(t, ok, "Result should contain an 'isError' boolean")
		assert.False(t, isError, "isError should be false")

		// Verify the call was recorded with the correct parameters
		params, ok := toolRegistry.Calls["test-tool"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test-value", params["param1"])
	})

	t.Run("Call with non-existent tool", func(t *testing.T) {
		paramsBytes, _ := json.Marshal(map[string]interface{}{
			"name": "non-existent-tool",
			"arguments": map[string]interface{}{
				"param1": "test-value",
			},
		})
		req := &mcppb.JsonRpcRequest{
			Jsonrpc: "2.0",
			Id: &mcppb.JsonRpcRequest_StringId{
				StringId: "2",
			},
			Method:     "tools/call",
			ParamsJson: string(paramsBytes),
		}

		resp, err := executor.HandleMethod(ctx, "tools/call", req)
		require.NoError(t, err, "Should not return an error for non-existent tool")

		// Parse the result
		var result map[string]interface{}
		err = json.Unmarshal([]byte(resp.GetResultJson()), &result)
		require.NoError(t, err)

		// Verify this is an error result
		isError, ok := result["isError"].(bool)
		require.True(t, ok, "Result should contain an 'isError' boolean")
		assert.True(t, isError, "isError should be true for non-existent tool")

		// Verify the content contains an error message
		content, ok := result["content"].([]interface{})
		require.True(t, ok, "Result should contain a 'content' array")
		require.Equal(t, 1, len(content), "Should have 1 content item")

		contentItem, ok := content[0].(map[string]interface{})
		require.True(t, ok, "Content item should be a map")
		assert.Equal(t, "text", contentItem["type"], "Content type should be 'text'")
		assert.Contains(t, contentItem["text"], "Error calling non-existent-tool", "Content should contain error message")
	})

	t.Run("Call with missing name parameter", func(t *testing.T) {
		paramsBytes, _ := json.Marshal(map[string]interface{}{
			"arguments": map[string]interface{}{
				"param1": "test-value",
			},
		})
		req := &mcppb.JsonRpcRequest{
			Jsonrpc: "2.0",
			Id: &mcppb.JsonRpcRequest_StringId{
				StringId: "3",
			},
			Method:     "tools/call",
			ParamsJson: string(paramsBytes),
		}

		resp, err := executor.HandleMethod(ctx, "tools/call", req)
		assert.Error(t, err, "Should return error for missing name parameter")
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "tool name is required")
	})

	t.Run("Call with empty name parameter", func(t *testing.T) {
		paramsBytes, _ := json.Marshal(map[string]interface{}{
			"name": "",
			"arguments": map[string]interface{}{
				"param1": "test-value",
			},
		})
		req := &mcppb.JsonRpcRequest{
			Jsonrpc: "2.0",
			Id: &mcppb.JsonRpcRequest_StringId{
				StringId: "4",
			},
			Method:     "tools/call",
			ParamsJson: string(paramsBytes),
		}

		resp, err := executor.HandleMethod(ctx, "tools/call", req)
		assert.Error(t, err, "Should return error for empty name parameter")
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "tool name must be a non-empty string")
	})

	t.Run("Call with invalid arguments parameter", func(t *testing.T) {
		paramsBytes, _ := json.Marshal(map[string]interface{}{
			"name":      "test-tool",
			"arguments": "not-an-object",
		})
		req := &mcppb.JsonRpcRequest{
			Jsonrpc: "2.0",
			Id: &mcppb.JsonRpcRequest_StringId{
				StringId: "5",
			},
			Method:     "tools/call",
			ParamsJson: string(paramsBytes),
		}

		resp, err := executor.HandleMethod(ctx, "tools/call", req)
		assert.Error(t, err, "Should return error for invalid arguments parameter")
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "parameters must be an object")
	})
}

func TestToolExecutor_HandleMethod_InvalidMethod(t *testing.T) {
	// Create a test server info
	serverInfo := NewTestServerInfo()

	// Create a tool executor
	executor := NewToolExecutor(serverInfo)

	// Test handling an invalid method
	ctx := context.Background()
	req := &mcppb.JsonRpcRequest{
		Jsonrpc: "2.0",
		Id: &mcppb.JsonRpcRequest_StringId{
			StringId: "1",
		},
		Method: "tools/invalid",
	}

	resp, err := executor.HandleMethod(ctx, "tools/invalid", req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestToolExecutor_HandleMethod_InvalidParams(t *testing.T) {
	// Create a test server info
	serverInfo := NewTestServerInfo()

	// Create a tool executor
	executor := NewToolExecutor(serverInfo)

	// Create a request with invalid JSON in params
	req := &mcppb.JsonRpcRequest{
		Jsonrpc: "2.0",
		Id: &mcppb.JsonRpcRequest_StringId{
			StringId: "1",
		},
		Method:     "tools/list",
		ParamsJson: "{invalid-json",
	}

	// Test handling invalid params
	ctx := context.Background()
	resp, err := executor.HandleMethod(ctx, "tools/list", req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestToolExecutor_HandleMethod_NilToolRegistry(t *testing.T) {
	// Create a test server info with nil tool registry
	serverInfo := NewTestServerInfo()
	serverInfo.FeatureRegistry = resources.FeatureRegistry{
		ToolRegistry: nil,
	}

	// Create a tool executor
	executor := NewToolExecutor(serverInfo)

	// Test handling a method with nil tool registry
	ctx := context.Background()
	req := &mcppb.JsonRpcRequest{
		Jsonrpc: "2.0",
		Id: &mcppb.JsonRpcRequest_StringId{
			StringId: "1",
		},
		Method: "tools/list",
	}

	resp, err := executor.HandleMethod(ctx, "tools/list", req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
