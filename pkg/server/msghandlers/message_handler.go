package msghandlers

//
//import (
//	"context"
//
//	"github.com/tochemey/goakt/v3/actor"
//	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/config"
//	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/protocol"
//	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/resources"
//	"github.com/traego/scaled-mcp/scaled-mcp-server/pkg/session/store"
//)
//
//type MessageHandler interface {
//	// HandleMessage handles a single message
//	HandleMessage(ctx context.Context, message protocol.JSONRPCMessage) (*protocol.JSONRPCMessage, error)
//
//	// HandleBatch handles a batch of messages
//	HandleBatch(ctx context.Context, messages []protocol.JSONRPCMessage) error
//}
//
//// requestHandler handles MCP protocol requests
//type requestHandler struct {
//	config       *config.ServerConfig
//	actorSystem  actor.ActorSystem
//	sessionStore store.SessionStore
//	registry     *resources.FeatureRegistry
//}
//
//func NewRequestHandler(config *config.ServerConfig, sessionStore store.SessionStore, actorSystem actor.ActorSystem, registry *resources.FeatureRegistry) MessageHandler {
//	return &requestHandler{
//		config:       config,
//		sessionStore: sessionStore,
//		actorSystem:  actorSystem,
//		registry:     registry,
//	}
//}
//
//func (r requestHandler) HandleMessage(ctx context.Context, message protocol.JSONRPCMessage) (*protocol.JSONRPCMessage, error) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (r requestHandler) HandleBatch(ctx context.Context, messages []protocol.JSONRPCMessage) error {
//	//TODO implement me
//	panic("implement me")
//}
//
//var _ MessageHandler = (*requestHandler)(nil)
//
////
////func (h *requestHandler) HandleSSEStandup(ctx context.Context) error {
////	sessionIdRaw := ctx.Value(protocol.SESSION_ID_CONTEXT_KEY)
////	var err error
////	newSession := false
////	var sessionId string
////	if sessionIdRaw == nil {
////		if h.config.BackwardCompatible20241105 {
////			sessionId, err = utils.GenerateSecureID(20)
////			if err != nil {
////				return fmt.Errorf("could not generate session id: %w", err)
////			}
////			newSession = true
////		} else {
////			return fmt.Errorf("session id is required")
////		}
////	} else {
////		var ok bool
////		if sessionId, ok = sessionIdRaw.(string); !ok {
////			return fmt.Errorf("session id must be a string")
////		}
////	}
////
////	if newSession {
////		// Create the session actor with the protocol version
////		sa := session.NewMcpSessionActor(h.config, sessionId)
////		sessionActorName := utils.GetSessionActorName(sessionId)
////		_, err = h.actorSystem.Spawn(ctx, sessionActorName, sa)
////		if err != nil {
////			return fmt.Errorf("error spawning session: %w", err)
////		}
////	} else {
////		_, _, err = h.actorSystem.ActorOf(ctx, "session")
////		if err != nil {
////			return fmt.Errorf("error getting session actor: %w", err)
////		}
////	}
////	// Create an SSE channel for communication
////	channel := channels.NewSSEChannel(h.w, h.r)
////
////	cca := clients.NewClientConnectionActor(h.config, sessionId, nil, channel, true, true)
////	clientActorName := fmt.Sprintf("%s-client", sessionId)
////	clientActor, err := h.actorSystem.Spawn(ctx, clientActorName, cca)
////	if err != nil {
////		return fmt.Errorf("error spawning session: %w", err)
////	}
////
////	sess := store.Session{
////		SessionId:           sessionId,
////		LongLivedConnection: true,
////	}
////	err = h.sessionStore.RegisterSession(ctx, sess, h.config.Session.TTL)
////	if err != nil {
////		return fmt.Errorf("error registering session: %w", err)
////	}
////
////	dw, dc := utils.NewDeathWatcher()
////	deathWatchName := fmt.Sprintf("%s-client-death-watcher", sessionId)
////	dwa, err := h.actorSystem.Spawn(ctx, deathWatchName, dw)
////	if err != nil {
////		return fmt.Errorf("error spawning death watcher: %w", err)
////	}
////	dwa.Watch(clientActor)
////	<-dc
////	return nil
////}
////
////func (h *requestHandler) HandleMessage(ctx context.Context, message protocol.JSONRPCMessage) (*protocol.JSONRPCMessage, error) {
////	sessionIdRaw := ctx.Value(protocol.SESSION_ID_CONTEXT_KEY)
////	if sessionIdRaw == nil {
////		if message.Method == "initialize" {
////			// In the 2025-03 spec, I think you can initialize with just a post
////			err := h.HandleInitializeMessage(ctx, message)
////			if err != nil {
////				return nil, fmt.Errorf("error handling initialize: %w", err)
////			}
////		} else {
////			return nil, fmt.Errorf("session id required unless initializing")
////		}
////	}
////
////	sessionId, ok := sessionIdRaw.(string)
////	if !ok {
////		return nil, fmt.Errorf("session id must be a string")
////	}
////
////	actorName := utils.GetSessionActorName(sessionId)
////	_, act, err := h.actorSystem.ActorOf(ctx, actorName)
////	if err != nil {
////		return nil, fmt.Errorf("error spawning session: %w", err)
////	}
////	session, err := h.sessionStore.GetSession(ctx, sessionId)
////	if err != nil {
////		return nil, fmt.Errorf("error getting session: %w", err)
////	}
////
////	protoReq, err := protocol.ConvertJSONToProtoRequest(message)
////
////	if session.LongLivedConnection {
////		err = actor.Tell(ctx, act, protoReq)
////		if err != nil {
////			return nil, fmt.Errorf("error sending message: %w", err)
////		}
////		return nil, nil
////	} else {
////		respRaw, err := actor.Ask(ctx, act, protoReq, h.config.RequestTimeout)
////		if err != nil {
////			return nil, fmt.Errorf("error sending message: %w", err)
////		}
////		resp, ok := respRaw.(*mcppb.JsonRpcResponse)
////		if !ok {
////			return nil, fmt.Errorf("error sending message: invalid response")
////		}
////		respMsg, err := protocol.ConvertProtoToJSONResponse(resp)
////		if err != nil {
////			return nil, fmt.Errorf("error parsing response message: %w", err)
////		}
////		return &respMsg, nil
////	}
////}
////
////func (h *requestHandler) HandleBatch(ctx context.Context, messages []protocol.JSONRPCMessage) error {
////	//TODO implement me
////	panic("implement me")
////}
////
////// NewRequestHandler creates a new request handler
////func NewRequestHandler(config *config.ServerConfig, sessionStore store.SessionStore, actorSystem actor.ActorSystem, w http.ResponseWriter, r *http.Request) MessageHandler {
////	return &requestHandler{
////		config:       config,
////		sessionStore: sessionStore,
////		actorSystem:  actorSystem,
////		w:            w,
////		r:            r,
////	}
////}
////
////var _ MessageHandler = (*requestHandler)(nil)
