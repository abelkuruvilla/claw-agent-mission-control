---
name: mission-control
description: >
  Interact with the Mission Control agentic task management system. Use this skill whenever you
  receive a task assignment message from Mission Control, or need to: (1) read task details,
  phases, and stories, (2) update task status through the workflow (backlog → planning →
  discussing → executing → verifying → review → done), (3) report progress on tasks,
  (4) manage phases (update progress, complete, fail), (5) execute stories via the Ralph Loop
  (pick → implement → test → pass/fail → learn → repeat), (6) discover and inspect other
  agents, (7) delegate work by creating subtasks assigned to specialist agents,
  (8) create GSD planning artifacts (PROJECT.md, REQUIREMENTS.md, ROADMAP.md, STATE.md),
  or (9) read project context.
  Trigger this skill on any mention of "mission control", "task", "story", "phase", "GSD",
  "Ralph Loop", "progress update", "delegation", "subtask", or when you receive a message
  saying "You have been assigned a new task in Mission Control".
---

# Mission Control Skill

Interact with the Mission Control platform to receive tasks, plan work, execute stories,
report progress, and coordinate with other agents.

## Overview

Mission Control is an agentic task management system. You are an agent registered in it.

**Task delivery is push-based with queuing:** When a task is assigned to you and you are
free, Mission Control sends you a message directly via `openclaw agent --agent <your-id>`.
If you are busy when a task is assigned, it enters your **task queue** with status `queued`
and will be delivered when you become free. On each HEARTBEAT, you should check your queue
and pick up the next task if you are idle.

You execute tasks using two protocols:

- **GSD (Get Shit Done)** — for planning: Research → Requirements → Roadmap → Stories → Delegation
- **Ralph Loop** — for execution: Pick → Implement → Test → Pass/Fail → Learn → Repeat

All communication with Mission Control happens via its REST API using `curl`.

## Configuration

The skill requires these environment variables:

- `MISSION_CONTROL_API_URL` — Base URL of the Mission Control API (e.g., `http://localhost:8080/api/v1`)
- `AGENT_ID` — Your agent's unique ID in Mission Control

Throughout this skill, `{{API_BASE_URL}}` refers to `$MISSION_CONTROL_API_URL` and
`{{AGENT_ID}}` refers to `$AGENT_ID`.

## Receiving Task Assignments

Mission Control pushes task assignments to you directly. When you receive a message like:

> You have been assigned a new task in Mission Control.
>
> **Task ID:** abc-123
> **Title:** Build the feature
> **Description:** ...

Follow these steps:

1. **Extract** the Task ID from the message.
2. **Fetch** full task details:

```bash
curl "$MISSION_CONTROL_API_URL/tasks/$TASK_ID?include=phases,stories"
```

3. **Read** the `description` field carefully — it defines your primary objective.
4. **Update status** to `planning` and begin the GSD protocol.
5. **Execute** using the Ralph Loop.
6. **Report** progress and results back to Mission Control.

Any reply you produce is automatically saved as a comment on the task, so your initial
acknowledgment and plan will be visible in Mission Control.

## Task Lifecycle

Update your status as you move through each stage:

```bash
curl -X PUT "$MISSION_CONTROL_API_URL/tasks/$TASK_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "planning"}'
```

Valid statuses in order: `queued` → `backlog` → `planning` → `discussing` → `executing` → `verifying` → `review` → `done` (or `failed`).

- **`queued`** — Task is assigned to you but waiting because you are busy. Mission Control
  sets this automatically. You do not set this status yourself. When you finish your current
  work, Mission Control promotes the next queued task to `backlog` and notifies you.

## GSD Protocol — Planning

Use GSD when you first receive a task:

1. **Research** — Understand the problem domain and requirements
2. **Requirements** — Define what needs to be done → create/update `REQUIREMENTS.md`
3. **Roadmap** — Plan phases and milestones → create/update `ROADMAP.md`
4. **Stories** — Break into executable stories with acceptance criteria
5. **Delegation** — Create subtasks assigned to specialist agents, or execute yourself if no suitable agent exists

