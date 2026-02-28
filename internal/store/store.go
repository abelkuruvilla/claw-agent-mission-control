package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/google/uuid"
)

type Store struct {
	db      *sql.DB
	queries *db.Queries
}

func New(database *sql.DB) *Store {
	return &Store{
		db:      database,
		queries: db.New(database),
	}
}

// Transaction helper
func (s *Store) WithTx(ctx context.Context, fn func(*Store) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txStore := &Store{
		db:      s.db,
		queries: db.New(tx),
	}

	if err := fn(txStore); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ============ Agents ============

func (s *Store) CreateAgent(ctx context.Context, params db.CreateAgentParams) (db.Agent, error) {
	if params.ID == "" {
		params.ID = uuid.New().String()
	}
	return s.queries.CreateAgent(ctx, params)
}

func (s *Store) GetAgent(ctx context.Context, id string) (db.Agent, error) {
	return s.queries.GetAgent(ctx, id)
}

func (s *Store) ListAgents(ctx context.Context) ([]db.Agent, error) {
	return s.queries.ListAgents(ctx)
}

func (s *Store) UpdateAgent(ctx context.Context, params db.UpdateAgentParams) (db.Agent, error) {
	return s.queries.UpdateAgent(ctx, params)
}

func (s *Store) DeleteAgent(ctx context.Context, id string) error {
	return s.queries.DeleteAgent(ctx, id)
}

func (s *Store) UpdateAgentStatus(ctx context.Context, id, status string) error {
	return s.queries.UpdateAgentStatus(ctx, db.UpdateAgentStatusParams{
		Status: sql.NullString{String: status, Valid: true},
		ID:     id,
	})
}

// ============ Tasks ============

func (s *Store) CreateTask(ctx context.Context, params db.CreateTaskParams) (db.Task, error) {
	if params.ID == "" {
		params.ID = uuid.New().String()
	}
	return s.queries.CreateTask(ctx, params)
}

func (s *Store) GetTask(ctx context.Context, id string) (db.Task, error) {
	return s.queries.GetTask(ctx, id)
}

func (s *Store) ListTasks(ctx context.Context) ([]db.Task, error) {
	return s.queries.ListTasks(ctx)
}

func (s *Store) ListTasksByStatus(ctx context.Context, status string) ([]db.Task, error) {
	return s.queries.ListTasksByStatus(ctx, sql.NullString{String: status, Valid: true})
}

func (s *Store) ListTasksByAgent(ctx context.Context, agentID string) ([]db.Task, error) {
	return s.queries.ListTasksByAgent(ctx, sql.NullString{String: agentID, Valid: true})
}

func (s *Store) UpdateTask(ctx context.Context, params db.UpdateTaskParams) (db.Task, error) {
	return s.queries.UpdateTask(ctx, params)
}

func (s *Store) UpdateTaskStatus(ctx context.Context, id, status string) error {
	return s.queries.UpdateTaskStatus(ctx, db.UpdateTaskStatusParams{
		Status: sql.NullString{String: status, Valid: true},
		ID:     id,
	})
}

func (s *Store) DeleteTask(ctx context.Context, id string) error {
	return s.queries.DeleteTask(ctx, id)
}

func (s *Store) ListQueuedTasksByAgent(ctx context.Context, agentID string) ([]db.Task, error) {
	return s.queries.ListQueuedTasksByAgent(ctx, sql.NullString{String: agentID, Valid: true})
}

func (s *Store) CountActiveTasksByAgent(ctx context.Context, agentID string) (int64, error) {
	return s.queries.CountActiveTasksByAgent(ctx, sql.NullString{String: agentID, Valid: true})
}

// ListStaleTasks returns tasks in active status (executing, planning, discussing, verifying)
// whose updated_at is older than the given cutoff (or NULL). Used by the stuck-task watchdog.
func (s *Store) ListStaleTasks(ctx context.Context, cutoff time.Time) ([]db.Task, error) {
	return s.queries.ListStaleTasks(ctx, sql.NullTime{Time: cutoff, Valid: true})
}

// IncrementTaskRetryCount bumps retry_count and updated_at for a task (watchdog re-notify).
func (s *Store) IncrementTaskRetryCount(ctx context.Context, taskID string) error {
	return s.queries.IncrementTaskRetryCount(ctx, taskID)
}

// ResetStuckTask sets status to backlog, clears agent_id and retry_count (watchdog after max retries).
func (s *Store) ResetStuckTask(ctx context.Context, taskID string) error {
	return s.queries.ResetStuckTask(ctx, taskID)
}

// ResetTaskRetryCount clears retry_count for a task (on normal status transition).
func (s *Store) ResetTaskRetryCount(ctx context.Context, taskID string) error {
	return s.queries.ResetTaskRetryCount(ctx, taskID)
}

// AppendProgressTxt appends content to a task's progress_txt and sets updated_at to now.
func (s *Store) AppendProgressTxt(ctx context.Context, taskID, content string) error {
	return s.queries.AppendProgressTxt(ctx, db.AppendProgressTxtParams{
		ProgressTxt: sql.NullString{String: content, Valid: true},
		ID:          taskID,
	})
}

// ============ Phases ============

func (s *Store) CreatePhase(ctx context.Context, params db.CreatePhaseParams) (db.Phase, error) {
	if params.ID == "" {
		params.ID = uuid.New().String()
	}
	return s.queries.CreatePhase(ctx, params)
}

func (s *Store) GetPhase(ctx context.Context, id string) (db.Phase, error) {
	return s.queries.GetPhase(ctx, id)
}

func (s *Store) ListPhasesByTask(ctx context.Context, taskID string) ([]db.Phase, error) {
	return s.queries.ListPhasesByTask(ctx, taskID)
}

func (s *Store) UpdatePhase(ctx context.Context, params db.UpdatePhaseParams) (db.Phase, error) {
	return s.queries.UpdatePhase(ctx, params)
}

func (s *Store) UpdatePhaseStatus(ctx context.Context, id, status string) error {
	return s.queries.UpdatePhaseStatus(ctx, db.UpdatePhaseStatusParams{
		Status: sql.NullString{String: status, Valid: true},
		ID:     id,
	})
}

func (s *Store) DeletePhase(ctx context.Context, id string) error {
	return s.queries.DeletePhase(ctx, id)
}

// ============ Stories ============

func (s *Store) CreateStory(ctx context.Context, params db.CreateStoryParams) (db.Story, error) {
	if params.ID == "" {
		params.ID = uuid.New().String()
	}
	return s.queries.CreateStory(ctx, params)
}

func (s *Store) GetStory(ctx context.Context, id string) (db.Story, error) {
	return s.queries.GetStory(ctx, id)
}

func (s *Store) ListStoriesByTask(ctx context.Context, taskID string) ([]db.Story, error) {
	return s.queries.ListStoriesByTask(ctx, taskID)
}

func (s *Store) GetNextPendingStory(ctx context.Context, taskID string) (db.Story, error) {
	return s.queries.GetNextPendingStory(ctx, taskID)
}

func (s *Store) UpdateStory(ctx context.Context, params db.UpdateStoryParams) (db.Story, error) {
	return s.queries.UpdateStory(ctx, params)
}

func (s *Store) MarkStoryPassed(ctx context.Context, id string) error {
	return s.queries.MarkStoryPassed(ctx, id)
}

func (s *Store) MarkStoryFailed(ctx context.Context, id, lastError string) error {
	return s.queries.MarkStoryFailed(ctx, db.MarkStoryFailedParams{
		LastError: sql.NullString{String: lastError, Valid: true},
		ID:        id,
	})
}

func (s *Store) DeleteStory(ctx context.Context, id string) error {
	return s.queries.DeleteStory(ctx, id)
}

func (s *Store) GetStoryProgress(ctx context.Context, taskID string) (passed, total int64, err error) {
	passed, err = s.queries.CountPassedStories(ctx, taskID)
	if err != nil {
		return 0, 0, err
	}
	total, err = s.queries.CountTotalStories(ctx, taskID)
	return passed, total, err
}

// ============ SubAgents ============

func (s *Store) CreateSubAgent(ctx context.Context, params db.CreateSubAgentParams) (db.SubAgent, error) {
	if params.ID == "" {
		params.ID = uuid.New().String()
	}
	return s.queries.CreateSubAgent(ctx, params)
}

func (s *Store) GetSubAgent(ctx context.Context, id string) (db.SubAgent, error) {
	return s.queries.GetSubAgent(ctx, id)
}

func (s *Store) ListSubAgentsByOrchestrator(ctx context.Context, orchestratorID string) ([]db.SubAgent, error) {
	return s.queries.ListSubAgentsByOrchestrator(ctx, orchestratorID)
}

func (s *Store) ListSubAgentsByTask(ctx context.Context, taskID string) ([]db.SubAgent, error) {
	return s.queries.ListSubAgentsByTask(ctx, sql.NullString{String: taskID, Valid: true})
}

func (s *Store) UpdateSubAgentStatus(ctx context.Context, id, status, output, errMsg string) error {
	return s.queries.UpdateSubAgentStatus(ctx, db.UpdateSubAgentStatusParams{
		Status: sql.NullString{String: status, Valid: true},
		Output: sql.NullString{String: output, Valid: output != ""},
		Error:  sql.NullString{String: errMsg, Valid: errMsg != ""},
		ID:     id,
	})
}

// ============ Events ============

func (s *Store) CreateEvent(ctx context.Context, params db.CreateEventParams) (db.Event, error) {
	if params.ID == "" {
		params.ID = uuid.New().String()
	}
	return s.queries.CreateEvent(ctx, params)
}

func (s *Store) ListEvents(ctx context.Context, limit int64) ([]db.Event, error) {
	return s.queries.ListEvents(ctx, limit)
}

func (s *Store) ListEventsByTask(ctx context.Context, taskID string, limit int64) ([]db.Event, error) {
	return s.queries.ListEventsByTask(ctx, db.ListEventsByTaskParams{
		TaskID: sql.NullString{String: taskID, Valid: true},
		Limit:  limit,
	})
}

func (s *Store) ListEventsByAgent(ctx context.Context, agentID string, limit int64) ([]db.Event, error) {
	return s.queries.ListEventsByAgent(ctx, db.ListEventsByAgentParams{
		AgentID: sql.NullString{String: agentID, Valid: true},
		Limit:   limit,
	})
}

// ============ Settings ============

func (s *Store) GetSettings(ctx context.Context) (db.Setting, error) {
	return s.queries.GetSettings(ctx)
}

func (s *Store) UpdateSettings(ctx context.Context, params db.UpdateSettingsParams) (db.Setting, error) {
	return s.queries.UpdateSettings(ctx, params)
}

// ============ Projects ============

func (s *Store) CreateProject(ctx context.Context, params db.CreateProjectParams) (db.Project, error) {
	if params.ID == "" {
		params.ID = uuid.New().String()
	}
	return s.queries.CreateProject(ctx, params)
}

func (s *Store) GetProject(ctx context.Context, id string) (db.Project, error) {
	return s.queries.GetProject(ctx, id)
}

func (s *Store) ListProjects(ctx context.Context) ([]db.Project, error) {
	return s.queries.ListProjects(ctx)
}

func (s *Store) ListProjectsByStatus(ctx context.Context, status sql.NullString) ([]db.Project, error) {
	return s.queries.ListProjectsByStatus(ctx, status)
}

func (s *Store) UpdateProject(ctx context.Context, params db.UpdateProjectParams) (db.Project, error) {
	return s.queries.UpdateProject(ctx, params)
}

func (s *Store) DeleteProject(ctx context.Context, id string) error {
	return s.queries.DeleteProject(ctx, id)
}

func (s *Store) GetProjectTaskCount(ctx context.Context, projectID sql.NullString) (int64, error) {
	return s.queries.GetProjectTaskCount(ctx, projectID)
}

func (s *Store) GetProjectDoneTaskCount(ctx context.Context, projectID sql.NullString) (int64, error) {
	return s.queries.GetProjectDoneTaskCount(ctx, projectID)
}

func (s *Store) ListTasksByProject(ctx context.Context, projectID sql.NullString) ([]db.Task, error) {
	return s.queries.ListTasksByProject(ctx, projectID)
}

// ============ Comments ============

func (s *Store) CreateComment(ctx context.Context, params db.CreateCommentParams) (db.Comment, error) {
	if params.ID == "" {
		params.ID = uuid.New().String()
	}
	return s.queries.CreateComment(ctx, params)
}

func (s *Store) GetComment(ctx context.Context, id string) (db.Comment, error) {
	return s.queries.GetComment(ctx, id)
}

func (s *Store) ListCommentsByTask(ctx context.Context, taskID string) ([]db.Comment, error) {
	return s.queries.ListCommentsByTask(ctx, taskID)
}

func (s *Store) DeleteComment(ctx context.Context, id string) error {
	return s.queries.DeleteComment(ctx, id)
}

// ============ Chat Sessions ============

func (s *Store) CreateChatSession(ctx context.Context, params db.CreateChatSessionParams) (db.ChatSession, error) {
	if params.ID == "" {
		params.ID = uuid.New().String()
	}
	return s.queries.CreateChatSession(ctx, params)
}

func (s *Store) GetChatSession(ctx context.Context, id string) (db.ChatSession, error) {
	return s.queries.GetChatSession(ctx, id)
}

func (s *Store) ListChatSessionsByAgent(ctx context.Context, agentID string) ([]db.ChatSession, error) {
	return s.queries.ListChatSessionsByAgent(ctx, agentID)
}

func (s *Store) EndChatSession(ctx context.Context, id string) error {
	return s.queries.EndChatSession(ctx, id)
}

func (s *Store) UpdateMessageCount(ctx context.Context, id string) error {
	return s.queries.UpdateMessageCount(ctx, id)
}

// ============ Chat Messages ============

func (s *Store) CreateChatMessage(ctx context.Context, params db.CreateChatMessageParams) (db.ChatMessage, error) {
	if params.ID == "" {
		params.ID = uuid.New().String()
	}
	return s.queries.CreateChatMessage(ctx, params)
}

func (s *Store) ListMessagesBySession(ctx context.Context, sessionID string) ([]db.ChatMessage, error) {
	return s.queries.ListMessagesBySession(ctx, sessionID)
}

// ============ Task Dependencies ============

func (s *Store) ListSubtasks(ctx context.Context, parentTaskID sql.NullString) ([]db.Task, error) {
	return s.queries.ListSubtasks(ctx, parentTaskID)
}

// ============ Task Scheduling ============

func (s *Store) SetTaskScheduledAt(ctx context.Context, id string, t time.Time) error {
	return s.queries.SetTaskScheduledAt(ctx, db.SetTaskScheduledAtParams{
		ScheduledAt: sql.NullTime{Time: t, Valid: true},
		ID:          id,
	})
}

func (s *Store) SetTaskRetryAt(ctx context.Context, id string, t time.Time) error {
	return s.queries.SetTaskRetryAt(ctx, db.SetTaskRetryAtParams{
		RetryAt: sql.NullTime{Time: t, Valid: true},
		ID:      id,
	})
}

func (s *Store) ClearTaskScheduledAt(ctx context.Context, id string) error {
	return s.queries.ClearTaskScheduledAt(ctx, id)
}

func (s *Store) ClearTaskRetryAt(ctx context.Context, id string) error {
	return s.queries.ClearTaskRetryAt(ctx, id)
}

func (s *Store) ListScheduledDueTasks(ctx context.Context) ([]db.Task, error) {
	return s.queries.ListScheduledDueTasks(ctx)
}

func (s *Store) ListRetryDueTasks(ctx context.Context) ([]db.Task, error) {
	return s.queries.ListRetryDueTasks(ctx)
}
