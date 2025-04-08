package config

import (
	"context"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/proto/mcppb"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/protocol"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/resources"
)

type McpServerInfo interface {
	GetFeatureRegistry() resources.FeatureRegistry
	GetServerCapabilities() protocol.ServerCapabilities
	GetServerConfig() *ServerConfig
	GetExecutors() MethodHandler
}

type MethodHandler interface {
	CanHandleMethod(method string) bool
	HandleMethod(ctx context.Context, method string, req *mcppb.JsonRpcRequest) (*mcppb.JsonRpcResponse, error)
}