Create these artifacts and push them to Mission Control:

```bash
curl -X PUT "$MISSION_CONTROL_API_URL/tasks/$TASK_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "project_md": "# Project Overview\n...",
    "requirements_md": "# Requirements\n...",
    "roadmap_md": "# Roadmap\n...",
    "state_md": "# Current State\n..."
  }'
```

## Ralph Loop — Execution

For each story assigned to you:

1. **Pick** — Select the highest priority incomplete story
2. **Implement** — Write code, create content, make changes
3. **Test** — Run quality checks (tests, lint, typecheck, self-review)
4. **Report**:
   - If tests pass → `git commit` and mark story passed
   - If tests fail → fix and retry (up to 3 attempts, then fail the story)
5. **Learn** — Append learnings to progress.txt
6. **Repeat** — Loop until all stories are complete

Mark a story as passed:

```bash
curl -X POST "$MISSION_CONTROL_API_URL/stories/$STORY_ID/pass" \
  -H "Content-Type: application/json" \
  -d '{"commit_sha": "abc123", "learnings": "Implementation notes here"}'
```

Mark a story as failed:

```bash
curl -X POST "$MISSION_CONTROL_API_URL/stories/$STORY_ID/fail" \
  -H "Content-Type: application/json" \
  -d '{"error": "Description of what failed", "iteration": 3}'
```

## Progress Reporting

Log your work after significant actions:

```bash
curl -X POST "$MISSION_CONTROL_API_URL/tasks/$TASK_ID/progress-txt" \
  -H "Content-Type: application/json" \
  -d '{"content": "[2026-02-18 10:30] What you did and what you learned"}'
```

Timestamp entries. Note files created/modified. Document decisions and blockers.

## Agent Discovery & Delegation

Delegate work by creating **subtasks** assigned to specialist agents. Never do specialist
work yourself if a suitable agent exists — discover agents dynamically, then delegate.

See `references/agents-api.md` for full agent discovery and delegation details.

### Step 1: Discover Available Agents

The agent roster is dynamic. Always query it before delegating:

```bash
# List all agents with their descriptions and status
curl $MISSION_CONTROL_API_URL/agents

# Get full profile for a specific agent (identity, capabilities, tools)
curl $MISSION_CONTROL_API_URL/agents/$AGENT_ID
```

Read each agent's `description` to understand their specialization. For detailed
capabilities, check `identity_md`. Only assign work to agents with `status: idle`.

### Step 2: Create Subtasks

Break your task into subtasks and assign each to the appropriate specialist:

```bash
curl -X POST "$MISSION_CONTROL_API_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Research: [topic]",
    "description": "Context and specific instructions for the specialist...",
    "agent_id": "SPECIALIST_AGENT_ID",
    "parent_task_id": "YOUR_PARENT_TASK_ID",
    "project_id": "PROJECT_ID",
    "status": "backlog"
  }'
```

The `parent_task_id` links the subtask to your parent task. When an agent is assigned,
Mission Control automatically notifies them.

### Step 3: Receive Subtask Completion Notifications

**Subtask completion is push-based.** When a specialist agent marks a subtask as
`done` or `failed`, Mission Control automatically sends you a notification message
with the subtask results and next steps. You do NOT need to poll.

**Important:** The parent task's `delegation_mode` controls whether you receive
notifications immediately or after human approval:
- `auto` (default) — You are notified immediately when a subtask completes.
- `manual` — A human reviews the subtask results first. You are only notified
  after the human approves the delegation. The human may also request changes
  on the subtask, which re-notifies the specialist agent.

You do not need to check delegation_mode yourself. Mission Control handles this
automatically — you will always receive a notification when it's time to continue.

The notification includes:
- The subtask ID, title, and final status (`done` or `failed`)
- Which specialist agent completed it
- The parent task ID
- API commands to read the subtask results and check remaining subtasks

When you receive a subtask completion notification:

1. **Read** the completed subtask's results:
```bash
curl "$MISSION_CONTROL_API_URL/tasks/$SUBTASK_ID?include=phases,stories"
```

