package handlers

import (
	"database/sql"
	
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
)

// Clean response types that serialize properly to JSON
// These avoid the sql.NullString {String: "", Valid: bool} issue

type AgentResponse struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	Status           string  `json:"status"`
	WorkspacePath    *string `json:"workspace_path,omitempty"`
	AgentDirPath     *string `json:"agent_dir_path,omitempty"`
	Model            *string `json:"model,omitempty"`
	MentionPatterns  *string `json:"mention_patterns,omitempty"`
	SoulMD           *string `json:"soul_md,omitempty"`
	AgentsMD         *string `json:"agents_md,omitempty"`
	IdentityMD       *string `json:"identity_md,omitempty"`
	UserMD           *string `json:"user_md,omitempty"`
	ToolsMD          *string `json:"tools_md,omitempty"`
	HeartbeatMD      *string `json:"heartbeat_md,omitempty"`
	MemoryMD         *string `json:"memory_md,omitempty"`
	ActiveSessionKey *string `json:"active_session_key,omitempty"`
	CurrentTaskID    *string `json:"current_task_id,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type TaskResponse struct {
	ID             string  `json:"id"`
	Title          string  `json:"title"`
	Description    *string `json:"description,omitempty"`
	AgentID        *string `json:"agent_id,omitempty"`
	ProjectID      *string `json:"project_id,omitempty"`
	ParentTaskID   *string `json:"parent_task_id,omitempty"`
	Status         string  `json:"status"`
	Priority       int     `json:"priority"`
	GitBranch      *string `json:"git_branch,omitempty"`
	ProjectMD      *string `json:"project_md,omitempty"`
	RequirementsMD *string `json:"requirements_md,omitempty"`
	RoadmapMD      *string `json:"roadmap_md,omitempty"`
	StateMD        *string `json:"state_md,omitempty"`
	PrdJSON        *string `json:"prd_json,omitempty"`
	ProgressTxt    *string `json:"progress_txt,omitempty"`
	QualityChecks  *string `json:"quality_checks,omitempty"`
	DelegationMode string  `json:"delegation_mode"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	StartedAt      *string `json:"started_at,omitempty"`
	CompletedAt    *string `json:"completed_at,omitempty"`
	ScheduledAt    *string `json:"scheduled_at,omitempty"`
	RetryAt        *string `json:"retry_at,omitempty"`
	StoriesTotal   int     `json:"stories_total,omitempty"`
	StoriesPassed  int     `json:"stories_passed,omitempty"`
}
// Note: No "approach" field - all tasks use GSD for planning and Ralph Loop for execution

// Conversion functions
func nullStr(s interface{ String() string; Valid() bool }) *string {
	// This won't work with sql.NullString directly, use the helpers below
	return nil
}

func strPtr(s string, valid bool) *string {
	if !valid || s == "" {
		return nil
	}
	return &s
}

func ToAgentResponse(a db.Agent) AgentResponse {
	status := "idle"
	if a.Status.Valid {
		status = a.Status.String
	}
	
	return AgentResponse{
		ID:               a.ID,
		Name:             a.Name,
		Description:      strPtr(a.Description.String, a.Description.Valid),
		Status:           status,
		WorkspacePath:    strPtr(a.WorkspacePath.String, a.WorkspacePath.Valid),
		AgentDirPath:     strPtr(a.AgentDirPath.String, a.AgentDirPath.Valid),
		Model:            strPtr(a.Model.String, a.Model.Valid),
		MentionPatterns:  strPtr(a.MentionPatterns.String, a.MentionPatterns.Valid),
		SoulMD:           strPtr(a.SoulMd.String, a.SoulMd.Valid),
		AgentsMD:         strPtr(a.AgentsMd.String, a.AgentsMd.Valid),
		IdentityMD:       strPtr(a.IdentityMd.String, a.IdentityMd.Valid),
		UserMD:           strPtr(a.UserMd.String, a.UserMd.Valid),
		ToolsMD:          strPtr(a.ToolsMd.String, a.ToolsMd.Valid),
		HeartbeatMD:      strPtr(a.HeartbeatMd.String, a.HeartbeatMd.Valid),
		MemoryMD:         strPtr(a.MemoryMd.String, a.MemoryMd.Valid),
		ActiveSessionKey: strPtr(a.ActiveSessionKey.String, a.ActiveSessionKey.Valid),
		CurrentTaskID:    strPtr(a.CurrentTaskID.String, a.CurrentTaskID.Valid),
		CreatedAt:        a.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        a.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
}

func ToAgentResponses(agents []db.Agent) []AgentResponse {
	result := make([]AgentResponse, len(agents))
	for i, a := range agents {
		result[i] = ToAgentResponse(a)
	}
	return result
}

func ToTaskResponse(t db.Task) TaskResponse {
	status := "backlog"
	if t.Status.Valid {
		status = t.Status.String
	}
	
	priority := 3
	if t.Priority.Valid {
		priority = int(t.Priority.Int64)
	}
	
	delegationMode := "auto"
	if t.DelegationMode.Valid && t.DelegationMode.String != "" {
		delegationMode = t.DelegationMode.String
	}

	resp := TaskResponse{
		ID:             t.ID,
		Title:          t.Title,
		Description:    strPtr(t.Description.String, t.Description.Valid),
		AgentID:        strPtr(t.AgentID.String, t.AgentID.Valid),
		ProjectID:      strPtr(t.ProjectID.String, t.ProjectID.Valid),
		ParentTaskID:   strPtr(t.ParentTaskID.String, t.ParentTaskID.Valid),
		Status:         status,
		Priority:       priority,
		GitBranch:      strPtr(t.GitBranch.String, t.GitBranch.Valid),
		ProjectMD:      strPtr(t.ProjectMd.String, t.ProjectMd.Valid),
		RequirementsMD: strPtr(t.RequirementsMd.String, t.RequirementsMd.Valid),
		RoadmapMD:      strPtr(t.RoadmapMd.String, t.RoadmapMd.Valid),
		StateMD:        strPtr(t.StateMd.String, t.StateMd.Valid),
		PrdJSON:        strPtr(t.PrdJson.String, t.PrdJson.Valid),
		ProgressTxt:    strPtr(t.ProgressTxt.String, t.ProgressTxt.Valid),
		QualityChecks:  strPtr(t.QualityChecks.String, t.QualityChecks.Valid),
		DelegationMode: delegationMode,
		CreatedAt:      t.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      t.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
	}
	
	if t.StartedAt.Valid {
		s := t.StartedAt.Time.Format("2006-01-02T15:04:05Z")
		resp.StartedAt = &s
	}
	if t.CompletedAt.Valid {
		s := t.CompletedAt.Time.Format("2006-01-02T15:04:05Z")
		resp.CompletedAt = &s
	}
	if t.ScheduledAt.Valid {
		s := t.ScheduledAt.Time.Format("2006-01-02T15:04:05Z")
		resp.ScheduledAt = &s
	}
	if t.RetryAt.Valid {
		s := t.RetryAt.Time.Format("2006-01-02T15:04:05Z")
		resp.RetryAt = &s
	}
	
	return resp
}

func ToTaskResponses(tasks []db.Task) []TaskResponse {
	result := make([]TaskResponse, len(tasks))
	for i, t := range tasks {
		result[i] = ToTaskResponse(t)
	}
	return result
}

// Helper function to convert sql.NullTime to string
func nullTimeToString(nt sql.NullTime) string {
	if nt.Valid {
		return nt.Time.Format("2006-01-02T15:04:05Z")
	}
	return ""
}
