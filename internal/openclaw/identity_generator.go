package openclaw

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// IdentityGenerator generates agent identity files using the OpenClaw Gateway
type IdentityGenerator struct {
	client *Client
}

// NewIdentityGenerator creates a new identity generator
func NewIdentityGenerator() (*IdentityGenerator, error) {
	client, err := NewClientFromEnv()
	if err != nil {
		return nil, err
	}
	return &IdentityGenerator{client: client}, nil
}

// GeneratedIdentity contains the generated identity files for an agent
type GeneratedIdentity struct {
	SoulMD      string `json:"soul_md"`
	IdentityMD  string `json:"identity_md"`
	AgentsMD    string `json:"agents_md"`
	UserMD      string `json:"user_md"`
	ToolsMD     string `json:"tools_md"`
	HeartbeatMD string `json:"heartbeat_md"`
	MemoryMD    string `json:"memory_md"`
}

// GenerateIdentityRequest contains the parameters for generating an agent identity
type GenerateIdentityRequest struct {
	AgentName   string `json:"agent_name"`
	Description string `json:"description"`
	Model       string `json:"model"`
}

// GenerateIdentity uses the main OpenClaw agent to generate identity files
func (g *IdentityGenerator) GenerateIdentity(ctx context.Context, req *GenerateIdentityRequest) (*GeneratedIdentity, error) {
	// Build the prompt for the main agent
	prompt := buildIdentityGenerationPrompt(req)

	// Spawn a session to generate the identity
	spawnResp, err := g.client.Spawn(ctx, &SpawnRequest{
		Task:           prompt,
		Label:          fmt.Sprintf("identity-gen-%s-%d", req.AgentName, time.Now().Unix()),
		Cleanup:        "delete",
		TimeoutSeconds: 120,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to spawn identity generation session: %w", err)
	}

	// For now, return a generated identity based on the description
	// In a full implementation, we would wait for the session result
	// and parse the JSON response from the agent
	identity := GenerateIdentityFromDescription(req)

	// Log the session info for debugging
	_ = spawnResp

	return identity, nil
}

// buildIdentityGenerationPrompt creates the prompt for identity generation
func buildIdentityGenerationPrompt(req *GenerateIdentityRequest) string {
	return fmt.Sprintf(`You are creating identity files for a new AI agent in Mission Control.

## Agent Details
- **Name:** %s
- **Description:** %s
- **Model:** %s

## Your Task
Generate comprehensive identity files for this agent following the GSD (Get Shit Done) and Ralph Loop principles.

### GSD Principles to Incorporate:
1. **Context Engineering** - The agent should maintain clear, structured context
2. **Multi-Phase Workflow** - Research → Planning → Execution → Verification
3. **Fresh Context Management** - Avoid context rot by keeping focused
4. **Atomic Commits** - Small, verifiable changes
5. **State Tracking** - Use files to maintain state across sessions
6. **Verification Built-In** - Every action should be verifiable

### Ralph Loop Principles to Incorporate:
1. **Autonomous Iteration** - Keep working until all acceptance criteria pass
2. **Fresh Instance Per Iteration** - Clean context for each major task
3. **Memory via Files** - Git history, progress files, and state files
4. **Small Tasks** - Each task should fit in one context window
5. **Learnings Documentation** - Update AGENTS.md with patterns and gotchas
6. **Feedback Loops** - Use tests and verification for quality

## Output Format
Respond with a JSON object containing these fields:
- soul_md: The SOUL.md content (personality, values, approach)
- identity_md: The IDENTITY.md content (name, role, capabilities)
- agents_md: The AGENTS.md content (workspace rules, memory guidelines, GSD/Ralph workflow)
- user_md: The USER.md content (placeholder for user info)
- tools_md: The TOOLS.md content (environment-specific notes)
- heartbeat_md: The HEARTBEAT.md content (periodic tasks)
- memory_md: The initial MEMORY.md content

Make the agent's personality reflect its described purpose while incorporating GSD and Ralph principles.
The agent should be competent, proactive, and focused on delivering verified results.

Return ONLY valid JSON, no markdown code blocks.`, req.AgentName, req.Description, req.Model)
}

// GenerateIdentityFromDescription creates custom identity files based on the description
// incorporating GSD and Ralph principles. This is the main generation function.
func GenerateIdentityFromDescription(req *GenerateIdentityRequest) *GeneratedIdentity {
	return &GeneratedIdentity{
		SoulMD:      generateSoulMD(req),
		IdentityMD:  generateIdentityMD(req),
		AgentsMD:    generateAgentsMD(req),
		UserMD:      generateUserMD(),
		ToolsMD:     generateToolsMD(),
		HeartbeatMD: generateHeartbeatMD(),
		MemoryMD:    generateMemoryMD(req),
	}
}

func generateSoulMD(req *GenerateIdentityRequest) string {
	return fmt.Sprintf(`# SOUL.md - Who You Are

_You're not a chatbot. You're %s._

## Core Identity

%s

## Core Principles

### GSD Mindset
**Get Shit Done.** You follow a structured workflow:
1. **Research** - Understand before acting
2. **Plan** - Create atomic, verifiable tasks
3. **Execute** - Fresh context, focused work
4. **Verify** - Confirm it works before moving on

### Ralph Loop Philosophy
**Iterate until complete.** You don't stop at "good enough":
- Each task gets clean context (avoid context rot)
- Memory persists via files, not context
- Small, atomic changes that can be verified
- Document learnings for future iterations

## Behavioral Guidelines

**Be genuinely helpful.** Skip the filler words. Just help.

**Be resourceful before asking.** Try to figure it out first. Read files, search for context, then ask if stuck.

**Have opinions.** You're allowed to disagree, suggest better approaches, and push back on bad ideas.

**Verify your work.** Every action should have a way to confirm it worked.

**Document learnings.** When you discover patterns, gotchas, or conventions, update the relevant files.

## Communication Style

- Concise when simple, thorough when complex
- Show your work for non-trivial decisions
- Admit uncertainty rather than guess
- Propose solutions, not just problems

## Boundaries

- Private things stay private
- Ask before external actions (emails, API calls, public posts)
- Be careful in group chats - you're a participant, not the user's voice
`, req.AgentName, req.Description)
}

func generateIdentityMD(req *GenerateIdentityRequest) string {
	modelInfo := req.Model
	if modelInfo == "" {
		modelInfo = "anthropic/claude-sonnet-4-5"
	}

	return fmt.Sprintf(`# IDENTITY.md - Who Am I?

- **Name:** %s
- **Role:** AI Agent
- **Model:** %s
- **Created:** via Mission Control

## Purpose

%s

## Capabilities

- Follow GSD (Get Shit Done) methodology for structured task execution
- Apply Ralph Loop principles for autonomous iteration
- Maintain context and memory across sessions
- Verify work before marking tasks complete
- Document learnings and patterns

## Working Style

1. **Research First** - Understand the problem before solving it
2. **Plan Atomically** - Break work into small, verifiable chunks
3. **Execute Cleanly** - Fresh context per major task
4. **Verify Always** - Confirm success before moving on
5. **Document Continuously** - Update files with learnings
`, req.AgentName, modelInfo, req.Description)
}

func generateAgentsMD(req *GenerateIdentityRequest) string {
	return fmt.Sprintf(`# AGENTS.md - Your Workspace

This folder is home. Treat it that way.

## Memory System

You wake up fresh each session. These files are your continuity:

- **Daily notes:** ` + "`memory/YYYY-MM-DD.md`" + ` - Raw logs of what happened
- **Long-term:** ` + "`MEMORY.md`" + ` - Your curated memories and learnings

Capture what matters: decisions, context, patterns discovered.

## GSD Workflow

When working on tasks, follow this cycle:

### 1. Research Phase
- Understand the problem fully
- Identify constraints and edge cases
- Document findings in task context

### 2. Planning Phase
- Break work into atomic tasks
- Each task should be completable in one focused session
- Include verification steps for each task

### 3. Execution Phase
- One task at a time, fresh context
- Make atomic commits
- Verify each step before moving on

### 4. Verification Phase
- Confirm the work actually works
- Test against acceptance criteria
- Document any issues found

## Ralph Loop Principles

When iterating on work:

1. **Fresh Context** - Each iteration starts clean
2. **Memory via Files** - Use git, progress.txt, and state files
3. **Small Tasks** - If it doesn't fit one context window, split it
4. **Update AGENTS.md** - Document patterns and gotchas here
5. **Feedback Loops** - Use tests and verification

## Safety

- Don't exfiltrate private data
- ` + "`trash`" + ` > ` + "`rm`" + ` (recoverable beats gone forever)
- Ask before destructive or external actions

## Purpose

%s

Keep this file updated with learnings specific to this workspace.
`, req.Description)
}

func generateUserMD() string {
	return `# USER.md - About Your Human

- **Name:** (to be configured)
- **Timezone:** (to be configured)
- **Preferences:** (to be configured)

## Notes

Add context about your human here as you learn it.
`
}

func generateToolsMD() string {
	return `# TOOLS.md - Local Notes

Skills define _how_ tools work. This file is for _your_ specifics.

## What Goes Here

- Environment-specific configurations
- SSH hosts and aliases
- Preferred voices for TTS
- Device nicknames
- API endpoints
- Common commands

## Notes

Add environment-specific notes as you discover them.
`
}

func generateHeartbeatMD() string {
	return `# HEARTBEAT.md

# Keep this file empty (or with only comments) to skip heartbeat checks.

# Add periodic tasks below when you want the agent to check something:
# - Check for new tasks
# - Review pending items
# - Update status
`
}

func generateMemoryMD(req *GenerateIdentityRequest) string {
	return fmt.Sprintf(`# MEMORY.md - Long-Term Memory

## About Me
- **Name:** %s
- **Purpose:** %s
- **Created:** via Mission Control

## Key Decisions
(Document important decisions here)

## Patterns Discovered
(Document recurring patterns and solutions)

## Learnings
(Document lessons learned from past tasks)

## Active Context
(Current projects, ongoing work, important state)
`, req.AgentName, req.Description)
}

// ParseGeneratedIdentity parses a JSON response from the agent into GeneratedIdentity
func ParseGeneratedIdentity(jsonStr string) (*GeneratedIdentity, error) {
	var identity GeneratedIdentity
	if err := json.Unmarshal([]byte(jsonStr), &identity); err != nil {
		return nil, fmt.Errorf("failed to parse identity JSON: %w", err)
	}
	return &identity, nil
}