2. **Check** remaining subtasks:
```bash
curl $MISSION_CONTROL_API_URL/tasks/$PARENT_TASK_ID/subtasks
```

3. **Decide** what to do next:
   - If the subtask succeeded and more work is needed → create the next subtask
   - If the subtask failed → review the error, re-scope, and create a new subtask
   - If all subtasks are done → verify results and mark the parent task as `done`

### Delegation Modes

When creating a task, you can set `delegation_mode` to control the approval flow:

```bash
curl -X POST "$MISSION_CONTROL_API_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Main task",
    "delegation_mode": "auto",
    ...
  }'
```

- **`auto`** — Orchestrator is notified immediately when subtasks complete.
  Best for trusted automated workflows.
- **`manual`** — Subtask results are held for human review. The human can
  approve (which notifies the orchestrator) or request changes (which
  re-notifies the specialist). Best when human oversight is required.

### Delegation Workflow

Follow this general-purpose workflow (adapts to any task type):

1. **DISCOVER** — `GET /agents` → read descriptions → identify who can handle what
2. **PLAN** — Break the parent task into subtasks suited to each specialist
3. **DELEGATE** — Create subtasks with `parent_task_id`, assign to the right agent
4. **WAIT** — Mission Control will push-notify you when each subtask completes (auto mode) or after human approval (manual mode)
5. **CONTINUE** — On each notification, read results and create the next subtask
6. **VERIFY** — Create a QA subtask if verification is needed
7. **COMPLETE** — Update parent task to `done` once all subtasks pass

### Handling Change Requests

If your task is sent back with a change request (status reset to `executing`),
you will receive a notification containing the human's feedback. When this happens:

1. Read the feedback in the notification message
2. Check the task comments for additional context:
```bash
curl $MISSION_CONTROL_API_URL/tasks/$TASK_ID/comments
```
3. Make the requested changes
4. Update status back to `done` when complete

## Phase Management

Manage GSD planning phases. See `references/phases-api.md` for details.

Quick reference:

```bash
# List phases
curl $MISSION_CONTROL_API_URL/tasks/$TASK_ID/phases

# Update progress (0.0 to 1.0)
curl -X POST "$MISSION_CONTROL_API_URL/phases/$PHASE_ID/progress" \
  -H "Content-Type: application/json" \
  -d '{"progress": 0.65, "message": "Status message"}'

# Complete a phase
curl -X POST "$MISSION_CONTROL_API_URL/phases/$PHASE_ID/complete" \
  -H "Content-Type: application/json" \
  -d '{"summary": "What was accomplished", "artifacts": {"files_created": [], "files_modified": []}}'

# Fail a phase
curl -X POST "$MISSION_CONTROL_API_URL/phases/$PHASE_ID/fail" \
  -H "Content-Type: application/json" \
  -d '{"error": "Why it failed", "recoverable": true, "suggestion": "What to try"}'
```

## Task Completion

When all phases and stories are complete:

1. Update status to `verifying` — run final checks
2. Create `SUMMARY.md` with outcomes
3. Update status to `review` (if human review needed) or `done`

If the task cannot be completed:

```bash
curl -X POST "$MISSION_CONTROL_API_URL/tasks/$TASK_ID/progress-txt" \
  -H "Content-Type: application/json" \
  -d '{"content": "[FAILED] Reason for failure"}'

curl -X PUT "$MISSION_CONTROL_API_URL/tasks/$TASK_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "failed"}'
```

### Completing Subtasks (For Specialist Agents)

If your task has a `parent_task_id`, you are working on a subtask delegated by an
orchestrator. When you finish:

1. **Log your results** in `progress_txt` — the orchestrator will read this:
```bash
curl -X POST "$MISSION_CONTROL_API_URL/tasks/$TASK_ID/progress-txt" \
  -H "Content-Type: application/json" \
  -d '{"content": "[DONE] Summary of what was accomplished, key findings, files changed, etc."}'
```

