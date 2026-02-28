package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
	ws "github.com/abelkuruvilla/claw-agent-mission-control/internal/websocket"
)

type ReportingHandler struct {
	store *store.Store
	hub   *ws.Hub
}

func NewReportingHandler(s *store.Store, hub *ws.Hub) *ReportingHandler {
	return &ReportingHandler{store: s, hub: hub}
}

// Phase reporting
type PhaseProgressRequest struct {
	Progress float64 `json:"progress"`
	Message  string  `json:"message"`
}

type PhaseCompleteRequest struct {
	Summary   string                 `json:"summary"`
	Artifacts map[string]interface{} `json:"artifacts"`
}

type PhaseFailRequest struct {
	Error       string `json:"error"`
	Recoverable bool   `json:"recoverable"`
	Suggestion  string `json:"suggestion,omitempty"`
}

func (h *ReportingHandler) UpdatePhaseProgress(c echo.Context) error {
	phaseID := c.Param("id")
	var req PhaseProgressRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get phase to find task
	phase, err := h.store.GetPhase(c.Request().Context(), phaseID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Phase not found")
	}

	// Create event for progress
	h.store.CreateEvent(c.Request().Context(), db.CreateEventParams{
		TaskID:  sql.NullString{String: phase.TaskID, Valid: true},
		Type:    "phase_progress",
		Message: req.Message,
	})

	// Broadcast via WebSocket
	if h.hub != nil {
		h.hub.BroadcastTaskStatus(phase.TaskID, "executing", req.Progress)
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "progress_updated"})
}

func (h *ReportingHandler) CompletePhase(c echo.Context) error {
	phaseID := c.Param("id")
	var req PhaseCompleteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Update phase status
	if err := h.store.UpdatePhaseStatus(c.Request().Context(), phaseID, "done"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	phase, _ := h.store.GetPhase(c.Request().Context(), phaseID)

	// Create completion event
	h.store.CreateEvent(c.Request().Context(), db.CreateEventParams{
		TaskID:  sql.NullString{String: phase.TaskID, Valid: true},
		Type:    "phase_completed",
		Message: req.Summary,
	})

	// Broadcast
	if h.hub != nil {
		h.hub.Broadcast(&ws.Message{
			Type:    ws.EventPhaseUpdated,
			Payload: phase,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "completed"})
}

func (h *ReportingHandler) FailPhase(c echo.Context) error {
	phaseID := c.Param("id")
	var req PhaseFailRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	status := "failed"
	if req.Recoverable {
		status = "error"
	}

	if err := h.store.UpdatePhaseStatus(c.Request().Context(), phaseID, status); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	phase, _ := h.store.GetPhase(c.Request().Context(), phaseID)

	h.store.CreateEvent(c.Request().Context(), db.CreateEventParams{
		TaskID:  sql.NullString{String: phase.TaskID, Valid: true},
		Type:    "phase_failed",
		Message: req.Error,
	})

	return c.JSON(http.StatusOK, map[string]string{"status": status})
}

// Story reporting (Ralph)
type StoryPassRequest struct {
	CommitSHA string `json:"commit_sha"`
	Learnings string `json:"learnings"`
}

type StoryFailRequest struct {
	Error     string `json:"error"`
	Iteration int    `json:"iteration"`
}

type ProgressTxtRequest struct {
	Content string `json:"content"`
}

func (h *ReportingHandler) PassStory(c echo.Context) error {
	storyID := c.Param("id")
	var req StoryPassRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.store.MarkStoryPassed(c.Request().Context(), storyID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	story, _ := h.store.GetStory(c.Request().Context(), storyID)

	h.store.CreateEvent(c.Request().Context(), db.CreateEventParams{
		TaskID:  sql.NullString{String: story.TaskID, Valid: true},
		Type:    "story_passed",
		Message: "Story passed: " + story.Title,
	})

	if h.hub != nil {
		h.hub.Broadcast(&ws.Message{
			Type:    ws.EventStoryUpdated,
			Payload: story,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "passed"})
}

func (h *ReportingHandler) FailStory(c echo.Context) error {
	storyID := c.Param("id")
	var req StoryFailRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.store.MarkStoryFailed(c.Request().Context(), storyID, req.Error); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	story, _ := h.store.GetStory(c.Request().Context(), storyID)

	h.store.CreateEvent(c.Request().Context(), db.CreateEventParams{
		TaskID:  sql.NullString{String: story.TaskID, Valid: true},
		Type:    "story_failed",
		Message: req.Error,
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "failed",
		"iteration": strconv.Itoa(req.Iteration),
	})
}

func (h *ReportingHandler) AppendProgressTxt(c echo.Context) error {
	taskID := c.Param("id")
	var req ProgressTxtRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.store.AppendProgressTxt(c.Request().Context(), taskID, req.Content); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "appended"})
}
