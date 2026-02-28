# Agent Discovery & Delegation

## Listing Agents

Discover which agents exist and their current state:

```bash
curl $MISSION_CONTROL_API_URL/agents
```

Each agent entry includes:

| Field | Description |
|-------|-------------|
| `id` | Unique agent ID (use for `/agents/:id` and delegation) |
| `name` | Display name |
| `description` | Short summary of role and focus — best first signal of what they do |
| `status` | Current state: `idle`, `working`, `paused`, `error` |
| `model` | Model in use (e.g., `anthropic/claude-sonnet-4-5`) |
| `mention_patterns` | How to mention this agent (e.g., `["@researcher"]`) |
| `workspace_path` | Path to their workspace |
| `agent_dir_path` | Path to their agent configuration directory |
| `current_task_id` | Set when the agent is busy |
| `active_session_key` | Set when the agent has an active session |

You can assign work to any agent. If the agent is busy, Mission Control
automatically queues the task and delivers it when the agent becomes free.
Tasks are queued in priority order (lower number = higher priority), then FIFO.

## Agent Full Profile

Get an agent's complete profile including capabilities and configuration:

```bash
curl $MISSION_CONTROL_API_URL/agents/$AGENT_ID
```

Returns all list fields plus the full identity and ability documents:

| Field | Description |
|-------|-------------|
| `soul_md` | SOUL.md — values, execution protocols, vibe |
| `identity_md` | IDENTITY.md — name, role, capabilities, working style |
| `agents_md` | AGENTS.md — workspace rules, memory locations, execution principles |
| `user_md` | USER.md — user context (if set) |
| `tools_md` | TOOLS.md — environment notes, hosts, devices |
| `heartbeat_md` | HEARTBEAT.md — periodic tasks |
| `memory_md` | MEMORY.md — long-term memory summary |

Use `description` for a quick "what they're for" and `identity_md` / `soul_md`
for detailed capabilities when choosing or briefing an agent.

## Delegating via Subtasks

Delegate work by creating **subtasks** linked to your parent task and assigned to
specialist agents. Mission Control automatically notifies the assigned agent if they
are free, or queues the task if they are busy.

```bash
curl -X POST "$MISSION_CONTROL_API_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Research: [topic from parent task]",
    "description": "Full context and specific instructions for the specialist...",
    "agent_id": "SPECIALIST_AGENT_ID",
    "parent_task_id": "YOUR_PARENT_TASK_ID",
    "project_id": "PROJECT_ID",
    "status": "backlog"
  }'
```

### When to Create Subtasks

- The task requires specialist expertise (research, engineering, design, QA, content)
- Different parts of the task map to different agent capabilities
- Parallel execution across agents would speed up delivery
- A task spans multiple domains (e.g., code + design + content)
- Quality verification is needed after implementation

### Delegation Modes

The parent task's `delegation_mode` field controls what happens when a subtask
completes:

| Mode | Behavior |
|------|----------|
| `auto` (default) | Orchestrator notified immediately on subtask completion |
| `manual` | Subtask held for human approval; orchestrator notified only after human approves |

In **manual mode**, a human can:
- **Approve** — triggers the orchestrator notification to continue
- **Request Changes** — resets subtask to `executing`, sends feedback to the specialist

You do not need to handle this yourself. Mission Control manages the approval
flow — you will always receive a notification when it is time to proceed.

### Subtask Completion Notifications

**Subtask completion is push-based.** When a specialist agent updates a subtask's
status to `done` or `failed`, Mission Control automatically sends a notification
to the orchestrator (in auto mode) or holds for human approval (in manual mode).

The notification message includes:
- Subtask ID, title, and final status
- Which specialist agent completed it
- Parent task ID and title
- API commands to read results and check remaining subtasks

When you receive a subtask completion notification:

1. Read the completed subtask's results:
```bash
curl "$MISSION_CONTROL_API_URL/tasks/$SUBTASK_ID?include=phases,stories"
```

2. Check remaining subtasks:
```bash
curl $MISSION_CONTROL_API_URL/tasks/$PARENT_TASK_ID/subtasks
```

3. Decide the next action:
   - **Subtask succeeded, more work needed** → Create the next subtask
   - **Subtask failed** → Review error, re-scope, create a new subtask or retry
   - **All subtasks complete** → Verify overall results, update parent to `done`

### Handling Change Requests (For Specialist Agents)

If a human requests changes on your subtask, you will receive a notification
with the feedback. Your subtask status will be reset to `executing`.

1. Read the feedback from the notification message
2. Check task comments for additional context:
```bash
curl $MISSION_CONTROL_API_URL/tasks/$TASK_ID/comments
```
3. Make the requested changes
4. Mark the subtask `done` again when complete

### Delegation Context Checklist

When creating a subtask for a specialist agent, include in the `description`:

1. **What** — Clear objective and acceptance criteria
2. **Why** — Context from the parent task, PROJECT.md, and REQUIREMENTS.md
3. **Dependencies** — What must be done first, what's already done by other subtasks
4. **Where** — File paths for inputs and expected output locations
5. **How** — Any decisions already made that constrain the implementation
6. **Quality checks** — How the specialist should verify their own work

## Agent Task Queue

Each agent has a per-agent task queue. When you assign a task to a busy agent,
it enters their queue with status `queued`.

### Checking Your Queue (Heartbeat)

On each HEARTBEAT, check if you have queued tasks:

```bash
curl "$MISSION_CONTROL_API_URL/agents/$AGENT_ID/queue"
```

Response:

```json
{
  "agent_id": "your-agent-id",
  "queue_depth": 2,
  "tasks": [
    {"id": "task-1", "title": "...", "priority": 1, "status": "queued", ...},
    {"id": "task-2", "title": "...", "priority": 3, "status": "queued", ...}
  ]
}
```

### Picking Up Queued Work

If you are idle and have queued tasks, pick up the next one:

```bash
curl -X POST "$MISSION_CONTROL_API_URL/agents/$AGENT_ID/queue/next"
```

This atomically:
1. Takes the highest-priority task from your queue
2. Updates its status from `queued` to `backlog`
3. Sends you the full task assignment notification
4. Returns the task details

Returns 409 Conflict if you still have active tasks.

### Automatic Dispatch

You do not always need to manually dequeue. Mission Control automatically
dispatches the next queued task when:
- You finish a task (status becomes `done`, `failed`, or `cancelled`)
- A periodic background check runs (every 10 minutes)