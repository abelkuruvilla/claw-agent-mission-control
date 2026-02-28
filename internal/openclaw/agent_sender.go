package openclaw

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// AgentSendResult holds the structured output from `openclaw agent --json`
type AgentSendResult struct {
	Reply string   `json:"reply"`
	Media []string `json:"media,omitempty"`
}

// AgentSender sends messages directly to OpenClaw agents using the CLI.
// Uses `openclaw agent --agent <id> --message <text>` to push task
// notifications without polling.
type AgentSender struct {
	missionControlURL string
	timeout           time.Duration
}

// AgentSendCallback is called asynchronously when the agent produces a result
// or an error. Implementations should persist the result (e.g. as a comment
// or progress update on the task).
type AgentSendCallback func(taskID, agentID, reply string, err error)

// NewAgentSender creates an AgentSender.
// missionControlURL is the base URL agents can reach the MC API at
// (e.g. "http://localhost:8080/api/v1").
func NewAgentSender(missionControlURL string) *AgentSender {
	timeout := 5 * time.Minute
	return &AgentSender{
		missionControlURL: missionControlURL,
		timeout:           timeout,
	}
}

// buildTaskMessage constructs the message to send to the agent about a new task assignment.
func buildTaskMessage(taskID, title, description, missionControlURL string) string {
	var sb strings.Builder
	sb.WriteString("You have been assigned a new task in Mission Control.\n\n")
	sb.WriteString("## Task Details\n")
	sb.WriteString(fmt.Sprintf("- **Task ID:** %s\n", taskID))
	sb.WriteString(fmt.Sprintf("- **Title:** %s\n", title))
	if description != "" {
		sb.WriteString(fmt.Sprintf("- **Description:** %s\n", description))
	}
	sb.WriteString("\n## API Endpoint\n")
	sb.WriteString("Fetch full task details (including phases and stories) from:\n")
	sb.WriteString(fmt.Sprintf("```\ncurl \"%s/tasks/%s?include=phases,stories\"\n```\n\n", missionControlURL, taskID))
	sb.WriteString("## Instructions\n")
	sb.WriteString("1. Read the full task details from the API above.\n")
	sb.WriteString("2. Follow the GSD protocol to plan the work (Research → Requirements → Roadmap → Stories).\n")
	sb.WriteString("3. Execute each story using the Ralph Loop (Pick → Implement → Test → Pass/Fail → Learn → Repeat).\n")
	sb.WriteString("4. Update your task status and progress via the Mission Control API as you work.\n")
	sb.WriteString(fmt.Sprintf("5. Report progress: `curl -X POST \"%s/tasks/%s/progress-txt\" -H 'Content-Type: application/json' -d '{\"content\": \"[timestamp] your update\"}'`\n", missionControlURL, taskID))
	sb.WriteString(fmt.Sprintf("6. **CRITICAL — When complete, you MUST update status to `done`**: `curl -X PUT \"%s/tasks/%s/status\" -H 'Content-Type: application/json' -d '{\"status\": \"done\"}'`\n", missionControlURL, taskID))
	sb.WriteString("   This triggers an automatic notification to the orchestrator agent who delegated this task. If you do not update the status, the orchestrator will never know you finished.\n")
	return sb.String()
}

// newSessionCommand is the command sent to the agent to start a fresh session
// so that previous task context does not carry over.
const newSessionCommand = "/new"

// NotifyAgentAsync sends a task assignment message to the specified agent
// in a background goroutine. It first sends /new to start a fresh session,
// then sends the task details. When the agent responds to the task message,
// the callback is invoked with the reply text (or error). The caller should NOT block on this.
func (s *AgentSender) NotifyAgentAsync(agentID, taskID, title, description string, callback AgentSendCallback) {
	go func() {
		log.Printf("[AgentSender] Sending task %s notification to agent %s", taskID, agentID)

		// Note: /new is NOT sent here to allow the agent to continue from its previous context.
		// This enables proper retry behavior for failed tasks.

		message := buildTaskMessage(taskID, title, description, s.missionControlURL)

		reply, err := s.sendToAgentWithRetry(agentID, message)
		if err != nil {
			log.Printf("[AgentSender] ERROR sending to agent %s for task %s: %v", agentID, taskID, err)
		} else {
			log.Printf("[AgentSender] Agent %s acknowledged task %s (reply length: %d)", agentID, taskID, len(reply))
		}

		if callback != nil {
			callback(taskID, agentID, reply, err)
		}
	}()
}

