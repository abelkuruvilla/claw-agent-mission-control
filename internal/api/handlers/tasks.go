package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/openclaw"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
	ws "github.com/abelkuruvilla/claw-agent-mission-control/internal/websocket"
)

type TaskHandler struct {
	store        *store.Store
	hub          *ws.Hub
	orchestrator Orchestrator
	agentSender  *openclaw.AgentSender
}

type Orchestrator interface {
	StartTask(ctx context.Context, taskID string) error
	StopTask(taskID string) error
	PauseTask(taskID string) error
	GetRunningTasks() []string
	IsRunning(taskID string) bool
}

func NewTaskHandler(s *store.Store, hub *ws.Hub, agentSender *openclaw.AgentSender) *TaskHandler {
	return &TaskHandler{
		store:        s,
		hub:          hub,
		orchestrator: nil,
		agentSender:  agentSender,
	}
}

func (h *TaskHandler) SetOrchestrator(orch Orchestrator) {
	h.orchestrator = orch
}

// logEvent creates a persistent event record and broadcasts it via WebSocket.
func (h *TaskHandler) logEvent(ctx context.Context, taskID, agentID, eventType, message, details string) {
	event, err := h.store.CreateEvent(ctx, db.CreateEventParams{
		TaskID:  sql.NullString{String: taskID, Valid: taskID != ""},
		AgentID: sql.NullString{String: agentID, Valid: agentID != ""},
		Type:    eventType,
		Message: message,
		Details: sql.NullString{String: details, Valid: details != ""},
	})
	if err != nil {
		log.Printf("[TaskHandler] Failed to create event (%s): %v", eventType, err)
		return
	}
	if h.hub != nil {
		h.hub.BroadcastEvent(event)
	}
}

// NotifyAssignedAgent is the exported hook for the stuck-task watchdog to re-notify an agent.
func (h *TaskHandler) NotifyAssignedAgent(agentID, taskID, title, description string) {
	h.notifyAssignedAgent(agentID, taskID, title, description)
}

// NotifyParentTaskAgent is the exported hook for the watchdog to notify the parent's orchestrator (e.g. after reset).
func (h *TaskHandler) NotifyParentTaskAgent(ctx context.Context, subtask db.Task, newStatus string) {
	h.notifyParentTaskAgent(ctx, subtask, newStatus)
}

// notifyAssignedAgent fires an async notification to the agent about a task assignment.
// It saves the agent's reply (or error) as a comment on the task.
func (h *TaskHandler) notifyAssignedAgent(agentID, taskID, title, description string) {
	if h.agentSender == nil {
		log.Printf("[TaskHandler] Agent sender not configured, skipping notification for task %s", taskID)
		return
	}
	if agentID == "" || agentID == "unassigned" {
		return
	}

	log.Printf("[TaskHandler] Dispatching async notification to agent %s for task %s", agentID, taskID)

	h.agentSender.NotifyAgentAsync(agentID, taskID, title, description, func(tID, aID, reply string, err error) {
		ctx := context.Background()

		if err != nil {
			log.Printf("[TaskHandler] Agent %s failed to process task %s: %v", aID, tID, err)
			h.store.CreateComment(ctx, db.CreateCommentParams{
				TaskID:  tID,
				Author:  "system",
				Content: "[Agent Notification Error] Failed to notify agent " + aID + ": " + err.Error(),
			})
			return
		}

		if reply == "" {
			log.Printf("[TaskHandler] Agent %s returned empty reply for task %s", aID, tID)
			return
		}

		log.Printf("[TaskHandler] Saving agent %s reply as comment on task %s (len=%d)", aID, tID, len(reply))
		_, commentErr := h.store.CreateComment(ctx, db.CreateCommentParams{
			TaskID:  tID,
			Author:  aID,
			Content: reply,
		})
		if commentErr != nil {
			log.Printf("[TaskHandler] ERROR saving agent reply as comment: %v", commentErr)
		}
	})
}

// isAgentBusy returns true if the agent currently has active tasks
// (executing, planning, discussing, or verifying).
func (h *TaskHandler) isAgentBusy(ctx context.Context, agentID string) bool {
	if agentID == "" || agentID == "unassigned" {
		return false
	}
	count, err := h.store.CountActiveTasksByAgent(ctx, agentID)
	if err != nil {
		log.Printf("[TaskHandler] Error checking agent %s busy status: %v", agentID, err)
		return false
	}
	log.Printf("[TaskHandler] Agent %s has %d active tasks", agentID, count)
	return count > 0
}

