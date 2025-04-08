package store

import (
	"context"
	"time"
)

type Session struct {
	SessionId           string `json:"session_id"`
	ProtocolVersion     string `json:"protocol_version"`
	LongLivedConnection bool   `json:"long_lived_connection"`
}

// SessionStore defines the interface for session storage
type SessionStore interface {
	// RegisterSession registers a new session with the given session ID and actor ID
	RegisterSession(ctx context.Context, session Session, ttl time.Duration) error

	// GetSession retrieves the actor ID for a session
	GetSession(ctx context.Context, sessionId string) (Session, error)

	// RemoveSession removes a session
	RemoveSession(ctx context.Context, sessionId string) error

	// RefreshSession refreshes the TTL for a session
	RefreshSession(ctx context.Context, sessionId string, ttl time.Duration) error

	// SessionExists checks if a session exists
	SessionExists(ctx context.Context, sessionId string) (bool, error)

	// Close closes the session store
	Close() error
}