// buildSubtaskCompletionMessage constructs the message sent to the orchestrator
// when a subtask reaches a terminal status (done/failed).
func buildSubtaskCompletionMessage(
	subtaskID, subtaskTitle, subtaskStatus,
	parentTaskID, parentTaskTitle,
	specialistAgentID, missionControlURL string,
) string {
	var sb strings.Builder
	sb.WriteString("A subtask you delegated has completed.\n\n")
	sb.WriteString("## Subtask Result\n")
	sb.WriteString(fmt.Sprintf("- **Subtask ID:** %s\n", subtaskID))
	sb.WriteString(fmt.Sprintf("- **Title:** %s\n", subtaskTitle))
	sb.WriteString(fmt.Sprintf("- **Status:** %s\n", subtaskStatus))
	sb.WriteString(fmt.Sprintf("- **Completed by:** %s\n", specialistAgentID))
	sb.WriteString(fmt.Sprintf("- **Parent Task ID:** %s\n", parentTaskID))
	sb.WriteString(fmt.Sprintf("- **Parent Task Title:** %s\n", parentTaskTitle))

	sb.WriteString("\n## Next Steps\n")
	sb.WriteString("1. Read the subtask results and progress:\n")
	sb.WriteString(fmt.Sprintf("```\ncurl \"%s/tasks/%s?include=phases,stories\"\n```\n", missionControlURL, subtaskID))
	sb.WriteString("2. Check remaining subtasks:\n")
	sb.WriteString(fmt.Sprintf("```\ncurl \"%s/tasks/%s/subtasks\"\n```\n", missionControlURL, parentTaskID))
	sb.WriteString("3. Based on the results:\n")
	if subtaskStatus == "done" {
		sb.WriteString("   - If more work is needed, create the next subtask and assign to the appropriate agent.\n")
		sb.WriteString("   - If all subtasks are complete, verify the results and update the parent task to `done`.\n")
	} else {
		sb.WriteString("   - Review the failure, re-scope if needed, and create a new subtask or mark the parent task as `failed`.\n")
	}
	sb.WriteString(fmt.Sprintf("4. Update parent task status: `curl -X PUT \"%s/tasks/%s/status\" -H 'Content-Type: application/json' -d '{\"status\": \"done\"}'`\n", missionControlURL, parentTaskID))
	return sb.String()
}

// NotifySubtaskCompletionAsync sends a message to the orchestrator agent
// (the parent task's assigned agent) informing them that a subtask has completed.
// This enables the orchestrator to continue the delegation chain without polling.
func (s *AgentSender) NotifySubtaskCompletionAsync(
	orchestratorAgentID,
	subtaskID, subtaskTitle, subtaskStatus,
	parentTaskID, parentTaskTitle,
	specialistAgentID string,
	callback AgentSendCallback,
) {
	go func() {
		log.Printf("[AgentSender] Notifying orchestrator %s: subtask %s (%s) completed with status %s",
			orchestratorAgentID, subtaskID, subtaskTitle, subtaskStatus)

		message := buildSubtaskCompletionMessage(
			subtaskID, subtaskTitle, subtaskStatus,
			parentTaskID, parentTaskTitle,
			specialistAgentID, s.missionControlURL,
		)

		reply, err := s.sendToAgentWithRetry(orchestratorAgentID, message)
		if err != nil {
			log.Printf("[AgentSender] ERROR notifying orchestrator %s about subtask %s: %v",
				orchestratorAgentID, subtaskID, err)
		} else {
			log.Printf("[AgentSender] Orchestrator %s acknowledged subtask %s completion (reply length: %d)",
				orchestratorAgentID, subtaskID, len(reply))
		}

		if callback != nil {
			callback(parentTaskID, orchestratorAgentID, reply, err)
		}
	}()
}

// isRetryableError returns true if the error is likely transient
// (session locked, timeout) and the send should be retried.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "session file locked") ||
		strings.Contains(msg, "timed out") ||
		strings.Contains(msg, "All models failed")
}

// sendToAgentWithRetry wraps sendToAgent with exponential backoff retry.
func (s *AgentSender) sendToAgentWithRetry(agentID, message string) (string, error) {
	const maxRetries = 10
	const initialBackoff = 30 * time.Second
	const maxBackoff = 5 * time.Minute

	backoff := initialBackoff
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		reply, err := s.sendToAgent(agentID, message)
		if err == nil {
			if attempt > 1 {
				log.Printf("[AgentSender] Agent %s succeeded on attempt %d", agentID, attempt)
			}
			return reply, nil
		}

		lastErr = err
		if !isRetryableError(err) {
			log.Printf("[AgentSender] Non-retryable error sending to agent %s: %v", agentID, err)
			return "", err
		}

		if attempt < maxRetries {
			log.Printf("[AgentSender] Agent %s session locked/busy (attempt %d/%d), retrying in %v",
				agentID, attempt, maxRetries, backoff)
			time.Sleep(backoff)
			backoff = min(backoff*2, maxBackoff)
		}
	}

	return "", fmt.Errorf("agent %s failed after %d attempts: %w", agentID, maxRetries, lastErr)
}

// sendToAgent executes `openclaw agent --agent <id> --message <text> --json`
// and returns the agent's reply text.
func (s *AgentSender) sendToAgent(agentID, message string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	args := []string{
		"agent",
		"--agent", agentID,
		"--message", message,
		"--json",
	}

	log.Printf("[AgentSender] Executing: openclaw %s", strings.Join(args[:3], " "))

	cmd := exec.CommandContext(ctx, "openclaw", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("agent send timed out after %v: %w", s.timeout, err)
		}
		return "", fmt.Errorf("openclaw agent send failed: %s - %w", string(output), err)
	}

	var result AgentSendResult
	if err := json.Unmarshal(output, &result); err != nil {
		log.Printf("[AgentSender] Could not parse JSON response, using raw output (len=%d)", len(output))
		return strings.TrimSpace(string(output)), nil
	}

	return result.Reply, nil
}
