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

type RalphEngine struct {
	apiBaseURL     string
	openclawClient *openclaw.Client
	store          *store.Store
	hub            *ws.Hub
	maxIterations  int
}

func NewRalphEngine(apiBaseURL string, oc *openclaw.Client, s *store.Store, hub *ws.Hub, maxIter int) *RalphEngine {
	if maxIter <= 0 {
		maxIter = 10
	}
	return &RalphEngine{
		apiBaseURL:     apiBaseURL,
		openclawClient: oc,
		store:          s,
		hub:            hub,
		maxIterations:  maxIter,
	}
}

// Run executes the Ralph loop until all stories pass or max iterations reached
func (e *RalphEngine) Run(ctx context.Context, task db.Task) error {
	for iteration := 0; iteration < e.maxIterations; iteration++ {
		// Refresh task from DB
		task, _ = e.store.GetTask(ctx, task.ID)

		// Check if all stories pass
		passed, total, _ := e.store.GetStoryProgress(ctx, task.ID)
		if passed == total && total > 0 {
			e.store.UpdateTaskStatus(ctx, task.ID, "done")
			e.logEvent(ctx, task.ID, "task_completed", fmt.Sprintf("All %d stories passed", total))
			return nil
		}

		// Get next pending story
		story, err := e.store.GetNextPendingStory(ctx, task.ID)
		if err != nil {
			// No more pending stories
			e.store.UpdateTaskStatus(ctx, task.ID, "done")
			return nil
		}

		// Execute story
		if err := e.ExecuteStory(ctx, task, story, iteration); err != nil {
			e.logEvent(ctx, task.ID, "story_error", err.Error())
			// Continue to next iteration
		}

		// Small delay between iterations
		time.Sleep(2 * time.Second)
	}

	e.store.UpdateTaskStatus(ctx, task.ID, "failed")
	return fmt.Errorf("max iterations (%d) reached", e.maxIterations)
}

// ExecuteStory runs a single story iteration
func (e *RalphEngine) ExecuteStory(ctx context.Context, task db.Task, story db.Story, iteration int) error {
	// Generate token
	token := fmt.Sprintf("ralph-%s-%d", story.ID, time.Now().Unix())

	// Build prompt
	prompt := e.buildStoryPrompt(task, story, iteration, token)

	// Spawn fresh session
	resp, err := e.openclawClient.Spawn(ctx, &openclaw.SpawnRequest{
		Task:           prompt,
		AgentID:        task.AgentID.String,
		Label:          fmt.Sprintf("ralph-%s-story-%s-iter-%d", task.ID, story.ID, iteration),
		Cleanup:        "delete",
		TimeoutSeconds: 1200, // 20 minutes per story
	})
	if err != nil {
		return fmt.Errorf("failed to spawn session: %w", err)
	}

	// Log event
	e.logEvent(ctx, task.ID, "story_started",
		fmt.Sprintf("Story '%s' iteration %d started (session: %s)", story.Title, iteration, resp.ChildSessionKey))

	// Broadcast
	if e.hub != nil {
		passed, total, _ := e.store.GetStoryProgress(ctx, task.ID)
		progress := float64(passed) / float64(total)
		e.hub.BroadcastTaskStatus(task.ID, "executing", progress)
	}

	return nil
}

func (e *RalphEngine) buildStoryPrompt(task db.Task, story db.Story, iteration int, token string) string {
	return fmt.Sprintf(`# Ralph Loop Execution Context

## Mission Control API
**Base URL:** %s
**Auth Token:** %s

### Required API Calls

1. **Mark Story Passed** (when tests pass):
curl -X POST %s/api/v1/stories/%s/pass \
  -H "Content-Type: application/json" \
  -d '{"commit_sha": "<sha>", "learnings": "<what you learned>"}'

2. **Mark Story Failed** (if tests fail):
curl -X POST %s/api/v1/stories/%s/fail \
  -H "Content-Type: application/json" \
  -d '{"error": "<error message>", "iteration": %d}'

3. **Append Learnings**:
curl -X POST %s/api/v1/tasks/%s/progress-txt \
  -H "Content-Type: application/json" \
  -d '{"content": "<learnings from this iteration>"}'

---

## Task: %s

**Task ID:** %s
**Working Directory:** %s
**Iteration:** %d of %d

---

## Current Story

**Story ID:** %s
**Title:** %s
**Priority:** %d

**Description:**
%s

**Acceptance Criteria:**
%s

---

## Ralph Workflow

1. Read and understand the story
2. Implement the feature/fix
3. Run quality checks / tests
4. If tests PASS:
   - git add + commit with descriptive message
   - Call /stories/%s/pass with commit SHA and learnings
5. If tests FAIL:
   - Call /stories/%s/fail with error details
   - DO NOT commit broken code

---

## Begin

Implement this story. Focus on THIS STORY ONLY.
When done, call the appropriate API endpoint.
`,
		e.apiBaseURL, token,
		e.apiBaseURL, story.ID,
		e.apiBaseURL, story.ID, iteration,
		e.apiBaseURL, task.ID,
		task.Title, task.ID, task.WorkDir.String, iteration, e.maxIterations,
		story.ID, story.Title, story.Priority.Int64,
		story.Description.String,
		story.AcceptanceCriteria.String,
		story.ID, story.ID,
	)
}

func (e *RalphEngine) logEvent(ctx context.Context, taskID, eventType, message string) {
	e.store.CreateEvent(ctx, db.CreateEventParams{
		TaskID:  sql.NullString{String: taskID, Valid: true},
		Type:    eventType,
		Message: message,
	})
}