// ProcessAgentQueue dequeues the next queued task for the given agent
// and notifies them. Called when an agent finishes a task or periodically.
func (h *TaskHandler) ProcessAgentQueue(ctx context.Context, agentID string) {
	if agentID == "" || agentID == "unassigned" {
		return
	}
	if h.isAgentBusy(ctx, agentID) {
		log.Printf("[QueueProcessor] Agent %s still busy, skipping queue processing", agentID)
		return
	}

	queued, err := h.store.ListQueuedTasksByAgent(ctx, agentID)
	if err != nil {
		log.Printf("[QueueProcessor] Error fetching queue for agent %s: %v", agentID, err)
		return
	}
	if len(queued) == 0 {
		log.Printf("[QueueProcessor] No queued tasks for agent %s", agentID)
		return
	}

	next := queued[0]
	log.Printf("[QueueProcessor] Dequeuing task %s (%s) for agent %s (queue depth: %d)", next.ID, next.Title, agentID, len(queued))

	if err := h.store.UpdateTaskStatus(ctx, next.ID, "backlog"); err != nil {
		log.Printf("[QueueProcessor] Error updating task %s status to backlog: %v", next.ID, err)
		return
	}

	h.logEvent(ctx, next.ID, agentID, "task_dequeued",
		fmt.Sprintf("Task dequeued for agent %s (was position 1 of %d)", agentID, len(queued)),
		fmt.Sprintf(`{"queue_depth":%d,"priority":%d}`, len(queued), next.Priority.Int64))

	if h.hub != nil {
		h.hub.BroadcastTaskStatus(next.ID, "backlog", 0)
	}

	desc := ""
	if next.Description.Valid {
		desc = next.Description.String
	}
	h.notifyAssignedAgent(agentID, next.ID, next.Title, desc)
}

// Request types
type CreateTaskRequest struct {
	Title          string `json:"title" validate:"required"`
	Description    string `json:"description"`
	AgentID        string `json:"agent_id"`
	ProjectID      string `json:"project_id"`
	ParentTaskID   string `json:"parent_task_id"`
	Status         string `json:"status"`
	Priority       int    `json:"priority"`
	QualityChecks  string `json:"quality_checks"`
	DelegationMode string `json:"delegation_mode"`
	ScheduledAt    string `json:"scheduled_at"`
	GitBranch      string `json:"git_branch"`
}

type UpdateTaskRequest struct {
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	AgentID        *string `json:"agent_id"`
	ProjectID      *string `json:"project_id"`
	Status         string  `json:"status"`
	Priority       int     `json:"priority"`
	ProjectMD      string  `json:"project_md"`
	RequirementsMD string  `json:"requirements_md"`
	RoadmapMD      string  `json:"roadmap_md"`
	StateMD        string  `json:"state_md"`
	PrdJSON        string  `json:"prd_json"`
	ProgressTxt    string  `json:"progress_txt"`
	GitBranch      string  `json:"git_branch"`
	QualityChecks  string  `json:"quality_checks"`
	DelegationMode string  `json:"delegation_mode"`
	ScheduledAt   string  `json:"scheduled_at"`
	ClearSchedule bool    `json:"clear_schedule"`
}

type CreatePhaseRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}

