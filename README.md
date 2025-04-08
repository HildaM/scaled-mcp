# Scaled MCP Server

A horizontally scalable MCP (Message Context Protocol) server implementation that supports load-balanced deployments.

## Overview

The Scaled MCP Server is a Go library that implements the MCP 2025-03 specification with support for horizontal scaling. It's designed to be embedded in your application and provides flexible configuration options.

## Features

- **HTTP Transport**: Flexible HTTP transport with main `/mcp` endpoint, optional SSE endpoint, and capabilities negotiation
- **Session Management**: Distributed session management with Redis or in-memory options
- **Actor System**: Uses an actor-based architecture for handling sessions and message routing
- **Horizontal Scaling**: Support for load-balanced deployments across multiple nodes

## Usage

### Basic Server

```go
package main

import (
	"context"
	"log"

	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/config"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/server"
)

func main() {
	// Create a server with default configuration
	cfg := config.DefaultConfig()
	
	// Use in-memory session store for simplicity
	cfg.Session.UseInMemory = true
	
	// Create and start the server
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	
	// Start the server
	if err := srv.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	
	// Wait for termination signal
	<-make(chan struct{})
}
```

### Using an External HTTP Server

You can use your own HTTP server with the MCP transport:

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/config"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/server"
	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/transport"
)

func main() {
	// Create a server with default configuration
	cfg := config.DefaultConfig()
	cfg.Session.UseInMemory = true
	
	// Create the MCP server but don't start the HTTP server
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	
	// Create a custom router
	r := chi.NewRouter()
	
	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	
	// Add CORS middleware - important when using an external server
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	
	// Add your custom routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome to the MCP server!"))
	})
	
	// Create the HTTP transport with the custom router
	httpTransport := transport.NewHTTPTransport(
		cfg,
		srv.GetActorSystem(),
		srv.GetSessionManager(),
		transport.WithExternalRouter(r),
	)
	
	// Start the MCP server without HTTP
	if err := srv.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	
	// Start the HTTP transport
	if err := httpTransport.Start(); err != nil {
		log.Fatalf("Failed to start HTTP transport: %v", err)
	}
	
	// Start your HTTP server
	log.Printf("Starting HTTP server on %s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
```

## Important Notes

### CORS Configuration

When using an external HTTP server with the MCP transport, you need to configure CORS settings on your router. The MCP transport will not apply CORS settings when using an external router, as shown in the example above.

### Session Management

For production deployments, it's recommended to use Redis for session management to support horizontal scaling. The in-memory session store should only be used for development or testing.

## API Documentation

See the [GoDoc](https://pkg.go.dev/github.com/traego/scaled-mcp/scaled-mcp-server) for detailed API documentation.
