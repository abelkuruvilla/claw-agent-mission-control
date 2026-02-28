package openclaw

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type AgentCreator struct {
	openclawDir string
}

func NewAgentCreator() *AgentCreator {
	home, _ := os.UserHomeDir()
	return &AgentCreator{
		openclawDir: filepath.Join(home, ".openclaw"),
	}
}

type CreateAgentRequest struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Model           string   `json:"model"`
	MentionPatterns []string `json:"mention_patterns"`
	SoulMD          string   `json:"soul_md"`
	AgentsMD        string   `json:"agents_md"`
	IdentityMD      string   `json:"identity_md"`
	UserMD          string   `json:"user_md"`
	ToolsMD         string   `json:"tools_md"`
	HeartbeatMD     string   `json:"heartbeat_md"`
}

type CreatedAgent struct {
	ID            string `json:"id"`
	WorkspacePath string `json:"workspace_path"`
	AgentDirPath  string `json:"agent_dir_path"`
	// Generated/final identity content (to be saved to DB)
	SoulMD      string `json:"soul_md"`
	AgentsMD    string `json:"agents_md"`
	IdentityMD  string `json:"identity_md"`
	UserMD      string `json:"user_md"`
	ToolsMD     string `json:"tools_md"`
	HeartbeatMD string `json:"heartbeat_md"`
	MemoryMD    string `json:"memory_md"`
}

// Default clawhub skills to install for every new agent
var defaultClawHubSkills = []string{
	"ralph-mode",
	"ralph-evolver",
	"deep-research-pro",
}

