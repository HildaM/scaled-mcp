package httphandlers

import (
	"fmt"
	"net/http"

	actors2 "github.com/traego/scaled-mcp/internal/actors"
	"github.com/traego/scaled-mcp/internal/channels"
	"github.com/traego/scaled-mcp/pkg/utils"
)

func (h *MCPHandler) SSEGetWithBasePath(basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.SSEGetFunc(w, r, basePath)
	}
}

// This is backwards compatibility for 2024 SSE sessions, for server to client messages
func (h *MCPHandler) HandleSSEGet(w http.ResponseWriter, r *http.Request) {
	h.SSEGetFunc(w, r, "")
}

func (h *MCPHandler) SSEGetFunc(w http.ResponseWriter, r *http.Request, basePath string) {
	// I think this is easy...spin up the death watcher, spin up the connection watcher, wait for death to come
	ctx := r.Context() // TODO Add logging details around these

	// err will be reused throughout this function
	var err error

	// Attempt to retrieve the session ID from cookie (set during initial connection)
	var sessionId string
	cookie, cerr := r.Cookie("mcp_session_id")
	if cerr == nil && cookie != nil && cookie.Value != "" {
		sessionId = cookie.Value
	} else {
		// Fallback to generating a fresh secure session ID
		var gerr error
		sessionId, gerr = utils.GenerateSecureID(20)
		if gerr != nil {
			handleError(w, gerr, "")
			return
		}
	}

	san := utils.GetSessionActorName(sessionId)

	// Ensure the session actor exists; spawn only if we don't find it running
	_, existingPid, _ := h.actorSystem.ActorOf(ctx, san)
	if existingPid == nil {
		sa := actors2.NewMcpSessionStateMachine(h.serverInfo, sessionId)
		_, err = h.actorSystem.Spawn(ctx, san, sa)
		if err != nil {
			handleError(w, err, "")
			return
		}
	}

	// Create an SSE channel for communication
	channel := channels.NewSSEChannel(w, r, sessionId)

	cca := actors2.NewClientConnectionActor(h.config, sessionId, nil, channel, true, true, basePath)
	clientActorName := fmt.Sprintf("%s-client", sessionId)
	clientActor, err := h.actorSystem.Spawn(ctx, clientActorName, cca)
	if err != nil {
		respErr := fmt.Errorf("error spawning sse session: %w", err)
		handleError(w, respErr, "")
	}

	_, dc, err := actors2.SpawnDeathWatcher(ctx, h.actorSystem, clientActor)
	if err != nil {
		handleError(w, err, "")
	}

	select {
	case <-dc:
	case <-channel.Done:
	}
}