type CreateStoryRequest struct {
	Title              string   `json:"title" validate:"required"`
	Description        string   `json:"description"`
	Priority           int      `json:"priority"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
}

// fuzzyMatch returns true if query matches text via substring OR sequential character match.
// Both comparisons are case-insensitive.
func fuzzyMatch(query, text string) bool {
	if query == "" {
		return true
	}
	q := strings.ToLower(query)
	t := strings.ToLower(text)
	// Fast path: substring match
	if strings.Contains(t, q) {
		return true
	}
	// Sequential character match (fuzzy)
	qi := 0
	for ti := 0; ti < len(t) && qi < len(q); ti++ {
		if t[ti] == q[qi] {
			qi++
		}
	}
	return qi == len(q)
}

// Task CRUD
func (h *TaskHandler) List(c echo.Context) error {
	status := c.QueryParam("status")
	agentID := c.QueryParam("agent_id")

	var tasks []db.Task
	var err error

	if status != "" {
		tasks, err = h.store.ListTasksByStatus(c.Request().Context(), status)
	} else if agentID != "" {
		tasks, err = h.store.ListTasksByAgent(c.Request().Context(), agentID)
	} else {
		tasks, err = h.store.ListTasks(c.Request().Context())
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Search filter
	if search := c.QueryParam("search"); search != "" {
		filtered := tasks[:0]
		for _, t := range tasks {
			if fuzzyMatch(search, t.Title) {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	// Sort tasks
	sortBy := c.QueryParam("sort_by")
	sortOrder := c.QueryParam("sort_order")
	if sortBy == "" {
		sortBy = "created_at"
	}
	if sortOrder == "" {
		if sortBy == "name" || sortBy == "priority" {
			sortOrder = "asc"
		} else {
			sortOrder = "desc"
		}
	}

	// Helper to get time from NullTime
	getTime := func(t sql.NullTime) time.Time {
		if t.Valid {
			return t.Time
		}
		return time.Time{}
	}
	getPriority := func(p sql.NullInt64) int64 {
		if p.Valid {
			return p.Int64
		}
		return 0
	}

	sort.Slice(tasks, func(i, j int) bool {
		var less bool
		switch sortBy {
		case "updated_at":
			less = getTime(tasks[i].UpdatedAt).Before(getTime(tasks[j].UpdatedAt))
		case "name":
			less = strings.ToLower(tasks[i].Title) < strings.ToLower(tasks[j].Title)
		case "priority":
			less = getPriority(tasks[i].Priority) < getPriority(tasks[j].Priority)
		default: // created_at
			less = getTime(tasks[i].CreatedAt).Before(getTime(tasks[j].CreatedAt))
		}
		if sortOrder == "asc" {
			return less
		}
		return !less
	})

	return c.JSON(http.StatusOK, ToTaskResponses(tasks))
}

func (h *TaskHandler) Get(c echo.Context) error {
	id := c.Param("id")
	task, err := h.store.GetTask(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found")
	}

	// Get phases and stories
	phases, _ := h.store.ListPhasesByTask(c.Request().Context(), id)
	stories, _ := h.store.ListStoriesByTask(c.Request().Context(), id)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"task":    ToTaskResponse(task),
		"phases":  phases,
		"stories": stories,
	})
}

func (h *TaskHandler) Create(c echo.Context) error {
	var req CreateTaskRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	status := req.Status
	if status == "" {
		status = "backlog"
	}

	delegationMode := req.DelegationMode
	if delegationMode == "" {
		delegationMode = "auto"
	}

	var scheduledAt sql.NullTime
	isScheduled := false
	if req.ScheduledAt != "" {
		if t, err := time.Parse(time.RFC3339, req.ScheduledAt); err == nil && t.After(time.Now()) {
			scheduledAt = sql.NullTime{Time: t, Valid: true}
			isScheduled = true
		}
	}

	// If this is a subtask (has parent_task_id), inherit the parent's git_branch
	gitBranch := req.GitBranch
	if req.ParentTaskID != "" && gitBranch == "" {
		parentTask, err := h.store.GetTask(c.Request().Context(), req.ParentTaskID)
		if err == nil && parentTask.GitBranch.Valid {
			gitBranch = parentTask.GitBranch.String
		}
	}

	task, err := h.store.CreateTask(c.Request().Context(), db.CreateTaskParams{
		Title:          req.Title,
		Description:    sql.NullString{String: req.Description, Valid: req.Description != ""},
		AgentID:        sql.NullString{String: req.AgentID, Valid: req.AgentID != "" && req.AgentID != "unassigned"},
		ProjectID:      sql.NullString{String: req.ProjectID, Valid: req.ProjectID != ""},
		ParentTaskID:   sql.NullString{String: req.ParentTaskID, Valid: req.ParentTaskID != ""},
		Status:         sql.NullString{String: status, Valid: true},
		Priority:       sql.NullInt64{Int64: int64(req.Priority), Valid: true},
		QualityChecks:  sql.NullString{String: req.QualityChecks, Valid: req.QualityChecks != ""},
		DelegationMode: sql.NullString{String: delegationMode, Valid: true},
		ScheduledAt:    scheduledAt,
		GitBranch:      sql.NullString{String: gitBranch, Valid: gitBranch != ""},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	ctx := c.Request().Context()

	if req.ParentTaskID != "" {
		h.logEvent(ctx, req.ParentTaskID, req.AgentID, "subtask_created",
			fmt.Sprintf("Subtask created: %s", req.Title),
			fmt.Sprintf(`{"subtask_id":"%s","assigned_to":"%s"}`, task.ID, req.AgentID))
	} else {
		h.logEvent(ctx, task.ID, req.AgentID, "task_created",
			fmt.Sprintf("Task created: %s", req.Title), "")
	}

	if req.AgentID != "" && req.AgentID != "unassigned" && !isScheduled {
		if h.isAgentBusy(ctx, req.AgentID) {
			log.Printf("[TaskHandler] Agent %s is busy, queuing task %s", req.AgentID, task.ID)
			if err := h.store.UpdateTaskStatus(ctx, task.ID, "queued"); err != nil {
				log.Printf("[TaskHandler] Error setting task %s to queued: %v", task.ID, err)
			} else {
				task.Status = sql.NullString{String: "queued", Valid: true}
			}
			h.logEvent(ctx, task.ID, req.AgentID, "task_queued",
				fmt.Sprintf("Task queued for agent %s (agent is busy)", req.AgentID), "")
			if h.hub != nil {
				h.hub.BroadcastTaskStatus(task.ID, "queued", 0)
			}
		} else {
			h.logEvent(ctx, task.ID, req.AgentID, "agent_notified",
				fmt.Sprintf("Notifying agent %s of task assignment", req.AgentID), "")
			h.notifyAssignedAgent(req.AgentID, task.ID, req.Title, req.Description)
		}
	} else if isScheduled {
		log.Printf("[TaskHandler] Task %s scheduled for %s — skipping immediate dispatch", task.ID, req.ScheduledAt)
	}

	return c.JSON(http.StatusCreated, ToTaskResponse(task))
}

func (h *TaskHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateTaskRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get existing task first
	existing, err := h.store.GetTask(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found")
	}

	// Build update params, using existing values as defaults when new value is empty
	params := db.UpdateTaskParams{
		ID: id,
	}

	if req.Title != "" {
		params.Title = req.Title
	} else {
		params.Title = existing.Title
	}

	if req.Description != "" {
		params.Description = sql.NullString{String: req.Description, Valid: true}
	} else {
		params.Description = existing.Description
	}

	if req.AgentID != nil {
		agentVal := *req.AgentID
		params.AgentID = sql.NullString{String: agentVal, Valid: agentVal != "" && agentVal != "unassigned"}
	} else {
		params.AgentID = existing.AgentID
	}

	if req.ProjectID != nil {
		projectVal := *req.ProjectID
		params.ProjectID = sql.NullString{String: projectVal, Valid: projectVal != "" && projectVal != "none"}
	} else {
		params.ProjectID = existing.ProjectID
	}

	if req.Status != "" {
		params.Status = sql.NullString{String: req.Status, Valid: true}
	} else {
		params.Status = existing.Status
	}
	// When unassigning, move task to backlog so the card appears in the backlog column
	if req.AgentID != nil && (*req.AgentID == "" || *req.AgentID == "unassigned") {
		params.Status = sql.NullString{String: "backlog", Valid: true}
		log.Printf("[TaskHandler] Unassigning task %s: moving to backlog", id)
	}

	if req.Priority != 0 {
		params.Priority = sql.NullInt64{Int64: int64(req.Priority), Valid: true}
	} else {
		params.Priority = existing.Priority
	}

	if req.ProjectMD != "" {
		params.ProjectMd = sql.NullString{String: req.ProjectMD, Valid: true}
	} else {
		params.ProjectMd = existing.ProjectMd
	}

	if req.RequirementsMD != "" {
		params.RequirementsMd = sql.NullString{String: req.RequirementsMD, Valid: true}
	} else {
		params.RequirementsMd = existing.RequirementsMd
	}

	if req.RoadmapMD != "" {
		params.RoadmapMd = sql.NullString{String: req.RoadmapMD, Valid: true}
	} else {
		params.RoadmapMd = existing.RoadmapMd
	}

	if req.StateMD != "" {
		params.StateMd = sql.NullString{String: req.StateMD, Valid: true}
	} else {
		params.StateMd = existing.StateMd
	}

	if req.PrdJSON != "" {
		params.PrdJson = sql.NullString{String: req.PrdJSON, Valid: true}
	} else {
		params.PrdJson = existing.PrdJson
	}

	if req.ProgressTxt != "" {
		params.ProgressTxt = sql.NullString{String: req.ProgressTxt, Valid: true}
	} else {
		params.ProgressTxt = existing.ProgressTxt
	}

	if req.GitBranch != "" {
		params.GitBranch = sql.NullString{String: req.GitBranch, Valid: true}
	} else {
		params.GitBranch = existing.GitBranch
	}

	if req.QualityChecks != "" {
		params.QualityChecks = sql.NullString{String: req.QualityChecks, Valid: true}
	} else {
		params.QualityChecks = existing.QualityChecks
	}

	if req.DelegationMode != "" {
		params.DelegationMode = sql.NullString{String: req.DelegationMode, Valid: true}
	} else {
		params.DelegationMode = existing.DelegationMode
	}

	// Handle scheduled_at update/clear
	if req.ClearSchedule {
		// Clear schedule — execute immediately
		params.ScheduledAt = sql.NullTime{Valid: false}
	} else if req.ScheduledAt != "" {
		// Set/update schedule
		if t, err := time.Parse(time.RFC3339, req.ScheduledAt); err == nil {
			params.ScheduledAt = sql.NullTime{Time: t, Valid: true}
		} else {
			params.ScheduledAt = existing.ScheduledAt
		}
	} else {
		// No change
		params.ScheduledAt = existing.ScheduledAt
	}
	params.RetryAt = existing.RetryAt

	updated, err := h.store.UpdateTask(c.Request().Context(), params)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if h.hub != nil && updated.Status.Valid {
		h.hub.BroadcastTaskStatus(updated.ID, updated.Status.String, 0)
	}

	// If schedule was cleared and task is in backlog with an agent, notify immediately
	if req.ClearSchedule && updated.AgentID.Valid && updated.AgentID.String != "" {
		if updated.Status.Valid && updated.Status.String == "backlog" {
			desc := ""
			if updated.Description.Valid {
				desc = updated.Description.String
			}
			h.logEvent(c.Request().Context(), updated.ID, updated.AgentID.String,
				"agent_notified", "Task unscheduled — notifying agent immediately", "")
			h.notifyAssignedAgent(updated.AgentID.String, updated.ID, updated.Title, desc)
		}
	}

	// Notify agent if the assignment changed to a new (different) agent
	oldAgentID := ""
	if existing.AgentID.Valid {
		oldAgentID = existing.AgentID.String
	}
	newAgentID := ""
	if req.AgentID != nil && *req.AgentID != "" && *req.AgentID != "unassigned" {
		newAgentID = *req.AgentID
	}
	if newAgentID != "" && newAgentID != oldAgentID {
		h.logEvent(c.Request().Context(), updated.ID, newAgentID, "task_assigned",
			fmt.Sprintf("Task reassigned to agent %s", newAgentID), "")
		desc := ""
		if updated.Description.Valid {
			desc = updated.Description.String
		}

		if h.isAgentBusy(c.Request().Context(), newAgentID) {
			log.Printf("[TaskHandler] Agent %s is busy, queuing reassigned task %s", newAgentID, updated.ID)
			if err := h.store.UpdateTaskStatus(c.Request().Context(), updated.ID, "queued"); err != nil {
				log.Printf("[TaskHandler] Error setting task %s to queued: %v", updated.ID, err)
			} else {
				updated.Status = sql.NullString{String: "queued", Valid: true}
			}
			h.logEvent(c.Request().Context(), updated.ID, newAgentID, "task_queued",
				fmt.Sprintf("Task queued for agent %s (agent is busy)", newAgentID), "")
			if h.hub != nil {
				h.hub.BroadcastTaskStatus(updated.ID, "queued", 0)
			}
		} else {
			h.logEvent(c.Request().Context(), updated.ID, newAgentID, "agent_notified",
				fmt.Sprintf("Notifying agent %s of task assignment", newAgentID), "")
			h.notifyAssignedAgent(newAgentID, updated.ID, updated.Title, desc)
		}
	}

	return c.JSON(http.StatusOK, ToTaskResponse(updated))
}

func (h *TaskHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.store.DeleteTask(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *TaskHandler) UpdateStatus(c echo.Context) error {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.store.UpdateTaskStatus(c.Request().Context(), id, req.Status); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// Clear watchdog retry count on any status transition so normal progress is not treated as stuck
	if err := h.store.ResetTaskRetryCount(c.Request().Context(), id); err != nil {
		log.Printf("[TaskHandler] Failed to reset retry count for task %s: %v", id, err)
	}

	task, err := h.store.GetTask(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	ctx := c.Request().Context()
	agentID := ""
	if task.AgentID.Valid {
		agentID = task.AgentID.String
	}

	h.logEvent(ctx, id, agentID, "status_changed",
		fmt.Sprintf("Status changed to %s", req.Status), "")

	if h.hub != nil {
		h.hub.BroadcastTaskStatus(id, req.Status, 0)
	}

	if req.Status == "done" || req.Status == "failed" || req.Status == "cancelled" {
		h.notifyParentTaskAgent(ctx, task, req.Status)

		if agentID != "" {
			go h.ProcessAgentQueue(context.Background(), agentID)
		}
	}

	return c.JSON(http.StatusOK, ToTaskResponse(task))
}

// RetryTask resets retry_count, sets status to backlog, and re-notifies the assigned agent.
// Used when a task is stuck (e.g. after rate limiting) to give it another chance.
func (h *TaskHandler) RetryTask(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()

	task, err := h.store.GetTask(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Task not found")
	}

	// Check for scheduled retry
	var retryReq struct {
		RetryAt string `json:"retry_at"`
	}
	if err := json.NewDecoder(c.Request().Body).Decode(&retryReq); err == nil && retryReq.RetryAt != "" {
		if t, err := time.Parse(time.RFC3339, retryReq.RetryAt); err == nil && t.After(time.Now()) {
			if err := h.store.ResetTaskRetryCount(ctx, id); err != nil {
				log.Printf("[TaskHandler] Failed to reset retry count for task %s: %v", id, err)
			}
			if err := h.store.SetTaskRetryAt(ctx, id, t); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
			task, _ = h.store.GetTask(ctx, id)
			agentID := ""
			if task.AgentID.Valid {
				agentID = task.AgentID.String
			}
			h.logEvent(ctx, id, agentID, "task_retry_scheduled",
				fmt.Sprintf("Task retry scheduled for %s", retryReq.RetryAt), "")
			if h.hub != nil {
				h.hub.BroadcastTaskStatus(id, "backlog", 0)
			}
			return c.JSON(http.StatusOK, ToTaskResponse(task))
		}
	}

	// Immediate retry (existing behavior)
	if err := h.store.ResetTaskRetryCount(ctx, id); err != nil {
		log.Printf("[TaskHandler] Failed to reset retry count for task %s: %v", id, err)
	}
	if err := h.store.UpdateTaskStatus(ctx, id, "backlog"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	task, _ = h.store.GetTask(ctx, id)
	agentID := ""
	if task.AgentID.Valid {
		agentID = task.AgentID.String
	}

	h.logEvent(ctx, id, agentID, "task_retry",
		fmt.Sprintf("Task \"%s\" manually retried (status set to backlog)", task.Title), "")

	if h.hub != nil {
		h.hub.BroadcastTaskStatus(id, "backlog", 0)
	}

	if agentID != "" && agentID != "unassigned" {
		desc := ""
		if task.Description.Valid {
			desc = task.Description.String
		}
		h.notifyAssignedAgent(agentID, id, task.Title, desc)
	}

	return c.JSON(http.StatusOK, ToTaskResponse(task))
}

// notifyParentTaskAgent checks if a completed/failed task is a subtask,
// and if so, sends a push notification to the parent task's assigned agent
// (the orchestrator) so it can continue the delegation chain.
// If the parent task's delegation_mode is "manual", the notification is held
// for human approval instead of being sent immediately.
func (h *TaskHandler) notifyParentTaskAgent(ctx context.Context, subtask db.Task, newStatus string) {
	if h.agentSender == nil {
		return
	}
	if !subtask.ParentTaskID.Valid || subtask.ParentTaskID.String == "" {
		return
	}

	parentTaskID := subtask.ParentTaskID.String
	parentTask, err := h.store.GetTask(ctx, parentTaskID)
	if err != nil {
		log.Printf("[TaskHandler] Could not fetch parent task %s for subtask completion notification: %v", parentTaskID, err)
		return
	}

	if !parentTask.AgentID.Valid || parentTask.AgentID.String == "" {
		log.Printf("[TaskHandler] Parent task %s has no assigned agent, skipping subtask completion notification", parentTaskID)
		return
	}

	orchestratorID := parentTask.AgentID.String
	subtaskAgentID := ""
	if subtask.AgentID.Valid {
		subtaskAgentID = subtask.AgentID.String
	}

	delegationMode := "auto"
	if parentTask.DelegationMode.Valid && parentTask.DelegationMode.String != "" {
		delegationMode = parentTask.DelegationMode.String
	}

	h.logEvent(ctx, parentTaskID, orchestratorID, "subtask_result_received",
		fmt.Sprintf("Subtask \"%s\" completed with status: %s", subtask.Title, newStatus),
		fmt.Sprintf(`{"subtask_id":"%s","status":"%s","specialist":"%s"}`, subtask.ID, newStatus, subtaskAgentID))

	if delegationMode == "manual" {
		log.Printf("[TaskHandler] Parent task %s has manual delegation — holding subtask %s for human approval", parentTaskID, subtask.ID)
		h.logEvent(ctx, parentTaskID, orchestratorID, "pending_approval",
			fmt.Sprintf("Subtask \"%s\" awaiting human approval before notifying orchestrator", subtask.Title),
			fmt.Sprintf(`{"subtask_id":"%s","status":"%s"}`, subtask.ID, newStatus))
		return
	}

	log.Printf("[TaskHandler] Subtask %s (%s) reached status %s — notifying orchestrator %s on parent task %s",
		subtask.ID, subtask.Title, newStatus, orchestratorID, parentTaskID)

	h.logEvent(ctx, parentTaskID, orchestratorID, "orchestrator_notified",
		fmt.Sprintf("Notifying orchestrator %s: subtask \"%s\" is %s", orchestratorID, subtask.Title, newStatus), "")

	h.agentSender.NotifySubtaskCompletionAsync(
		orchestratorID,
		subtask.ID, subtask.Title, newStatus,
		parentTaskID, parentTask.Title,
		subtaskAgentID,
		func(tID, aID, reply string, err error) {
			bgCtx := context.Background()
			if err != nil {
				log.Printf("[TaskHandler] Failed to notify orchestrator %s about subtask %s: %v", aID, subtask.ID, err)
				h.logEvent(bgCtx, tID, aID, "notification_error",
					fmt.Sprintf("Failed to notify orchestrator %s about subtask completion: %s", aID, err.Error()), "")
				h.store.CreateComment(bgCtx, db.CreateCommentParams{
					TaskID:  tID,
					Author:  "system",
					Content: "[Subtask Notification Error] Failed to notify orchestrator " + aID + " about subtask " + subtask.ID + " completion: " + err.Error(),
				})
				return
			}
			h.logEvent(bgCtx, tID, aID, "orchestrator_acknowledged",
				fmt.Sprintf("Orchestrator %s acknowledged subtask \"%s\" completion", aID, subtask.Title), "")
			if reply != "" {
				log.Printf("[TaskHandler] Orchestrator %s replied to subtask %s completion (len=%d)", aID, subtask.ID, len(reply))
				h.store.CreateComment(bgCtx, db.CreateCommentParams{
					TaskID:  tID,
					Author:  aID,
					Content: reply,
				})
			}
		},
	)
}

// Phase handlers
func (h *TaskHandler) ListPhases(c echo.Context) error {
	taskID := c.Param("id")
	phases, err := h.store.ListPhasesByTask(c.Request().Context(), taskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, phases)
}

func (h *TaskHandler) CreatePhase(c echo.Context) error {
	taskID := c.Param("id")
	var req CreatePhaseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get next sequence number
	phases, _ := h.store.ListPhasesByTask(c.Request().Context(), taskID)
	seq := int64(len(phases) + 1)

	phase, err := h.store.CreatePhase(c.Request().Context(), db.CreatePhaseParams{
		TaskID:      taskID,
		Sequence:    seq,
		Title:       req.Title,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Status:      sql.NullString{String: "pending", Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, phase)
}

// Story handlers
func (h *TaskHandler) ListStories(c echo.Context) error {
	taskID := c.Param("id")
	stories, err := h.store.ListStoriesByTask(c.Request().Context(), taskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, stories)
}

func (h *TaskHandler) CreateStory(c echo.Context) error {
	taskID := c.Param("id")
	var req CreateStoryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	stories, _ := h.store.ListStoriesByTask(c.Request().Context(), taskID)
	seq := int64(len(stories) + 1)

	// Convert acceptance criteria to JSON
	acJSON := "[]"
	if len(req.AcceptanceCriteria) > 0 {
		acBytes, err := json.Marshal(req.AcceptanceCriteria)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid acceptance criteria format")
		}
		acJSON = string(acBytes)
	}

	story, err := h.store.CreateStory(c.Request().Context(), db.CreateStoryParams{
		TaskID:             taskID,
		Sequence:           seq,
		Title:              req.Title,
		Description:        sql.NullString{String: req.Description, Valid: req.Description != ""},
		Priority:           sql.NullInt64{Int64: int64(req.Priority), Valid: true},
		AcceptanceCriteria: sql.NullString{String: acJSON, Valid: true},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, story)
}

// Subtask endpoints
func (h *TaskHandler) ListSubtasks(c echo.Context) error {
	parentID := c.Param("id")
	log.Printf("[TaskHandler] Listing subtasks for parent task %s", parentID)

	subtasks, err := h.store.ListSubtasks(c.Request().Context(), sql.NullString{String: parentID, Valid: true})
	if err != nil {
		log.Printf("[TaskHandler] ERROR listing subtasks for %s: %v", parentID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	log.Printf("[TaskHandler] Found %d subtasks for parent task %s", len(subtasks), parentID)
	return c.JSON(http.StatusOK, ToTaskResponses(subtasks))
}

// Execution endpoints
func (h *TaskHandler) StartTask(c echo.Context) error {
	id := c.Param("id")
	if h.orchestrator == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Orchestrator not available")
	}
	if err := h.orchestrator.StartTask(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "started"})
}

func (h *TaskHandler) StopTask(c echo.Context) error {
	id := c.Param("id")
	if h.orchestrator == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "Orchestrator not available")
	}
	if err := h.orchestrator.StopTask(id); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "stopped"})
}

// ApproveDelegation approves a completed subtask and triggers the orchestrator notification.
// Used when the parent task has delegation_mode = "manual".
func (h *TaskHandler) ApproveDelegation(c echo.Context) error {
	subtaskID := c.Param("id")
	ctx := c.Request().Context()

	subtask, err := h.store.GetTask(ctx, subtaskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Subtask not found")
	}

	if !subtask.ParentTaskID.Valid || subtask.ParentTaskID.String == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Task is not a subtask")
	}

	status := "backlog"
	if subtask.Status.Valid {
		status = subtask.Status.String
	}
	if status != "done" && status != "failed" {
		return echo.NewHTTPError(http.StatusBadRequest, "Subtask must be done or failed to approve delegation")
	}

	parentTaskID := subtask.ParentTaskID.String
	parentTask, err := h.store.GetTask(ctx, parentTaskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not fetch parent task")
	}

	if !parentTask.AgentID.Valid || parentTask.AgentID.String == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Parent task has no assigned orchestrator agent")
	}

	orchestratorID := parentTask.AgentID.String
	subtaskAgentID := ""
	if subtask.AgentID.Valid {
		subtaskAgentID = subtask.AgentID.String
	}

	log.Printf("[TaskHandler] Human approved delegation for subtask %s — notifying orchestrator %s", subtaskID, orchestratorID)

	h.logEvent(ctx, parentTaskID, "", "delegation_approved",
		fmt.Sprintf("Human approved subtask \"%s\" — notifying orchestrator", subtask.Title),
		fmt.Sprintf(`{"subtask_id":"%s","status":"%s"}`, subtaskID, status))

	h.agentSender.NotifySubtaskCompletionAsync(
		orchestratorID,
		subtask.ID, subtask.Title, status,
		parentTaskID, parentTask.Title,
		subtaskAgentID,
		func(tID, aID, reply string, sendErr error) {
			bgCtx := context.Background()
			if sendErr != nil {
				log.Printf("[TaskHandler] Failed to notify orchestrator %s after approval: %v", aID, sendErr)
				h.logEvent(bgCtx, tID, aID, "notification_error",
					fmt.Sprintf("Failed to notify orchestrator after approval: %s", sendErr.Error()), "")
				return
			}
			h.logEvent(bgCtx, tID, aID, "orchestrator_acknowledged",
				fmt.Sprintf("Orchestrator %s acknowledged approved subtask \"%s\"", aID, subtask.Title), "")
			if reply != "" {
				h.store.CreateComment(bgCtx, db.CreateCommentParams{
					TaskID:  tID,
					Author:  aID,
					Content: reply,
				})
			}
		},
	)

	return c.JSON(http.StatusOK, map[string]string{"status": "approved", "subtask_id": subtaskID})
}

// GetAgentQueue returns all queued tasks for a specific agent, ordered by priority then FIFO.
// Agents call this on heartbeat to check for pending work.
func (h *TaskHandler) GetAgentQueue(c echo.Context) error {
	agentID := c.Param("id")
	ctx := c.Request().Context()

	log.Printf("[TaskHandler] Agent %s checking queue", agentID)

	queued, err := h.store.ListQueuedTasksByAgent(ctx, agentID)
	if err != nil {
		log.Printf("[TaskHandler] Error fetching queue for agent %s: %v", agentID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	log.Printf("[TaskHandler] Agent %s has %d queued tasks", agentID, len(queued))
	return c.JSON(http.StatusOK, map[string]interface{}{
		"agent_id":    agentID,
		"queue_depth": len(queued),
		"tasks":       ToTaskResponses(queued),
	})
}

// DequeueNextTask picks the next task from an agent's queue, transitions it
// from "queued" to "backlog", notifies the agent, and returns the task.
// Agents call this to self-serve pickup during heartbeat.
func (h *TaskHandler) DequeueNextTask(c echo.Context) error {
	agentID := c.Param("id")
	ctx := c.Request().Context()

	log.Printf("[TaskHandler] Agent %s requesting next queued task", agentID)

	if h.isAgentBusy(ctx, agentID) {
		log.Printf("[TaskHandler] Agent %s is still busy, cannot dequeue", agentID)
		return c.JSON(http.StatusConflict, map[string]string{
			"error": "Agent is currently busy with active tasks",
		})
	}

	queued, err := h.store.ListQueuedTasksByAgent(ctx, agentID)
	if err != nil {
		log.Printf("[TaskHandler] Error fetching queue for agent %s: %v", agentID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if len(queued) == 0 {
		log.Printf("[TaskHandler] No queued tasks for agent %s", agentID)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"agent_id": agentID,
			"task":     nil,
			"message":  "No queued tasks",
		})
	}

	next := queued[0]
	log.Printf("[TaskHandler] Dequeuing task %s (%s) for agent %s", next.ID, next.Title, agentID)

	if err := h.store.UpdateTaskStatus(ctx, next.ID, "backlog"); err != nil {
		log.Printf("[TaskHandler] Error updating task %s status: %v", next.ID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	h.logEvent(ctx, next.ID, agentID, "task_dequeued",
		fmt.Sprintf("Task dequeued by agent %s via heartbeat pickup (was position 1 of %d)", agentID, len(queued)),
		fmt.Sprintf(`{"queue_depth":%d,"priority":%d,"trigger":"heartbeat"}`, len(queued), next.Priority.Int64))

	if h.hub != nil {
		h.hub.BroadcastTaskStatus(next.ID, "backlog", 0)
	}

	desc := ""
	if next.Description.Valid {
		desc = next.Description.String
	}
	h.notifyAssignedAgent(agentID, next.ID, next.Title, desc)

	updatedTask, err := h.store.GetTask(ctx, next.ID)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"agent_id":        agentID,
			"task":            ToTaskResponse(next),
			"remaining_queue": len(queued) - 1,
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"agent_id":        agentID,
		"task":            ToTaskResponse(updatedTask),
		"remaining_queue": len(queued) - 1,
	})
}

// RequestChanges sends a change request to the specialist agent working on a subtask.
// Resets the subtask status to "executing" and notifies the agent with the comment.
func (h *TaskHandler) RequestChanges(c echo.Context) error {
	subtaskID := c.Param("id")
	ctx := c.Request().Context()

	var req struct {
		Comment string `json:"comment" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Comment == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Comment is required")
	}

	subtask, err := h.store.GetTask(ctx, subtaskID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Subtask not found")
	}

	if !subtask.AgentID.Valid || subtask.AgentID.String == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Subtask has no assigned agent")
	}

	agentID := subtask.AgentID.String
	parentTaskID := ""
	if subtask.ParentTaskID.Valid {
		parentTaskID = subtask.ParentTaskID.String
	}

	h.store.CreateComment(ctx, db.CreateCommentParams{
		TaskID:  subtaskID,
		Author:  "human",
		Content: req.Comment,
	})

	if err := h.store.UpdateTaskStatus(ctx, subtaskID, "executing"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	h.logEvent(ctx, subtaskID, agentID, "changes_requested",
		fmt.Sprintf("Human requested changes: %s", req.Comment), "")

	if parentTaskID != "" {
		h.logEvent(ctx, parentTaskID, "", "changes_requested",
			fmt.Sprintf("Changes requested on subtask \"%s\"", subtask.Title),
			fmt.Sprintf(`{"subtask_id":"%s"}`, subtaskID))
	}

	if h.agentSender != nil {
		changeMsg := fmt.Sprintf(
			"Changes have been requested on your task.\n\n"+
				"## Change Request\n"+
				"- **Task ID:** %s\n"+
				"- **Title:** %s\n"+
				"- **Feedback:** %s\n\n"+
				"Please review the feedback, make the requested changes, and update the task status to `done` when complete.",
			subtaskID, subtask.Title, req.Comment,
		)

		h.agentSender.NotifyAgentAsync(agentID, subtaskID, subtask.Title, changeMsg,
			func(tID, aID, reply string, sendErr error) {
				bgCtx := context.Background()
				if sendErr != nil {
					log.Printf("[TaskHandler] Failed to notify agent %s about change request: %v", aID, sendErr)
					return
				}
				if reply != "" {
					h.store.CreateComment(bgCtx, db.CreateCommentParams{
						TaskID:  tID,
						Author:  aID,
						Content: reply,
					})
				}
			},
		)
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "changes_requested", "subtask_id": subtaskID})
}
