package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/openclaw"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
)

// SyncService handles syncing agents from OpenClaw config to the database
type SyncService struct {
	store        *store.Store
	configReader *openclaw.ConfigReader
	stopChan     chan struct{}
	running      bool
}

// NewSyncService creates a new sync service
func NewSyncService(st *store.Store, configReader *openclaw.ConfigReader) *SyncService {
	return &SyncService{
		store:        st,
		configReader: configReader,
		stopChan:     make(chan struct{}),
	}
}

// SyncOnce performs a one-time sync of agents from OpenClaw config to database
func (s *SyncService) SyncOnce(ctx context.Context) error {
	log.Println("Starting agent sync from OpenClaw config...")
	
	// Read agents from OpenClaw config
	agents, err := s.configReader.ReadAgents()
	if err != nil {
		return fmt.Errorf("failed to read agents from config: %w", err)
	}
	
	log.Printf("Found %d agents in OpenClaw config", len(agents))
	
	// Get existing agents from database
	existingAgents, err := s.store.ListAgents(ctx)
	if err != nil {
		return fmt.Errorf("failed to list existing agents: %w", err)
	}
	
	// Create a map of existing agents by ID
	existingMap := make(map[string]db.Agent)
	for _, agent := range existingAgents {
		existingMap[agent.ID] = agent
	}
	
	// Track sync results
	var added, updated, unchanged int
	
	// Process each agent from config
	for _, agentConfig := range agents {
		existing, exists := existingMap[agentConfig.ID]
		
		if !exists {
			// Create new agent
			if err := s.createAgent(ctx, agentConfig); err != nil {
				log.Printf("Error creating agent %s: %v", agentConfig.ID, err)
				continue
			}
			added++
			log.Printf("✓ Added agent: %s (%s)", agentConfig.ID, agentConfig.Name)
		} else {
			// Check if agent needs update
			if s.needsUpdate(existing, agentConfig) {
				if err := s.updateAgent(ctx, agentConfig); err != nil {
					log.Printf("Error updating agent %s: %v", agentConfig.ID, err)
					continue
				}
				updated++
				log.Printf("✓ Updated agent: %s (%s)", agentConfig.ID, agentConfig.Name)
			} else {
				unchanged++
			}
		}
		
		// Remove from existing map (so we can track orphans)
		delete(existingMap, agentConfig.ID)
	}
	
	// Mark orphaned agents (exist in DB but not in config)
	for _, orphan := range existingMap {
		log.Printf("⚠ Agent %s exists in DB but not in OpenClaw config (orphaned)", orphan.ID)
		// Optionally mark as orphaned or delete
		// For now, we just log it
	}
	
	log.Printf("Sync complete: %d added, %d updated, %d unchanged", added, updated, unchanged)
	
	return nil
}

// createAgent creates a new agent in the database
func (s *SyncService) createAgent(ctx context.Context, agentConfig openclaw.AgentConfig) error {
	mentionPatternsJSON := ""
	if len(agentConfig.MentionPatterns) > 0 {
		data, _ := json.Marshal(agentConfig.MentionPatterns)
		mentionPatternsJSON = string(data)
	}
	
	_, err := s.store.CreateAgent(ctx, db.CreateAgentParams{
		ID:              agentConfig.ID,
		Name:            agentConfig.Name,
		Description:     toNullString(agentConfig.Description),
		Status:          toNullString("active"),
		WorkspacePath:   toNullString(agentConfig.WorkspacePath),
		AgentDirPath:    toNullString(agentConfig.AgentDirPath),
		Model:           toNullString(agentConfig.Model),
		MentionPatterns: toNullString(mentionPatternsJSON),
		SoulMd:          toNullString(agentConfig.SoulMD),
		AgentsMd:        toNullString(agentConfig.AgentsMD),
		IdentityMd:      toNullString(agentConfig.IdentityMD),
		UserMd:          toNullString(agentConfig.UserMD),
		ToolsMd:         toNullString(agentConfig.ToolsMD),
		HeartbeatMd:     toNullString(agentConfig.HeartbeatMD),
		MemoryMd:        toNullString(agentConfig.MemoryMD),
	})
	
	return err
}

