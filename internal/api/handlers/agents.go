package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/openclaw"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
)

type AgentHandler struct {
	store        *store.Store
	agentCreator *openclaw.AgentCreator
}

func NewAgentHandler(s *store.Store) *AgentHandler {
	return &AgentHandler{
		store:        s,
		agentCreator: openclaw.NewAgentCreator(),
	}
}

// Request/Response types
type CreateAgentRequest struct {
	ID              string   `json:"id,omitempty"`
	Name            string   `json:"name" validate:"required"`
	Description     string   `json:"description"`
	Model           string   `json:"model"`
	MentionPatterns []string `json:"mention_patterns"`
	SoulMD          string   `json:"soul_md"`
	AgentsMD        string   `json:"agents_md"`
	IdentityMD      string   `json:"identity_md"`
	UserMD          string   `json:"user_md"`
	ToolsMD         string   `json:"tools_md"`
	HeartbeatMD     string   `json:"heartbeat_md"`
}

type UpdateAgentRequest struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Model           string   `json:"model"`
	MentionPatterns []string `json:"mention_patterns"`
	SoulMD          string   `json:"soul_md"`
	AgentsMD        string   `json:"agents_md"`
	IdentityMD      string   `json:"identity_md"`
	UserMD          string   `json:"user_md"`
	ToolsMD         string   `json:"tools_md"`
	HeartbeatMD     string   `json:"heartbeat_md"`
}

// Handlers
func (h *AgentHandler) List(c echo.Context) error {
	agents, err := h.store.ListAgents(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, ToAgentResponses(agents))
}

func (h *AgentHandler) Get(c echo.Context) error {
	id := c.Param("id")
	agent, err := h.store.GetAgent(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Agent not found")
	}
	return c.JSON(http.StatusOK, ToAgentResponse(agent))
}

func (h *AgentHandler) Create(c echo.Context) error {
	var req CreateAgentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Validate required fields
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	// Generate agent ID if not provided
	if req.ID == "" {
		req.ID = strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
	}

	// Create agent workspace and OpenClaw configuration
	// This will also generate identity files if description is provided
	createdAgent, err := h.agentCreator.CreateAgent(&openclaw.CreateAgentRequest{
		ID:              req.ID,
		Name:            req.Name,
		Description:     req.Description,
		Model:           req.Model,
		MentionPatterns: req.MentionPatterns,
		SoulMD:          req.SoulMD,
		AgentsMD:        req.AgentsMD,
		IdentityMD:      req.IdentityMD,
		UserMD:          req.UserMD,
		ToolsMD:         req.ToolsMD,
		HeartbeatMD:     req.HeartbeatMD,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create agent workspace: "+err.Error())
	}

	// Convert mention patterns to JSON string for database
	mentionJSON := "[]"
	if len(req.MentionPatterns) > 0 {
		jsonBytes, err := json.Marshal(req.MentionPatterns)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid mention_patterns format")
		}
		mentionJSON = string(jsonBytes)
	}

	// Save agent to database with the FINAL generated content (not request data)
	agent, err := h.store.CreateAgent(c.Request().Context(), db.CreateAgentParams{
		ID:              createdAgent.ID,
		Name:            req.Name,
		Description:     sql.NullString{String: req.Description, Valid: req.Description != ""},
		Status:          sql.NullString{String: "active", Valid: true},
		Model:           sql.NullString{String: req.Model, Valid: req.Model != ""},
		MentionPatterns: sql.NullString{String: mentionJSON, Valid: true},
		// Use the generated/final content from createdAgent, not req
		SoulMd:      sql.NullString{String: createdAgent.SoulMD, Valid: createdAgent.SoulMD != ""},
		AgentsMd:    sql.NullString{String: createdAgent.AgentsMD, Valid: createdAgent.AgentsMD != ""},
		IdentityMd:  sql.NullString{String: createdAgent.IdentityMD, Valid: createdAgent.IdentityMD != ""},
		UserMd:      sql.NullString{String: createdAgent.UserMD, Valid: createdAgent.UserMD != ""},
		ToolsMd:     sql.NullString{String: createdAgent.ToolsMD, Valid: createdAgent.ToolsMD != ""},
		HeartbeatMd: sql.NullString{String: createdAgent.HeartbeatMD, Valid: createdAgent.HeartbeatMD != ""},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, ToAgentResponse(agent))
}

func (h *AgentHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateAgentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Check if agent exists
	existing, err := h.store.GetAgent(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Agent not found")
	}

	// Use existing values if not provided in request
	name := req.Name
	if name == "" {
		name = existing.Name
	}

	// Convert mention patterns to JSON string
	mentionJSON := "[]"
	if len(req.MentionPatterns) > 0 {
		jsonBytes, err := json.Marshal(req.MentionPatterns)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid mention_patterns format")
		}
		mentionJSON = string(jsonBytes)
	} else if existing.MentionPatterns.Valid {
		mentionJSON = existing.MentionPatterns.String
	}

	agent, err := h.store.UpdateAgent(c.Request().Context(), db.UpdateAgentParams{
		ID:               id,
		Name:             name,
		Description:      sql.NullString{String: req.Description, Valid: req.Description != ""},
		Status:           existing.Status,
		Model:            sql.NullString{String: req.Model, Valid: req.Model != ""},
		MentionPatterns:  sql.NullString{String: mentionJSON, Valid: true},
		SoulMd:           sql.NullString{String: req.SoulMD, Valid: req.SoulMD != ""},
		AgentsMd:         sql.NullString{String: req.AgentsMD, Valid: req.AgentsMD != ""},
		IdentityMd:       sql.NullString{String: req.IdentityMD, Valid: req.IdentityMD != ""},
		UserMd:           sql.NullString{String: req.UserMD, Valid: req.UserMD != ""},
		ToolsMd:          sql.NullString{String: req.ToolsMD, Valid: req.ToolsMD != ""},
		HeartbeatMd:      sql.NullString{String: req.HeartbeatMD, Valid: req.HeartbeatMD != ""},
		ActiveSessionKey: existing.ActiveSessionKey,
		CurrentTaskID:    existing.CurrentTaskID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ToAgentResponse(agent))
}

func (h *AgentHandler) Delete(c echo.Context) error {
	id := c.Param("id")

	// Delete from database
	if err := h.store.DeleteAgent(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Delete agent workspace and OpenClaw configuration
	if err := h.agentCreator.DeleteAgent(id); err != nil {
		// Log error but don't fail the request since DB deletion succeeded
		c.Logger().Error("Failed to delete agent workspace:", err)
	}

	return c.NoContent(http.StatusNoContent)
}
