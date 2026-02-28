package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/openclaw"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
)

type ChatHandler struct {
	store  *store.Store
	client *openclaw.Client
}

func NewChatHandler(s *store.Store, client *openclaw.Client) *ChatHandler {
	return &ChatHandler{
		store:  s,
		client: client,
	}
}

// Request/Response types
type StartSessionRequest struct {
	InitialMessage string `json:"initial_message,omitempty"`
}

type SendMessageRequest struct {
	Content string `json:"content" validate:"required"`
}

type ChatSessionResponse struct {
	ID                 string  `json:"id"`
	AgentID            string  `json:"agent_id"`
	OpenclawSessionKey *string `json:"openclaw_session_key,omitempty"`
	Status             string  `json:"status"`
	StartedAt          string  `json:"started_at"`
	EndedAt            *string `json:"ended_at,omitempty"`
	MessageCount       int     `json:"message_count"`
}

type ChatMessageResponse struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// Handlers

// StartSession - POST /api/v1/agents/:id/sessions
// Creates a chat session that sends messages directly to the agent
func (h *ChatHandler) StartSession(c echo.Context) error {
	agentID := c.Param("id")

	// Verify agent exists
	_, err := h.store.GetAgent(c.Request().Context(), agentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Agent not found")
	}

	var req StartSessionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// The session key for the agent's direct session
	// Format: agent:<agentId>:main or we can spawn a dedicated chat session
	agentSessionKey := fmt.Sprintf("agent:%s:main", agentID)

	// Save session to database
	session, err := h.store.CreateChatSession(c.Request().Context(), db.CreateChatSessionParams{
		AgentID:            agentID,
		OpenclawSessionKey: sql.NullString{String: agentSessionKey, Valid: true},
		Status:             "active",
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// If initial message provided, send it
	if req.InitialMessage != "" {
		// Save user message
		_, err := h.store.CreateChatMessage(c.Request().Context(), db.CreateChatMessageParams{
			SessionID: session.ID,
			Role:      "user",
			Content:   req.InitialMessage,
		})
		if err != nil {
			c.Logger().Error("Failed to save initial message:", err)
		}

		// Send to agent
		if err := h.client.SendMessage(c.Request().Context(), agentSessionKey, req.InitialMessage); err != nil {
			c.Logger().Error("Failed to send initial message to agent:", err)
		}

		h.store.UpdateMessageCount(c.Request().Context(), session.ID)
	}

	return c.JSON(http.StatusCreated, ToChatSessionResponse(session))
}

// EndSession - DELETE /api/v1/agents/:id/sessions/:sessionId
func (h *ChatHandler) EndSession(c echo.Context) error {
	sessionID := c.Param("sessionId")

	// Verify session exists
	session, err := h.store.GetChatSession(c.Request().Context(), sessionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Session not found")
	}

	// Verify agent ID matches
	agentID := c.Param("id")
	if session.AgentID != agentID {
		return echo.NewHTTPError(http.StatusBadRequest, "Session does not belong to this agent")
	}

	// Update session status
	if err := h.store.EndChatSession(c.Request().Context(), sessionID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// ListSessions - GET /api/v1/agents/:id/sessions
func (h *ChatHandler) ListSessions(c echo.Context) error {
	agentID := c.Param("id")

	sessions, err := h.store.ListChatSessionsByAgent(c.Request().Context(), agentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ToChatSessionResponses(sessions))
}

// GetMessages - GET /api/v1/agents/:id/sessions/:sessionId/messages
// Returns messages from our DB, synced with OpenClaw history
func (h *ChatHandler) GetMessages(c echo.Context) error {
	sessionID := c.Param("sessionId")
	agentID := c.Param("id")

	// Verify session exists and belongs to agent
	session, err := h.store.GetChatSession(c.Request().Context(), sessionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Session not found")
	}

	if session.AgentID != agentID {
		return echo.NewHTTPError(http.StatusBadRequest, "Session does not belong to this agent")
	}

	// Get local messages
	localMessages, err := h.store.ListMessagesBySession(c.Request().Context(), sessionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Poll OpenClaw for new agent responses
	if session.OpenclawSessionKey.Valid && session.Status == "active" {
		h.syncAgentResponses(c, session, localMessages)
		
		// Re-fetch messages after sync
		localMessages, _ = h.store.ListMessagesBySession(c.Request().Context(), sessionID)
	}

	return c.JSON(http.StatusOK, ToChatMessageResponses(localMessages))
}

// syncAgentResponses fetches new responses from OpenClaw and saves them
func (h *ChatHandler) syncAgentResponses(c echo.Context, session db.ChatSession, existingMessages []db.ChatMessage) {
	history, err := h.client.GetSessionHistory(c.Request().Context(), session.OpenclawSessionKey.String, 50)
	if err != nil {
		c.Logger().Error("Failed to get session history:", err)
		return
	}

	// Find the last synced message timestamp
	var lastSyncTime time.Time
	for _, msg := range existingMessages {
		if msg.Role == "agent" && msg.CreatedAt.Time.After(lastSyncTime) {
			lastSyncTime = msg.CreatedAt.Time
		}
	}

	// Process new agent messages from history
	for _, histMsg := range history.Messages {
		if histMsg.Role != "assistant" {
			continue
		}

		// Extract text content from the message
		content := extractTextContent(histMsg.Content)
		if content == "" {
			continue
		}

		// Check if we already have this message (simple dedup by content)
		found := false
		for _, existing := range existingMessages {
			if existing.Role == "agent" && existing.Content == content {
				found = true
				break
			}
		}

		if !found {
			// Save new agent message
			_, err := h.store.CreateChatMessage(c.Request().Context(), db.CreateChatMessageParams{
				SessionID: session.ID,
				Role:      "agent",
				Content:   content,
			})
			if err != nil {
				c.Logger().Error("Failed to save agent message:", err)
			} else {
				h.store.UpdateMessageCount(c.Request().Context(), session.ID)
			}
		}
	}
}

// extractTextContent extracts text from message content
func extractTextContent(content string) string {
	// The content might be a string or might need parsing
	// For now, just trim and return if non-empty
	content = strings.TrimSpace(content)
	return content
}

// SendMessage - POST /api/v1/agents/:id/sessions/:sessionId/messages
func (h *ChatHandler) SendMessage(c echo.Context) error {
	sessionID := c.Param("sessionId")
	agentID := c.Param("id")

	var req SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if req.Content == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "content is required")
	}

	// Verify session exists and belongs to agent
	session, err := h.store.GetChatSession(c.Request().Context(), sessionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Session not found")
	}

	if session.AgentID != agentID {
		return echo.NewHTTPError(http.StatusBadRequest, "Session does not belong to this agent")
	}

	if session.Status != "active" {
		return echo.NewHTTPError(http.StatusBadRequest, "Session is not active")
	}

	// Save user message to database
	userMessage, err := h.store.CreateChatMessage(c.Request().Context(), db.CreateChatMessageParams{
		SessionID: sessionID,
		Role:      "user",
		Content:   req.Content,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Update message count
	h.store.UpdateMessageCount(c.Request().Context(), sessionID)

	// Send message to OpenClaw agent
	if session.OpenclawSessionKey.Valid {
		if err := h.client.SendMessage(c.Request().Context(), session.OpenclawSessionKey.String, req.Content); err != nil {
			c.Logger().Error("Failed to send message to agent:", err)
			// Don't fail the request - message is saved locally
		}
	}

	return c.JSON(http.StatusCreated, ToChatMessageResponse(userMessage))
}

// PollMessages - GET /api/v1/agents/:id/sessions/:sessionId/poll
// Long-polls for new messages (waits up to 30s for new messages)
func (h *ChatHandler) PollMessages(c echo.Context) error {
	sessionID := c.Param("sessionId")
	agentID := c.Param("id")

	// Verify session exists and belongs to agent
	session, err := h.store.GetChatSession(c.Request().Context(), sessionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Session not found")
	}

	if session.AgentID != agentID {
		return echo.NewHTTPError(http.StatusBadRequest, "Session does not belong to this agent")
	}

	// Get current message count
	initialMessages, _ := h.store.ListMessagesBySession(c.Request().Context(), sessionID)
	initialCount := len(initialMessages)

	// Poll for up to 30 seconds
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			// Return current messages on timeout
			messages, _ := h.store.ListMessagesBySession(c.Request().Context(), sessionID)
			return c.JSON(http.StatusOK, ToChatMessageResponses(messages))
		case <-ticker.C:
			// Sync with OpenClaw
			if session.OpenclawSessionKey.Valid && session.Status == "active" {
				existingMessages, _ := h.store.ListMessagesBySession(c.Request().Context(), sessionID)
				h.syncAgentResponses(c, session, existingMessages)
			}

			// Check for new messages
			messages, _ := h.store.ListMessagesBySession(c.Request().Context(), sessionID)
			if len(messages) > initialCount {
				return c.JSON(http.StatusOK, ToChatMessageResponses(messages))
			}
		case <-c.Request().Context().Done():
			return nil
		}
	}
}

// Helper conversion functions
func ToChatSessionResponse(s db.ChatSession) ChatSessionResponse {
	messageCount := 0
	if s.MessageCount.Valid {
		messageCount = int(s.MessageCount.Int64)
	}

	resp := ChatSessionResponse{
		ID:           s.ID,
		AgentID:      s.AgentID,
		Status:       s.Status,
		StartedAt:    s.StartedAt.Time.Format("2006-01-02T15:04:05Z"),
		MessageCount: messageCount,
	}

	if s.OpenclawSessionKey.Valid {
		resp.OpenclawSessionKey = &s.OpenclawSessionKey.String
	}

	if s.EndedAt.Valid {
		endedAt := s.EndedAt.Time.Format("2006-01-02T15:04:05Z")
		resp.EndedAt = &endedAt
	}

	return resp
}

func ToChatSessionResponses(sessions []db.ChatSession) []ChatSessionResponse {
	result := make([]ChatSessionResponse, len(sessions))
	for i, s := range sessions {
		result[i] = ToChatSessionResponse(s)
	}
	return result
}

func ToChatMessageResponse(m db.ChatMessage) ChatMessageResponse {
	return ChatMessageResponse{
		ID:        m.ID,
		SessionID: m.SessionID,
		Role:      m.Role,
		Content:   m.Content,
		CreatedAt: m.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

func ToChatMessageResponses(messages []db.ChatMessage) []ChatMessageResponse {
	result := make([]ChatMessageResponse, len(messages))
	for i, m := range messages {
		result[i] = ToChatMessageResponse(m)
	}
	return result
}
