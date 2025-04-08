package store

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// MemorySessionStore implements SessionStore using an in-memory map
type MemorySessionStore struct {
	sessions     map[string]sessionData
	mu           sync.RWMutex
	cleanupTimer *time.Ticker
	done         chan struct{}
}

// sessionData represents session data with expiration time
type sessionData struct {
	session  Session
	expireAt time.Time
}

// NewMemorySessionStore creates a new in-memory session store
func NewMemorySessionStore() SessionStore {
	store := &MemorySessionStore{
		sessions: make(map[string]sessionData),
		done:     make(chan struct{}),
	}

	// Start a cleanup goroutine to remove expired sessions
	store.cleanupTimer = time.NewTicker(1 * time.Minute)
	go store.cleanupExpiredSessions()

	slog.Info("Created in-memory session store")
	return store
}

// cleanupExpiredSessions periodically removes expired sessions
func (s *MemorySessionStore) cleanupExpiredSessions() {
	for {
		select {
		case <-s.cleanupTimer.C:
			s.mu.Lock()
			now := time.Now()
			for sessionId, data := range s.sessions {
				if now.After(data.expireAt) {
					delete(s.sessions, sessionId)
					slog.Debug("Removed expired session", "session_id", sessionId)
				}
			}
			s.mu.Unlock()
		case <-s.done:
			s.cleanupTimer.Stop()
			return
		}
	}
}

// RegisterSession registers a new session with the given session ID and actor ID
func (s *MemorySessionStore) RegisterSession(ctx context.Context, session Session, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.SessionId] = sessionData{
		session:  session,
		expireAt: time.Now().Add(ttl),
	}

	slog.Debug("Registered session", "session_id", session.SessionId, "ttl", ttl)
	return nil
}

// GetSession retrieves the actor ID for a session
func (s *MemorySessionStore) GetSession(ctx context.Context, sessionId string) (Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.sessions[sessionId]
	if !exists {
		return Session{}, fmt.Errorf("session %s not found", sessionId)
	}

	// Check if the session has expired
	if time.Now().After(data.expireAt) {
		return Session{}, fmt.Errorf("session %s has expired", sessionId)
	}

	return data.session, nil
}

// RemoveSession removes a session
func (s *MemorySessionStore) RemoveSession(ctx context.Context, sessionId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[sessionId]; !exists {
		return fmt.Errorf("session %s not found", sessionId)
	}

	delete(s.sessions, sessionId)
	slog.Debug("Removed session", "session_id", sessionId)
	return nil
}

// RefreshSession refreshes the TTL for a session
func (s *MemorySessionStore) RefreshSession(ctx context.Context, sessionId string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, exists := s.sessions[sessionId]
	if !exists {
		return fmt.Errorf("session %s not found", sessionId)
	}

	data.expireAt = time.Now().Add(ttl)
	s.sessions[sessionId] = data

	slog.Debug("Refreshed session", "session_id", sessionId, "ttl", ttl)
	return nil
}

// SessionExists checks if a session exists
func (s *MemorySessionStore) SessionExists(ctx context.Context, sessionId string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.sessions[sessionId]
	if !exists {
		return false, nil
	}

	// Check if the session has expired
	if time.Now().After(data.expireAt) {
		return false, nil
	}

	return true, nil
}

// Close stops the cleanup goroutine
func (s *MemorySessionStore) Close() error {
	close(s.done)
	return nil
}