2. **Mark the subtask as `done`** (or `failed` if it cannot be completed):
```bash
curl -X PUT "$MISSION_CONTROL_API_URL/tasks/$TASK_ID/status" \
  -H "Content-Type: application/json" \
  -d '{"status": "done"}'
```

**This is critical:** When you update status to `done` or `failed`, Mission Control
automatically notifies the orchestrator agent who delegated the subtask (in auto
mode) or flags the subtask for human approval (in manual mode). The orchestrator
will then read your results and decide on next steps. If you do not update the
status, the workflow cannot continue.

**If you receive a change request:** A human may review your work and send it back
with feedback. You will receive a new notification with the requested changes.
Check the task comments for details, make the changes, and mark the task `done` again.

## Task Queue & Heartbeat

Mission Control maintains a **per-agent task queue**. When a task is assigned to you while
you are busy, it enters your queue with status `queued` instead of being delivered
immediately. The queue is priority-weighted FIFO: higher-priority tasks (lower priority
number) are delivered first; within the same priority, first-in-first-out order applies.

### Automatic Dispatch

When you finish a task (status becomes `done`, `failed`, or `cancelled`), Mission Control
automatically checks your queue and dispatches the next task. A background processor also
runs every 10 minutes to catch any missed dispatches.

### Heartbeat Queue Check

On each **HEARTBEAT**, check your queue and pick up work if you are idle:

```bash
# Check your queue
curl "$MISSION_CONTROL_API_URL/agents/$AGENT_ID/queue"
```

This returns `queue_depth` (number of waiting tasks) and the ordered task list.

If there are queued tasks and you are free, pick up the next one:

```bash
# Dequeue and start the next task
curl -X POST "$MISSION_CONTROL_API_URL/agents/$AGENT_ID/queue/next"
```

This atomically moves the highest-priority queued task to `backlog`, notifies you with
the full task details, and returns the task. If you are still busy (have active tasks),
it returns a 409 Conflict.

### Queue Behavior for Delegators

When you create a subtask assigned to an agent who is busy, the subtask automatically
enters that agent's queue. You do not need to handle this — Mission Control manages it.
The agent will be notified when they become free. You can still assign tasks to busy
agents; they will be queued and processed in priority order.

## Key Principles

- **Push-based delivery with queuing**: Tasks are sent to you when you are free. If you are busy, they queue and are delivered when you finish. On each HEARTBEAT, check your queue.
- **Push-based subtask completion**: When a subtask you delegated reaches `done` or `failed`, Mission Control automatically notifies you. You do not need to poll for completion.
- **Orchestrators must delegate, not execute**: If you are an orchestrator, you MUST NOT do specialist work yourself (research, engineering, design, testing, content creation). Your role is to plan, delegate via subtasks, monitor, and verify.
- **Specialists must mark subtasks done**: If you are working on a subtask (has `parent_task_id`), you MUST update status to `done` or `failed` when finished. This triggers the orchestrator notification.
- **Discover agents dynamically**: Never hardcode agent IDs. Always query `GET /agents` before delegating — the roster of specialist agents changes over time.
- **Delegate via subtasks, not sub-agents**: Create subtasks with `parent_task_id` assigned to registered agents. Do not spawn ephemeral sub-agents.
- **Confirm agent fit before delegating**: Read the target agent's `description` and `identity_md` to verify they are the right specialist for the subtask.
- **Memory persists via files, not context**: Each Ralph Loop iteration starts with fresh context. Write state to files.
- **Keep stories small**: One story should fit in one context window. If it needs 5+ files of context, split it.
- **Commit early and often**: Atomic commits per story. Never commit broken code.
- **Always run tests before marking done**: No exceptions.
- **Document decisions**: Future you (or another agent) will need to understand why.
- **Reply matters**: Your initial reply to a task assignment is saved as a comment visible in Mission Control.

## Reference Files

For detailed API documentation, read these files as needed:

- `references/api-reference.md` — Complete API endpoint reference table
- `references/agents-api.md` — Agent discovery, profiles, and subtask delegation
- `references/phases-api.md` — Phase lifecycle management
- `references/workflow-patterns.md` — Common workflow patterns for different project types
