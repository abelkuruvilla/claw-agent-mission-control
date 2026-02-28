# Mission Control API Reference

Complete endpoint reference. All endpoints are relative to `$MISSION_CONTROL_API_URL`.

## Tasks

| Action | Method | Endpoint | Body |
|--------|--------|----------|------|
| Read task (with phases/stories) | GET | `/tasks/{id}?include=phases,stories` | — |
| Create task (or subtask) | POST | `/tasks` | `{"title": "...", "description": "...", "agent_id": "...", "parent_task_id": "...", "project_id": "...", "status": "backlog", "delegation_mode": "auto"}` |
| Update task status | PUT | `/tasks/{id}/status` | `{"status": "executing"}` |
| Update task deliverables | PUT | `/tasks/{id}` | `{"project_md": "...", "requirements_md": "...", "roadmap_md": "...", "state_md": "...", "delegation_mode": "auto"}` |
| Log progress | POST | `/tasks/{id}/progress-txt` | `{"content": "[timestamp] message"}` |
| List subtasks | GET | `/tasks/{id}/subtasks` | — |
| Approve subtask delegation | POST | `/tasks/{id}/approve` | — |
| Request changes on subtask | POST | `/tasks/{id}/request-changes` | `{"comment": "What needs to change..."}` |
| List task comments | GET | `/tasks/{id}/comments` | — |
| Add comment to task | POST | `/tasks/{id}/comments` | `{"author": "human", "content": "..."}` |

### Task Status Values

| Status | Phase | When to Use |
|--------|-------|-------------|
| `queued` | — | Assigned to an agent who is busy; waiting in their queue (set by system) |
| `backlog` | — | Not yet started |
| `planning` | GSD | Researching, creating requirements/roadmap |
| `discussing` | GSD | Clarifying scope, asking questions |
| `executing` | Ralph | Actively implementing stories |
| `verifying` | GSD | Testing, validation, QA |
| `review` | — | Ready for human review |
| `done` | — | Successfully completed |
| `failed` | — | Cannot complete (explain why in progress-txt first) |

### Task Deliverable Fields

| Field | Purpose |
|-------|---------|
| `project_md` | Project overview, architecture, design decisions |
| `requirements_md` | Functional and non-functional requirements |
| `roadmap_md` | Execution plan, milestones, timeline |
| `state_md` | Current state snapshot (done, pending, blocked) |
| `prd_json` | Structured PRD data (JSON format) |

## Projects

| Action | Method | Endpoint |
|--------|--------|----------|
| Read project | GET | `/projects/{id}` |
| List project tasks | GET | `/projects/{id}/tasks` |

Project response includes:
- `name` / `description` — Project scope
- `location` — Project directory path (use for file operations)
- `status` — `active` / `on-hold` / `completed`

## Stories

| Action | Method | Endpoint | Body |
|--------|--------|----------|------|
| List stories | GET | `/tasks/{id}/stories` | — |
| Pass story | POST | `/stories/{id}/pass` | `{"commit_sha": "...", "learnings": "..."}` |
| Fail story | POST | `/stories/{id}/fail` | `{"error": "...", "iteration": 3}` |

## Phases

| Action | Method | Endpoint | Body |
|--------|--------|----------|------|
| List phases | GET | `/tasks/{id}/phases` | — |
| Update progress | POST | `/phases/{id}/progress` | `{"progress": 0.65, "message": "..."}` |
| Complete phase | POST | `/phases/{id}/complete` | `{"summary": "...", "artifacts": {"files_created": [], "files_modified": [], "commit_sha": "..."}}` |
| Fail phase | POST | `/phases/{id}/fail` | `{"error": "...", "recoverable": true, "suggestion": "..."}` |

## Agents

| Action | Method | Endpoint | Body |
|--------|--------|----------|------|
| List all agents | GET | `/agents` | — |
| Get agent details | GET | `/agents/{id}` | — |
| Get agent's task queue | GET | `/agents/{id}/queue` | — |
| Dequeue next task | POST | `/agents/{id}/queue/next` | — |

### Agent Queue

The queue endpoint returns tasks assigned to the agent with status `queued`, ordered
by priority (ascending) then creation time (FIFO):

```json
{
  "agent_id": "agent-123",
  "queue_depth": 3,
  "tasks": [...]
}
```

The dequeue endpoint atomically promotes the next queued task to `backlog` and notifies
the agent. Returns 409 if the agent has active tasks.

## Events

| Action | Method | Endpoint |
|--------|--------|----------|
| List events | GET | `/events?task_id={id}` |