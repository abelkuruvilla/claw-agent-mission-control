package openclaw

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigReader reads OpenClaw configuration and agent workspace files
type ConfigReader struct {
	configPath string
}

// AgentConfig represents an agent from OpenClaw config with workspace files
type AgentConfig struct {
	ID              string
	Name            string
	Description     string
	WorkspacePath   string
	AgentDirPath    string
	Model           string
	MentionPatterns []string
	
	// Workspace files
	SoulMD      string
	AgentsMD    string
	IdentityMD  string
	UserMD      string
	ToolsMD     string
	HeartbeatMD string
	MemoryMD    string
}

// NewConfigReader creates a new config reader
func NewConfigReader(configPath string) *ConfigReader {
	// Default to ~/.openclaw/openclaw.json if empty
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			configPath = filepath.Join(os.Getenv("HOME"), ".openclaw", "openclaw.json")
		} else {
			configPath = filepath.Join(home, ".openclaw", "openclaw.json")
		}
	}
	
	return &ConfigReader{
		configPath: configPath,
	}
}

// ReadAgents reads all agents from OpenClaw configuration
func (r *ConfigReader) ReadAgents() ([]AgentConfig, error) {
	// Read config file
	data, err := os.ReadFile(r.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", r.configPath, err)
	}
	
	// Parse JSON
	var config struct {
		Agents struct {
			List []struct {
				ID        string   `json:"id"`
				Name      string   `json:"name"`
				Workspace string   `json:"workspace"`
				AgentDir  string   `json:"agentDir"`
				Model     string   `json:"model"`
				GroupChat *struct {
					MentionPatterns []string `json:"mentionPatterns"`
				} `json:"groupChat"`
			} `json:"list"`
		} `json:"agents"`
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	
	// Process each agent
	var agents []AgentConfig
	for _, agentEntry := range config.Agents.List {
		// Skip agents without workspace (like "main")
		if agentEntry.Workspace == "" {
			continue
		}
		
		agent := AgentConfig{
			ID:            agentEntry.ID,
			Name:          agentEntry.Name,
			WorkspacePath: agentEntry.Workspace,
			AgentDirPath:  agentEntry.AgentDir,
			Model:         agentEntry.Model,
		}
		
		// Extract mention patterns
		if agentEntry.GroupChat != nil {
			agent.MentionPatterns = agentEntry.GroupChat.MentionPatterns
		}
		
		// Read workspace files
		if err := r.readWorkspaceFiles(&agent); err != nil {
			// Log warning but don't fail - some files might be missing
			fmt.Printf("Warning: failed to read some workspace files for agent %s: %v\n", agent.ID, err)
		}
		
		agents = append(agents, agent)
	}
	
	return agents, nil
}

// readWorkspaceFiles reads all workspace markdown files for an agent
func (r *ConfigReader) readWorkspaceFiles(agent *AgentConfig) error {
	if agent.WorkspacePath == "" {
		return nil
	}
	
	// List of files to read
	files := map[string]*string{
		"SOUL.md":      &agent.SoulMD,
		"AGENTS.md":    &agent.AgentsMD,
		"IDENTITY.md":  &agent.IdentityMD,
		"USER.md":      &agent.UserMD,
		"TOOLS.md":     &agent.ToolsMD,
		"HEARTBEAT.md": &agent.HeartbeatMD,
		"MEMORY.md":    &agent.MemoryMD,
	}
	
	var errors []string
	for filename, target := range files {
		filePath := filepath.Join(agent.WorkspacePath, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			if !os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("%s: %v", filename, err))
			}
			// File doesn't exist - leave empty
			continue
		}
		*target = string(content)
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("errors reading files: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// GetConfigPath returns the config file path being used
func (r *ConfigReader) GetConfigPath() string {
	return r.configPath
}

// ModelConfig represents a model from OpenClaw config
type ModelConfig struct {
	ID    string `json:"id"`
	Alias string `json:"alias,omitempty"`
}

// ReadModels reads all configured models from OpenClaw configuration
func (r *ConfigReader) ReadModels() ([]ModelConfig, error) {
	// Read config file
	data, err := os.ReadFile(r.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", r.configPath, err)
	}

	// Parse JSON - look for agent.defaults.models path
	var config struct {
		Agents struct {
			Defaults struct {
				Models map[string]struct {
					Alias string `json:"alias"`
				} `json:"models"`
			} `json:"defaults"`
		} `json:"agents"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Pull models from agent.defaults.models
	var models []ModelConfig
	for modelID, modelConfig := range config.Agents.Defaults.Models {
		models = append(models, ModelConfig{
			ID:    modelID,
			Alias: modelConfig.Alias,
		})
	}

	return models, nil
}