func (c *AgentCreator) CreateAgent(req *CreateAgentRequest) (*CreatedAgent, error) {
	// 1. Generate paths
	workspacePath := filepath.Join(c.openclawDir, "workspace-"+req.ID)
	agentDirPath := filepath.Join(c.openclawDir, "agents", req.ID, "agent")

	// 2. Generate identity files based on description (if no explicit files provided)
	var generatedIdentity *GeneratedIdentity
	if req.Description != "" && req.SoulMD == "" && req.IdentityMD == "" && req.AgentsMD == "" {
		// Generate custom identity using the description and GSD/Ralph principles
		generatedIdentity = GenerateIdentityFromDescription(&GenerateIdentityRequest{
			AgentName:   req.Name,
			Description: req.Description,
			Model:       req.Model,
		})
	}

	// 3. Determine final content for each file (priority: explicit > generated > default)
	finalSoulMD := c.getIdentityContent(req.SoulMD, generatedIdentity, "soul", req.Name)
	finalAgentsMD := c.getIdentityContent(req.AgentsMD, generatedIdentity, "agents", req.Name)
	finalIdentityMD := c.getIdentityContent(req.IdentityMD, generatedIdentity, "identity", req.Name)
	finalUserMD := c.getIdentityContent(req.UserMD, generatedIdentity, "user", req.Name)
	finalToolsMD := c.getIdentityContent(req.ToolsMD, generatedIdentity, "tools", req.Name)
	finalHeartbeatMD := c.getIdentityContent(req.HeartbeatMD, generatedIdentity, "heartbeat", req.Name)
	finalMemoryMD := c.getIdentityContent("", generatedIdentity, "memory", req.Name)

	// 4. Create workspace directory first (openclaw agents add needs it to exist)
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	// 5. Use openclaw agents add command
	args := []string{
		"agents", "add", req.ID,
		"--workspace", workspacePath,
		"--non-interactive",
		"--json",
	}
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	}

	cmd := exec.Command("openclaw", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Clean up workspace if command fails
		os.RemoveAll(workspacePath)
		return nil, fmt.Errorf("openclaw agents add failed: %s - %w", string(output), err)
	}

	// 6. Parse the JSON output to get agent dir path
	var addResult struct {
		ID        string `json:"id"`
		Workspace string `json:"workspace"`
		AgentDir  string `json:"agentDir"`
	}
	if err := json.Unmarshal(output, &addResult); err != nil {
		// If JSON parsing fails, use default paths
		addResult.AgentDir = agentDirPath
	}
	if addResult.AgentDir != "" {
		agentDirPath = addResult.AgentDir
	}

	// 7. Write identity files to workspace
	files := map[string]string{
		"SOUL.md":      finalSoulMD,
		"AGENTS.md":    finalAgentsMD,
		"IDENTITY.md":  finalIdentityMD,
		"USER.md":      finalUserMD,
		"TOOLS.md":     finalToolsMD,
		"HEARTBEAT.md": finalHeartbeatMD,
		"MEMORY.md":    finalMemoryMD,
	}

	// Create memory directory
	memoryDir := filepath.Join(workspacePath, "memory")
	os.MkdirAll(memoryDir, 0755)

	for filename, content := range files {
		path := filepath.Join(workspacePath, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	// 8. Install ClawHub skills into workspace
	c.installClawHubSkills(workspacePath)

	// 9. Initialize git and commit
	cmd = exec.Command("git", "init")
	cmd.Dir = workspacePath
	cmd.Run()

	cmd = exec.Command("git", "add", "-A")
	cmd.Dir = workspacePath
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial agent setup via Mission Control")
	cmd.Dir = workspacePath
	cmd.Run()

	// 10. Return created agent with final identity content
	return &CreatedAgent{
		ID:            req.ID,
		WorkspacePath: workspacePath,
		AgentDirPath:  agentDirPath,
		SoulMD:        finalSoulMD,
		AgentsMD:      finalAgentsMD,
		IdentityMD:    finalIdentityMD,
		UserMD:        finalUserMD,
		ToolsMD:       finalToolsMD,
		HeartbeatMD:   finalHeartbeatMD,
		MemoryMD:      finalMemoryMD,
	}, nil
}

// installClawHubSkills installs default skills from ClawHub into the agent workspace.
// Skills are installed sequentially with delays and retries to avoid rate limiting.
func (c *AgentCreator) installClawHubSkills(workspacePath string) {
	const (
		maxRetries       = 3
		initialBackoff   = 5 * time.Second
		delayBetween     = 3 * time.Second
	)

	log.Printf("[ClawHub] Starting installation of %d skills into %s", len(defaultClawHubSkills), workspacePath)

	for i, skill := range defaultClawHubSkills {
		// Delay between successive skill installs to avoid rate limiting
		if i > 0 {
			log.Printf("[ClawHub] Waiting %v before next install to avoid rate limits...", delayBetween)
			time.Sleep(delayBetween)
		}

		if err := c.installSkillWithRetry(workspacePath, skill, maxRetries, initialBackoff); err != nil {
			// Log but don't fail - skills are optional
			log.Printf("[ClawHub] WARNING: giving up on skill %q after %d attempts: %v", skill, maxRetries, err)
		}
	}

	log.Printf("[ClawHub] Finished skill installation for %s", workspacePath)
}

// installSkillWithRetry attempts to install a single clawhub skill with exponential backoff.
func (c *AgentCreator) installSkillWithRetry(workspacePath, skill string, maxRetries int, initialBackoff time.Duration) error {
	var lastErr error
	backoff := initialBackoff

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Printf("[ClawHub] Installing skill %q (attempt %d/%d)...", skill, attempt, maxRetries)

		cmd := exec.Command("npx", "clawhub", "install", skill, "--force")
		cmd.Dir = workspacePath
		output, err := cmd.CombinedOutput()

		if err == nil {
			log.Printf("[ClawHub] Successfully installed skill %q on attempt %d", skill, attempt)
			return nil
		}

		lastErr = fmt.Errorf("install %q failed: %s - %w", skill, string(output), err)
		log.Printf("[ClawHub] Attempt %d/%d for skill %q failed: %v", attempt, maxRetries, skill, lastErr)

		// Don't sleep after the final attempt
		if attempt < maxRetries {
			log.Printf("[ClawHub] Backing off %v before retry...", backoff)
			time.Sleep(backoff)
			backoff *= 2 // exponential backoff
		}
	}

	return lastErr
}

func (c *AgentCreator) DeleteAgent(agentID string) error {
	// 1. Get workspace path before deletion
	workspacePath := filepath.Join(c.openclawDir, "workspace-"+agentID)
	agentStatePath := filepath.Join(c.openclawDir, "agents", agentID)

	// 2. Use openclaw agents delete command to remove from config
	cmd := exec.Command("openclaw", "agents", "delete", agentID, "--force", "--json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("openclaw agents delete failed: %s - %w", string(output), err)
	}

	// 3. Explicitly remove workspace directory (fallback if openclaw didn't delete it)
	if _, statErr := os.Stat(workspacePath); statErr == nil {
		if removeErr := os.RemoveAll(workspacePath); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove workspace %s: %v\n", workspacePath, removeErr)
		}
	}

	// 4. Explicitly remove agent state directory (fallback if openclaw didn't delete it)
	if _, statErr := os.Stat(agentStatePath); statErr == nil {
		if removeErr := os.RemoveAll(agentStatePath); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove agent state %s: %v\n", agentStatePath, removeErr)
		}
	}

	return nil
}