// updateAgent updates an existing agent in the database
func (s *SyncService) updateAgent(ctx context.Context, agentConfig openclaw.AgentConfig) error {
	mentionPatternsJSON := ""
	if len(agentConfig.MentionPatterns) > 0 {
		data, _ := json.Marshal(agentConfig.MentionPatterns)
		mentionPatternsJSON = string(data)
	}
	
	// Get existing agent to preserve fields not in UpdateAgentParams
	existing, err := s.store.GetAgent(ctx, agentConfig.ID)
	if err != nil {
		return fmt.Errorf("failed to get existing agent: %w", err)
	}
	
	// UpdateAgentParams doesn't include workspace_path, agent_dir_path, or memory_md
	// These are set during creation and shouldn't change during sync
	// We update the fields that can change via UpdateAgent
	_, err = s.store.UpdateAgent(ctx, db.UpdateAgentParams{
		ID:               agentConfig.ID,
		Name:             agentConfig.Name,
		Description:      toNullString(agentConfig.Description),
		Status:           existing.Status, // Preserve existing status
		Model:            toNullString(agentConfig.Model),
		MentionPatterns:  toNullString(mentionPatternsJSON),
		SoulMd:           toNullString(agentConfig.SoulMD),
		AgentsMd:         toNullString(agentConfig.AgentsMD),
		IdentityMd:       toNullString(agentConfig.IdentityMD),
		UserMd:           toNullString(agentConfig.UserMD),
		ToolsMd:          toNullString(agentConfig.ToolsMD),
		HeartbeatMd:      toNullString(agentConfig.HeartbeatMD),
		ActiveSessionKey: existing.ActiveSessionKey, // Preserve existing session
		CurrentTaskID:    existing.CurrentTaskID,    // Preserve existing task
	})
	
	// Note: memory_md is not updated via UpdateAgent - it's only set during creation
	// This is intentional to avoid overwriting runtime memory
	
	return err
}

// needsUpdate checks if an agent needs to be updated
func (s *SyncService) needsUpdate(existing db.Agent, config openclaw.AgentConfig) bool {
	// Compare key fields
	if existing.Name != config.Name {
		return true
	}
	
	if existing.Model.String != config.Model {
		return true
	}
	
	if existing.WorkspacePath.String != config.WorkspacePath {
		return true
	}
	
	// Compare workspace files (check if content changed)
	if existing.SoulMd.String != config.SoulMD {
		return true
	}
	
	if existing.AgentsMd.String != config.AgentsMD {
		return true
	}
	
	if existing.IdentityMd.String != config.IdentityMD {
		return true
	}
	
	if existing.UserMd.String != config.UserMD {
		return true
	}
	
	if existing.ToolsMd.String != config.ToolsMD {
		return true
	}
	
	if existing.HeartbeatMd.String != config.HeartbeatMD {
		return true
	}
	
	if existing.MemoryMd.String != config.MemoryMD {
		return true
	}
	
	return false
}

// StartPeriodicSync starts periodic syncing in the background
func (s *SyncService) StartPeriodicSync(ctx context.Context, interval time.Duration) {
	if s.running {
		log.Println("Periodic sync already running")
		return
	}
	
	s.running = true
	log.Printf("Starting periodic sync every %v", interval)
	
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				if err := s.SyncOnce(ctx); err != nil {
					log.Printf("Periodic sync error: %v", err)
				}
			case <-s.stopChan:
				log.Println("Stopping periodic sync")
				s.running = false
				return
			case <-ctx.Done():
				log.Println("Context cancelled, stopping periodic sync")
				s.running = false
				return
			}
		}
	}()
}

// StopPeriodicSync stops the periodic sync
func (s *SyncService) StopPeriodicSync() {
	if !s.running {
		return
	}
	close(s.stopChan)
	s.running = false
}

// Helper function to convert string to sql.NullString
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
