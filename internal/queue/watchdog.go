package queue

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
	ws "github.com/abelkuruvilla/claw-agent-mission-control/internal/websocket"
)

// StuckTaskNotifier is implemented by the task handler so the watchdog can
// re-notify agents and parent orchestrators without duplicating logic.
type StuckTaskNotifier interface {
	NotifyAssignedAgent(agentID, taskID, title, description string)
	NotifyParentTaskAgent(ctx context.Context, subtask db.Task, newStatus string)
}

// Watchdog periodically finds tasks stuck in active states (executing, planning,
// discussing, verifying) and either re-notifies the agent or resets the task.
type Watchdog struct {
	store            *store.Store
	hub              *ws.Hub
	notifier         StuckTaskNotifier
	staleThreshold   time.Duration
	maxRetries       int
	stopChan         chan struct{}
	running          bool
}

// NewWatchdog creates a Watchdog. staleThreshold is how long without updated_at
// before a task is considered stuck; maxRetries is how many times to re-notify
// before resetting to backlog.
func NewWatchdog(st *store.Store, hub *ws.Hub, notifier StuckTaskNotifier, staleThreshold time.Duration, maxRetries int) *Watchdog {
	return &Watchdog{
		store:          st,
		hub:            hub,
		notifier:       notifier,
		staleThreshold: staleThreshold,
		maxRetries:    maxRetries,
		stopChan:      make(chan struct{}),
	}
}

// CheckOnce finds stale tasks and either re-notifies the agent or resets the task.
func (w *Watchdog) CheckOnce(ctx context.Context) {
	cutoff := time.Now().Add(-w.staleThreshold)
	stale, err := w.store.ListStaleTasks(ctx, cutoff)
	if err != nil {
		log.Printf("[Watchdog] Error listing stale tasks: %v", err)
		return
	}
	if len(stale) == 0 {
		log.Printf("[Watchdog] No stale tasks (cutoff: %v)", cutoff.UTC().Format(time.RFC3339))
		return
	}
	log.Printf("[Watchdog] Found %d stale task(s), processing...", len(stale))

	retried := 0
	reset := 0
	for _, task := range stale {
		taskID := task.ID
		agentID := ""
		if task.AgentID.Valid {
			agentID = task.AgentID.String
		}
		title := task.Title
		description := ""
		if task.Description.Valid {
			description = task.Description.String
		}

		if agentID != "" && task.RetryCount < int64(w.maxRetries) {
			// Re-notify same agent
			if err := w.store.IncrementTaskRetryCount(ctx, taskID); err != nil {
				log.Printf("[Watchdog] Error incrementing retry count for task %s: %v", taskID, err)
				continue
			}
			event, _ := w.store.CreateEvent(ctx, db.CreateEventParams{
				TaskID:  sql.NullString{String: taskID, Valid: true},
				AgentID: sql.NullString{String: agentID, Valid: true},
				Type:    "task_stuck_retry",
				Message: fmt.Sprintf("Task \"%s\" stuck (no update for %v) — re-notifying agent %s (retry %d/%d)", title, w.staleThreshold, agentID, task.RetryCount+1, w.maxRetries),
				Details: sql.NullString{String: fmt.Sprintf(`{"retry_count":%d}`, task.RetryCount+1), Valid: true},
			})
			if event.ID != "" && w.hub != nil {
				w.hub.BroadcastEvent(event)
			}
			_, _ = w.store.CreateComment(ctx, db.CreateCommentParams{
				TaskID:  taskID,
				Author:  "system",
				Content: fmt.Sprintf("[Watchdog] Task considered stuck (no update for %v). Re-notifying agent %s (retry %d/%d).", w.staleThreshold, agentID, task.RetryCount+1, w.maxRetries),
			})
			log.Printf("[Watchdog] Re-notifying agent %s for stuck task %s (%s)", agentID, taskID, title)
			w.notifier.NotifyAssignedAgent(agentID, taskID, title, description)
			retried++
		} else {
			// Max retries exceeded or no agent — reset to backlog
			if err := w.store.ResetStuckTask(ctx, taskID); err != nil {
				log.Printf("[Watchdog] Error resetting stuck task %s: %v", taskID, err)
				continue
			}
			reason := "max retries exceeded"
			if agentID == "" {
				reason = "no assigned agent"
			}
			event, _ := w.store.CreateEvent(ctx, db.CreateEventParams{
				TaskID:  sql.NullString{String: taskID, Valid: true},
				AgentID: sql.NullString{String: agentID, Valid: agentID != ""},
				Type:    "task_stuck_reset",
				Message: fmt.Sprintf("Task \"%s\" reset to backlog (%s)", title, reason),
				Details: sql.NullString{Valid: false},
			})
			if event.ID != "" && w.hub != nil {
				w.hub.BroadcastEvent(event)
			}
			_, _ = w.store.CreateComment(ctx, db.CreateCommentParams{
				TaskID:  taskID,
				Author:  "system",
				Content: fmt.Sprintf("[Watchdog] Task reset to backlog (%s). You can re-assign or use Retry from the UI.", reason),
			})
			if w.hub != nil {
				w.hub.BroadcastTaskStatus(taskID, "backlog", 0)
			}
			// If this was a subtask, notify parent orchestrator so the chain can recover
			if task.ParentTaskID.Valid && task.ParentTaskID.String != "" {
				subtaskCopy := task
				subtaskCopy.Status = sql.NullString{String: "failed", Valid: true}
				w.notifier.NotifyParentTaskAgent(ctx, subtaskCopy, "failed")
			}
			log.Printf("[Watchdog] Reset stuck task %s (%s) to backlog", taskID, title)
			reset++
		}
	}
	log.Printf("[Watchdog] Check complete: %d re-notified, %d reset", retried, reset)
}

// Start runs the watchdog periodically. Interval is how often to run CheckOnce.
func (w *Watchdog) Start(ctx context.Context, interval time.Duration) {
	if w.running {
		log.Println("[Watchdog] Already running")
		return
	}
	w.running = true
	log.Printf("[Watchdog] Starting (interval=%v, stale_threshold=%v, max_retries=%d)", interval, w.staleThreshold, w.maxRetries)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.CheckOnce(ctx)
			case <-w.stopChan:
				log.Println("[Watchdog] Stopping")
				w.running = false
				return
			case <-ctx.Done():
				log.Println("[Watchdog] Context cancelled, stopping")
				w.running = false
				return
			}
		}
	}()
}

// Stop stops the watchdog.
func (w *Watchdog) Stop() {
	if !w.running {
		return
	}
	close(w.stopChan)
	w.running = false
}
