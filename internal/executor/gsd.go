package executor

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/openclaw"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
	ws "github.com/abelkuruvilla/claw-agent-mission-control/internal/websocket"
)

type GSDEngine struct {
	apiBaseURL     string
	openclawClient *openclaw.Client
	store          *store.Store
	hub            *ws.Hub
}

func NewGSDEngine(apiBaseURL string, oc *openclaw.Client, s *store.Store, hub *ws.Hub) *GSDEngine {
	return &GSDEngine{
		apiBaseURL:     apiBaseURL,
		openclawClient: oc,
		store:          s,
		hub:            hub,
	}
}

// ExecuteTask runs the GSD workflow for a task
func (e *GSDEngine) ExecuteTask(ctx context.Context, task db.Task) error {
	// Get all phases
	phases, err := e.store.ListPhasesByTask(ctx, task.ID)
	if err != nil {
		return err
	}

	for _, phase := range phases {
		if phase.Status.String == "done" {
			continue // Skip completed phases
		}

		if err := e.ExecutePhase(ctx, task, phase); err != nil {
			// Log error but continue to allow retry
			e.logEvent(ctx, task.ID, "phase_error", err.Error())
			return err
		}
	}

	// All phases complete
	e.store.UpdateTaskStatus(ctx, task.ID, "done")
	return nil
}

// ExecutePhase runs a single phase
func (e *GSDEngine) ExecutePhase(ctx context.Context, task db.Task, phase db.Phase) error {
	// Update phase status
	e.store.UpdatePhaseStatus(ctx, phase.ID, "executing")

	// Generate execution token (simplified - in production use JWT)
	token := fmt.Sprintf("exec-%s-%d", phase.ID, time.Now().Unix())

	// Build prompt
	prompt := e.buildExecutePrompt(task, phase, token)

	// Spawn fresh session
	resp, err := e.openclawClient.Spawn(ctx, &openclaw.SpawnRequest{
		Task:           prompt,
		AgentID:        task.AgentID.String,
		Label:          fmt.Sprintf("gsd-phase-%s", phase.ID),
		Cleanup:        "delete",
		TimeoutSeconds: 1800, // 30 minutes
	})
	if err != nil {
		e.store.UpdatePhaseStatus(ctx, phase.ID, "error")
		return fmt.Errorf("failed to spawn session: %w", err)
	}

	// Log event
	e.logEvent(ctx, task.ID, "phase_started", fmt.Sprintf("Phase %d started: %s (session: %s)", phase.Sequence, phase.Title, resp.ChildSessionKey))

	// Broadcast update
	if e.hub != nil {
		// Calculate progress based on phase count
		phases, _ := e.store.ListPhasesByTask(ctx, task.ID)
		progress := float64(phase.Sequence) / float64(len(phases)) * 100
		e.hub.BroadcastTaskStatus(task.ID, "executing", progress)
	}

	return nil
}

func (e *GSDEngine) buildExecutePrompt(task db.Task, phase db.Phase, token string) string {
	return fmt.Sprintf(`# Task Execution Context

## Mission Control API
**Base URL:** %s
**Auth Token:** %s

### Required API Calls

1. **Update Progress** (call periodically):
curl -X POST %s/api/v1/phases/%s/progress \
  -H "Content-Type: application/json" \
  -d '{"progress": 0.5, "message": "Working on..."}'

2. **Mark Phase Complete** (when done):
curl -X POST %s/api/v1/phases/%s/complete \
  -H "Content-Type: application/json" \
  -d '{"summary": "Completed...", "artifacts": {}}'

3. **Report Failure** (if blocked):
curl -X POST %s/api/v1/phases/%s/fail \
  -H "Content-Type: application/json" \
  -d '{"error": "...", "recoverable": true}'

---

## Task Details

**Task ID:** %s
**Title:** %s
**Description:** %s
**Working Directory:** %s

---

## Current Phase

**Phase ID:** %s
**Phase:** %d - %s
**Description:** %s

---

## GSD Workflow Instructions

You are executing Phase %d of a GSD workflow.

### Your Process:
1. Read and understand the phase requirements
2. Execute the work atomically
3. Commit changes to git with descriptive messages
4. Report progress via API periodically
5. Run any verification/tests
6. Call the complete API with summary when done

### On Error:
If you encounter a blocking issue, call the fail API with details.

---

## Begin Execution

Start working on this phase. Report progress and call complete when done.
`,
		e.apiBaseURL, token,
		e.apiBaseURL, phase.ID,
		e.apiBaseURL, phase.ID,
		e.apiBaseURL, phase.ID,
		task.ID, task.Title, task.Description.String, task.WorkDir.String,
		phase.ID, phase.Sequence, phase.Title, phase.Description.String,
		phase.Sequence,
	)
}

func (e *GSDEngine) logEvent(ctx context.Context, taskID, eventType, message string) {
	e.store.CreateEvent(ctx, db.CreateEventParams{
		TaskID:  sql.NullString{String: taskID, Valid: true},
		Type:    eventType,
		Message: message,
	})
}
