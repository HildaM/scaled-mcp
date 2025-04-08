package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisSessionStore implements SessionStore using Redis
type RedisSessionStore struct {
	client *redis.Client
	prefix string
}

// NewRedisSessionStore creates a new Redis session store
func NewRedisSessionStore(addresses []string, prefix string) (*RedisSessionStore, error) {
	// Use the first address for now (in a production environment, you might want to use a Redis cluster)
	addr := addresses[0]

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", addr, err)
	}

	slog.Info("Connected to Redis", "address", addr)

	return &RedisSessionStore{
		client: client,
		prefix: prefix,
	}, nil
}

// sessionKey returns the Redis key for a session
func (s *RedisSessionStore) sessionKey(sessionId string) string {
	return fmt.Sprintf("%s:session:%s", s.prefix, sessionId)
}

// RegisterSession registers a new session with the given session ID
func (s *RedisSessionStore) RegisterSession(ctx context.Context, session Session, ttl time.Duration) error {
	sessionKey := s.sessionKey(session.SessionId)

	// Serialize the session object
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to serialize session %s: %w", session.SessionId, err)
	}

	// Use a pipeline to set both values atomically
	pipe := s.client.Pipeline()
	pipe.Set(ctx, sessionKey, sessionData, ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to register session %s: %w", session.SessionId, err)
	}

	slog.Debug("Registered session",
		"session_id", session.SessionId,
		"long_lived_connectino", session.LongLivedConnection,
		"ttl", ttl)
	return nil
}

// GetSession retrieves the session for a session ID
func (s *RedisSessionStore) GetSession(ctx context.Context, sessionId string) (Session, error) {
	sessionKey := s.sessionKey(sessionId)

	// Get the session data
	sessionData, err := s.client.Get(ctx, sessionKey).Result()
	if err == redis.Nil {
		return Session{}, fmt.Errorf("session %s not found", sessionId)
	} else if err != nil {
		return Session{}, fmt.Errorf("failed to get session %s: %w", sessionId, err)
	}

	// Deserialize the session object
	var session Session
	err = json.Unmarshal([]byte(sessionData), &session)
	if err != nil {
		return Session{}, fmt.Errorf("failed to deserialize session %s: %w", sessionId, err)
	}

	return session, nil
}

// RemoveSession removes a session
func (s *RedisSessionStore) RemoveSession(ctx context.Context, sessionId string) error {
	sessionKey := s.sessionKey(sessionId)

	// Use a pipeline to delete both keys atomically
	pipe := s.client.Pipeline()
	pipe.Del(ctx, sessionKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove session %s: %w", sessionId, err)
	}

	slog.Debug("Removed session", "session_id", sessionId)
	return nil
}

// RefreshSession refreshes the TTL for a session
func (s *RedisSessionStore) RefreshSession(ctx context.Context, sessionId string, ttl time.Duration) error {
	sessionKey := s.sessionKey(sessionId)

	// Get the current session data
	session, err := s.GetSession(ctx, sessionId)
	if err != nil {
		return err
	}

	// Serialize the session object
	sessionData, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to serialize session %s: %w", sessionId, err)
	}

	// Use a pipeline to refresh both keys atomically
	pipe := s.client.Pipeline()
	pipe.Set(ctx, sessionKey, sessionData, ttl)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh session %s: %w", sessionId, err)
	}

	slog.Debug("Refreshed session", "session_id", sessionId, "ttl", ttl)
	return nil
}

// SessionExists checks if a session exists
func (s *RedisSessionStore) SessionExists(ctx context.Context, sessionId string) (bool, error) {
	sessionKey := s.sessionKey(sessionId)

	exists, err := s.client.Exists(ctx, sessionKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check if session %s exists: %w", sessionId, err)
	}

	return exists > 0, nil
}

// Close closes the Redis client
func (s *RedisSessionStore) Close() error {
	return s.client.Close()
}

var _ SessionStore = (*RedisSessionStore)(nil)