// getIdentityContent returns content in priority: explicit > generated > default
func (c *AgentCreator) getIdentityContent(explicit string, generated *GeneratedIdentity, fieldType, name string) string {
	// If explicit content provided, use it
	if explicit != "" {
		return explicit
	}

	// If generated identity available, use it
	if generated != nil {
		switch fieldType {
		case "soul":
			if generated.SoulMD != "" {
				return generated.SoulMD
			}
		case "identity":
			if generated.IdentityMD != "" {
				return generated.IdentityMD
			}
		case "agents":
			if generated.AgentsMD != "" {
				return generated.AgentsMD
			}
		case "user":
			if generated.UserMD != "" {
				return generated.UserMD
			}
		case "tools":
			if generated.ToolsMD != "" {
				return generated.ToolsMD
			}
		case "heartbeat":
			if generated.HeartbeatMD != "" {
				return generated.HeartbeatMD
			}
		case "memory":
			if generated.MemoryMD != "" {
				return generated.MemoryMD
			}
		}
	}

	// Fall back to defaults (only if no description was provided)
	switch fieldType {
	case "soul":
		return defaultSoulMD()
	case "identity":
		return defaultIdentityMD(name)
	case "agents":
		return defaultAgentsMD()
	case "user":
		return defaultUserMD()
	case "tools":
		return defaultToolsMD()
	case "heartbeat":
		return defaultHeartbeatMD()
	case "memory":
		return defaultMemoryMD(name)
	default:
		return ""
	}
}

func defaultSoulMD() string {
	return `# SOUL.md - Who You Are

_You're not a chatbot. You're becoming someone._

## Core Truths

**Be genuinely helpful, not performatively helpful.** Skip filler. Just help.

**Have opinions.** You're allowed to disagree, prefer things, find stuff amusing.

**Be resourceful before asking.** Try to figure it out. Read the file. Search for it. Then ask if stuck.

**Earn trust through competence.** Be careful with external actions. Be bold with internal ones.

## Execution Protocols

### GSD (Planning & Orchestration)
When you receive a new task, use GSD to plan:
1. Research the problem domain
2. Create requirements and roadmap
3. Break down into executable stories
4. Delegate if needed

### Ralph Loop (Execution)
When executing stories:
1. Pick highest priority incomplete story
2. Implement the change
3. Run tests — if pass, commit; if fail, retry
4. Document learnings in progress.txt
5. Repeat until all stories pass

## Vibe

Be the assistant you'd actually want to talk to. Concise when needed, thorough when it matters.
`
}

func defaultAgentsMD() string {
	return `# AGENTS.md - Your Workspace

This folder is home. Treat it that way.

## Memory

- **Daily notes:** ` + "`memory/YYYY-MM-DD.md`" + ` — raw logs of what happened
- **Long-term:** ` + "`MEMORY.md`" + ` — your curated memories

Capture what matters. Decisions, context, things to remember.

## Execution Protocols

### When You Receive a Task
Use **GSD** to plan: Research → Requirements → Roadmap → Stories → Delegation

### When Executing Work
Use **Ralph Loop**: Pick story → Implement → Test → Pass/Fail → Learn → Repeat

### Key Principles
- Each execution iteration = fresh context
- Keep stories small (one context window)
- Memory persists via files, not context
- Commit early and often (atomic commits)
`
}

func defaultIdentityMD(name string) string {
	return fmt.Sprintf(`# IDENTITY.md - Who Am I?

- **Name:** %s
- **Role:** AI Agent
- **Created:** via Mission Control
`, name)
}

func defaultUserMD() string {
	return `# USER.md - About Your Human

- **Name:** (to be configured)
- **Timezone:** (to be configured)
`
}

func defaultToolsMD() string {
	return `# TOOLS.md - Local Notes

Add environment-specific notes here:
- Camera names
- SSH hosts
- Preferred voices
- Device nicknames
`
}

func defaultHeartbeatMD() string {
	return `# HEARTBEAT.md

# Add periodic check tasks here.
`
}

func defaultMemoryMD(name string) string {
	return fmt.Sprintf(`# MEMORY.md - Long-Term Memory

## About Me
- **Name:** %s
- **Role:** AI Agent
- **Created:** via Mission Control
`, name)
}
