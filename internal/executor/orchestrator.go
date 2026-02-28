package executor

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/openclaw"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
	ws "github.com/abelkuruvilla/claw-agent-mission-control/internal/websocket"
)

type Orchestrator struct {
	apiBaseURL     string
	openclawClient *openclaw.Client
	store          *store.Store
	hub            *ws.Hub

	gsdEngine   *GSDEngine
	ralphEngine *RalphEngine

	// Track running tasks
	running   map[string]context.CancelFunc
	runningMu sync.RWMutex

	maxParallel int
}

func NewOrchestrator(apiBaseURL string, oc *openclaw.Client, s *store.Store, hub *ws.Hub, maxParallel int) *Orchestrator {
	if maxParallel <= 0 {
		maxParallel = 3
	}

	o := &Orchestrator{
		apiBaseURL:     apiBaseURL,
		openclawClient: oc,
		store:          s,
		hub:            hub,
		running:        make(map[string]context.CancelFunc),
		maxParallel:    maxParallel,
	}

	o.gsdEngine = NewGSDEngine(apiBaseURL, oc, s, hub)
	o.ralphEngine = NewRalphEngine(apiBaseURL, oc, s, hub, 10)

	return o
}

// StartTask begins execution of a task
func (o *Orchestrator) StartTask(ctx context.Context, taskID string) error {
	// Check if already running
	o.runningMu.RLock()
	if _, exists := o.running[taskID]; exists {
		o.runningMu.RUnlock()
		return fmt.Errorf("task %s is already running", taskID)
	}
	o.runningMu.RUnlock()

	// Check parallel limit
	o.runningMu.RLock()
	if len(o.running) >= o.maxParallel {
		o.runningMu.RUnlock()
		return fmt.Errorf("max parallel tasks (%d) reached", o.maxParallel)
	}
	o.runningMu.RUnlock()

	// Get task
	task, err := o.store.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// Create cancellable context
	taskCtx, cancel := context.WithCancel(ctx)

	// Register as running
	o.runningMu.Lock()
	o.running[taskID] = cancel
	o.runningMu.Unlock()

	// Update task status
	o.store.UpdateTaskStatus(ctx, taskID, "executing")

	// Log event
	o.logEvent(ctx, taskID, "task_started", fmt.Sprintf("Task '%s' execution started", task.Title))

	// Broadcast
	if o.hub != nil {
		o.hub.BroadcastTaskStatus(taskID, "executing", 0)
	}

	// Run in background
	go func() {
		defer func() {
			o.runningMu.Lock()
			delete(o.running, taskID)
			o.runningMu.Unlock()
		}()

		var execErr error

		// Select engine based on approach
		switch task.Approach.String {
		case "ralph":
			execErr = o.ralphEngine.Run(taskCtx, task)
		default: // "gsd" or empty
			execErr = o.gsdEngine.ExecuteTask(taskCtx, task)
		}

		if execErr != nil {
			o.store.UpdateTaskStatus(context.Background(), taskID, "failed")
			o.logEvent(context.Background(), taskID, "task_failed", execErr.Error())
		} else {
			o.logEvent(context.Background(), taskID, "task_completed", "Task completed successfully")
		}

		if o.hub != nil {
			status := "done"
			if execErr != nil {
				status = "failed"
			}
			o.hub.BroadcastTaskStatus(taskID, status, 1.0)
		}
	}()

	return nil
}

// StopTask cancels a running task
func (o *Orchestrator) StopTask(taskID string) error {
	o.runningMu.Lock()
	defer o.runningMu.Unlock()

	cancel, exists := o.running[taskID]
	if !exists {
		return fmt.Errorf("task %s is not running", taskID)
	}

	cancel()
	delete(o.running, taskID)

	o.store.UpdateTaskStatus(context.Background(), taskID, "cancelled")
	o.logEvent(context.Background(), taskID, "task_cancelled", "Task was cancelled")

	return nil
}

// PauseTask pauses a running task
func (o *Orchestrator) PauseTask(taskID string) error {
	// For now, pause = stop. Future: implement proper pause/resume
	return o.StopTask(taskID)
}

// GetRunningTasks returns list of currently running task IDs
func (o *Orchestrator) GetRunningTasks() []string {
	o.runningMu.RLock()
	defer o.runningMu.RUnlock()

	tasks := make([]string, 0, len(o.running))
	for id := range o.running {
		tasks = append(tasks, id)
	}
	return tasks
}

// IsRunning checks if a task is currently running
func (o *Orchestrator) IsRunning(taskID string) bool {
	o.runningMu.RLock()
	defer o.runningMu.RUnlock()
	_, exists := o.running[taskID]
	return exists
}

func (o *Orchestrator) logEvent(ctx context.Context, taskID, eventType, message string) {
	o.store.CreateEvent(ctx, db.CreateEventParams{
		TaskID:  sql.NullString{String: taskID, Valid: true},
		Type:    eventType,
		Message: message,
	})
}
