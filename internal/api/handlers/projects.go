package handlers

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/google/uuid"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
)

type ProjectHandler struct {
	store *store.Store
}

func NewProjectHandler(s *store.Store) *ProjectHandler {
	return &ProjectHandler{
		store: s,
	}
}

// Request types
type CreateProjectRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Status      string `json:"status"`   // "active" | "completed" | "on-hold"
	Color       string `json:"color"`    // hex color string
	Location    string `json:"location"` // project directory path
	DefaultBranch     string `json:"default_branch"`
	LocalExecBranch  string `json:"local_exec_branch"`
	RemoteMergeBranch string `json:"remote_merge_branch"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Color       string `json:"color"`
	Location    string `json:"location"`
	DefaultBranch     string `json:"default_branch"`
	LocalExecBranch  string `json:"local_exec_branch"`
	RemoteMergeBranch string `json:"remote_merge_branch"`
}

// Response types
type ProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Color       string `json:"color"`
	Location    string `json:"location"`
	DefaultBranch     string `json:"default_branch"`
	LocalExecBranch  string `json:"local_exec_branch"`
	RemoteMergeBranch string `json:"remote_merge_branch"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	TaskCount   int64  `json:"task_count,omitempty"`
	DoneCount   int64  `json:"done_count,omitempty"`
}

// List all projects
func (h *ProjectHandler) List(c echo.Context) error {
	status := c.QueryParam("status")

	var projects []db.Project
	var err error

	if status != "" {
		projects, err = h.store.ListProjectsByStatus(c.Request().Context(), sql.NullString{String: status, Valid: true})
	} else {
		projects, err = h.store.ListProjects(c.Request().Context())
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	responses := make([]ProjectResponse, len(projects))
	for i, p := range projects {
		responses[i] = toProjectResponse(p)
	}

	return c.JSON(http.StatusOK, responses)
}

// Get a single project with stats
func (h *ProjectHandler) Get(c echo.Context) error {
	id := c.Param("id")
	project, err := h.store.GetProject(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Project not found")
	}

	// Get task counts
	taskCount, _ := h.store.GetProjectTaskCount(c.Request().Context(), sql.NullString{String: id, Valid: true})
	doneCount, _ := h.store.GetProjectDoneTaskCount(c.Request().Context(), sql.NullString{String: id, Valid: true})

	response := toProjectResponse(project)
	response.TaskCount = taskCount
	response.DoneCount = doneCount

	return c.JSON(http.StatusOK, response)
}

// Create a new project
func (h *ProjectHandler) Create(c echo.Context) error {
	var req CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Generate UUID for new project
	id := uuid.New().String()

	// Set defaults
	status := req.Status
	if status == "" {
		status = "active"
	}

	color := req.Color
	if color == "" {
		color = "#8b5cf6" // default purple
	}

	project, err := h.store.CreateProject(c.Request().Context(), db.CreateProjectParams{
		ID:          id,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Status:      sql.NullString{String: status, Valid: true},
		Color:       sql.NullString{String: color, Valid: true},
		Location:    sql.NullString{String: req.Location, Valid: req.Location != ""},
		DefaultBranch:     sql.NullString{String: req.DefaultBranch, Valid: req.DefaultBranch != ""},
		LocalExecBranch:  sql.NullString{String: req.LocalExecBranch, Valid: req.LocalExecBranch != ""},
		RemoteMergeBranch: sql.NullString{String: req.RemoteMergeBranch, Valid: req.RemoteMergeBranch != ""},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, toProjectResponse(project))
}

// Update an existing project
func (h *ProjectHandler) Update(c echo.Context) error {
	id := c.Param("id")
	var req UpdateProjectRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get existing project first
	existing, err := h.store.GetProject(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Project not found")
	}

	// Build update params, using existing values as defaults when new value is empty
	name := req.Name
	if name == "" {
		name = existing.Name
	}

	description := req.Description
	if description == "" && existing.Description.Valid {
		description = existing.Description.String
	}

	status := req.Status
	if status == "" && existing.Status.Valid {
		status = existing.Status.String
	}

	color := req.Color
	if color == "" && existing.Color.Valid {
		color = existing.Color.String
	}

	location := req.Location
	if location == "" && existing.Location.Valid {
		location = existing.Location.String
	}

	updated, err := h.store.UpdateProject(c.Request().Context(), db.UpdateProjectParams{
		ID:          id,
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
		Status:      sql.NullString{String: status, Valid: status != ""},
		Color:       sql.NullString{String: color, Valid: color != ""},
		Location:    sql.NullString{String: location, Valid: location != ""},
		DefaultBranch:     sql.NullString{String: req.DefaultBranch, Valid: req.DefaultBranch != ""},
		LocalExecBranch:  sql.NullString{String: req.LocalExecBranch, Valid: req.LocalExecBranch != ""},
		RemoteMergeBranch: sql.NullString{String: req.RemoteMergeBranch, Valid: req.RemoteMergeBranch != ""},
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, toProjectResponse(updated))
}

// Delete a project
func (h *ProjectHandler) Delete(c echo.Context) error {
	id := c.Param("id")
	if err := h.store.DeleteProject(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// Get all tasks for a project
func (h *ProjectHandler) ListTasks(c echo.Context) error {
	id := c.Param("id")
	
	// Verify project exists
	_, err := h.store.GetProject(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Project not found")
	}

	tasks, err := h.store.ListTasksByProject(c.Request().Context(), sql.NullString{String: id, Valid: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ToTaskResponses(tasks))
}

// Helper functions
func toProjectResponse(p db.Project) ProjectResponse {
	return ProjectResponse{
		ID:                p.ID,
		Name:              p.Name,
		Description:       nullStringToString(p.Description),
		Status:            nullStringToString(p.Status),
		Color:             nullStringToString(p.Color),
		Location:          nullStringToString(p.Location),
		DefaultBranch:     nullStringToString(p.DefaultBranch),
		LocalExecBranch:   nullStringToString(p.LocalExecBranch),
		RemoteMergeBranch: nullStringToString(p.RemoteMergeBranch),
		CreatedAt:         nullTimeToString(p.CreatedAt),
		UpdatedAt:         nullTimeToString(p.UpdatedAt),
	}
}

func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
