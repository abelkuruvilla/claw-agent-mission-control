package queue

import (
	"context"
	"log"
	"time"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/openclaw"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
	ws "github.com/abelkuruvilla/claw-agent-mission-control/internal/websocket"
)

// AgentQueueProcessor is the interface that the task handler implements
// for dequeuing and notifying agents about queued tasks.
type AgentQueueProcessor interface {
	ProcessAgentQueue(ctx context.Context, agentID string)
}

// Processor periodically checks all agent queues and dispatches
// queued tasks to agents that have become free.
type Processor struct {
	store       *store.Store
	agentSender *openclaw.AgentSender
	hub         *ws.Hub
	handler     AgentQueueProcessor
	stopChan    chan struct{}
	running     bool
}

func NewProcessor(st *store.Store, agentSender *openclaw.AgentSender, hub *ws.Hub, handler AgentQueueProcessor) *Processor {
	return &Processor{
		store:       st,
		agentSender: agentSender,
		hub:         hub,
		handler:     handler,
		stopChan:    make(chan struct{}),
	}
}

// ProcessScheduledTasks dispatches due scheduled and retry tasks directly to agents.
// Unlike ProcessAgentQueue which only handles 'queued' tasks, this handles
// scheduled tasks that have status 'backlog' with a past scheduled_at time.
func (p *Processor) ProcessScheduledTasks(ctx context.Context) {
	dueTasks, err := p.store.ListScheduledDueTasks(ctx)
	if err != nil {
		log.Printf("[QueueProcessor] Error listing scheduled tasks: %v", err)
	} else {
		for _, task := range dueTasks {
			log.Printf("[QueueProcessor] Scheduled task %s (%s) is due — dispatching", task.ID, task.Title)
			if err := p.store.ClearTaskScheduledAt(ctx, task.ID); err != nil {
				log.Printf("[QueueProcessor] Error clearing scheduled_at for %s: %v", task.ID, err)
				continue
			}
			if task.AgentID.Valid && task.AgentID.String != "" {
				desc := ""
				if task.Description.Valid {
					desc = task.Description.String
				}
				p.dispatchTaskToAgent(ctx, task.ID, task.AgentID.String, task.Title, desc)
			}
		}
	}

	retryTasks, err := p.store.ListRetryDueTasks(ctx)
	if err != nil {
		log.Printf("[QueueProcessor] Error listing retry tasks: %v", err)
	} else {
		for _, task := range retryTasks {
			log.Printf("[QueueProcessor] Scheduled retry for task %s (%s) is due — dispatching", task.ID, task.Title)
			if err := p.store.ClearTaskRetryAt(ctx, task.ID); err != nil {
				log.Printf("[QueueProcessor] Error clearing retry_at for %s: %v", task.ID, err)
				continue
			}
			if task.AgentID.Valid && task.AgentID.String != "" {
				desc := ""
				if task.Description.Valid {
					desc = task.Description.String
				}
				p.dispatchTaskToAgent(ctx, task.ID, task.AgentID.String, task.Title, desc)
			}
		}
	}
}

// dispatchTaskToAgent sends a specific task to an agent.
// If the agent is busy, the task is queued instead.
func (p *Processor) dispatchTaskToAgent(ctx context.Context, taskID, agentID, title, description string) {
	// Check if agent is busy
	activeCount, err := p.store.CountActiveTasksByAgent(ctx, agentID)
	if err != nil {
		log.Printf("[QueueProcessor] Error checking active tasks for agent %s: %v", agentID, err)
		return
	}

	if activeCount > 0 {
		// Agent busy - put task in queue
		log.Printf("[QueueProcessor] Agent %s busy with %d active tasks, queueing task %s", agentID, activeCount, taskID)
		if err := p.store.UpdateTaskStatus(ctx, taskID, "queued"); err != nil {
			log.Printf("[QueueProcessor] Error queueing task %s: %v", taskID, err)
		}
		return
	}

	// Agent free - notify directly
	log.Printf("[QueueProcessor] Notifying agent %s about task %s (%s)", agentID, taskID, title)

	p.agentSender.NotifyAgentAsync(agentID, taskID, title, description, func(tID, aID, reply string, err error) {
		if err != nil {
			log.Printf("[QueueProcessor] Failed to notify agent %s for task %s: %v", agentID, taskID, err)
			// Put back in queue on failure
			p.store.UpdateTaskStatus(ctx, taskID, "queued")
		} else {
			log.Printf("[QueueProcessor] Agent %s notified for task %s", agentID, taskID)
			// Update status to 'backlog' since it's now being worked on
			p.store.UpdateTaskStatus(ctx, taskID, "backlog")

			// Broadcast to websocket
			if p.hub != nil {
				p.hub.BroadcastTaskStatus(taskID, "backlog", 0)
			}
		}
	})
}

func (p *Processor) ProcessOnce(ctx context.Context) {
	p.ProcessScheduledTasks(ctx)

	log.Println("[QueueProcessor] Starting periodic queue check...")

	agents, err := p.store.ListAgents(ctx)
	if err != nil {
		log.Printf("[QueueProcessor] Error listing agents: %v", err)
		return
	}

	processed := 0
	for _, agent := range agents {
		activeCount, err := p.store.CountActiveTasksByAgent(ctx, agent.ID)
		if err != nil {
			log.Printf("[QueueProcessor] Error checking active tasks for agent %s: %v", agent.ID, err)
			continue
		}

		if activeCount > 0 {
			continue
		}

		queued, err := p.store.ListQueuedTasksByAgent(ctx, agent.ID)
		if err != nil {
			log.Printf("[QueueProcessor] Error checking queue for agent %s: %v", agent.ID, err)
			continue
		}

		if len(queued) == 0 {
			continue
		}

		log.Printf("[QueueProcessor] Agent %s is free with %d queued tasks — dispatching next", agent.ID, len(queued))
		p.handler.ProcessAgentQueue(ctx, agent.ID)
		processed++
	}

	log.Printf("[QueueProcessor] Periodic check complete: processed %d agents with queued tasks", processed)
}

func (p *Processor) Start(ctx context.Context, interval time.Duration) {
	if p.running {
		log.Println("[QueueProcessor] Already running")
		return
	}

	p.running = true
	log.Printf("[QueueProcessor] Starting periodic queue processor every %v", interval)

	go func() {
		// Run immediately on startup to catch any overdue scheduled tasks
		p.ProcessOnce(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.ProcessOnce(ctx)
			case <-p.stopChan:
				log.Println("[QueueProcessor] Stopping periodic queue processor")
				p.running = false
				return
			case <-ctx.Done():
				log.Println("[QueueProcessor] Context cancelled, stopping queue processor")
				p.running = false
				return
			}
		}
	}()
}

func (p *Processor) Stop() {
	if !p.running {
		return
	}
	close(p.stopChan)
	p.running = false
}
