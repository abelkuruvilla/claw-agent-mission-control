# API Reference

**Version:** 1.0.0  
**Base URL:** `http://localhost:8080/api/v1`  
**Protocol:** REST (JSON) + WebSocket

---

## Table of Contents

- [Authentication](#authentication)
- [Common Headers](#common-headers)
- [Response Format](#response-format)
- [Error Handling](#error-handling)
- [Rate Limiting](#rate-limiting)
- [Endpoints](#endpoints)
  - [Health & Status](#health--status)
  - [Agents](#agents)
  - [Tasks](#tasks)
  - [Phases (GSD)](#phases-gsd)
  - [Stories (Ralph)](#stories-ralph)
  - [Events](#events)
  - [Settings](#settings)
  - [Projects](#projects)
  - [Comments](#comments)
- [WebSocket Events](#websocket-events)
- [Agent Self-Reporting](#agent-self-reporting)

---

## Authentication

**Current Version:** No authentication required (single-user system)

**Future Versions:** Will support token-based authentication for multi-user deployments.

---

## CORS (Cross-Origin Resource Sharing)

The API supports cross-origin requests to enable frontend development and private network access.

### CORS Configuration

**Development Mode** (`ENV=development`):
- **Allowed Origins:** `*` (all origins)
- **Credentials:** Enabled
- **Methods:** GET, POST, PUT, DELETE, PATCH, OPTIONS
- **Headers:** All requested headers allowed
- **Purpose:** Enable local frontend development on different ports (e.g., `localhost:3000` → `localhost:8080`)

**Production Mode** (`ENV=production`):
- **Allowed Origins:** `*` (all origins, configurable via `CORS_ALLOWED_ORIGINS`)
- **Credentials:** Enabled
- **Methods:** GET, POST, PUT, DELETE, PATCH, OPTIONS
- **Headers:** All requested headers allowed
- **Purpose:** Private network deployment (internal dashboards, development tools)

### Network Access

This application is designed for **private network deployment**:
- Intended for internal tools, team dashboards, and development environments
- Default configuration allows access from any origin on the local network
- Suitable for home labs, private VPNs, and internal corporate networks

**Security Considerations:**
- **Do not expose directly to the public internet** without implementing authentication
- For public deployments, configure specific allowed origins and enable authentication
- The permissive CORS policy assumes a trusted network environment

### Custom CORS Configuration

To restrict origins in production, set the `CORS_ALLOWED_ORIGINS` environment variable:

```bash
# Single origin
CORS_ALLOWED_ORIGINS=https://dashboard.example.com

# Multiple origins (comma-separated)
CORS_ALLOWED_ORIGINS=https://dashboard.example.com,https://app.example.com
```

---

## Common Headers

### Request Headers

```http
Content-Type: application/json
Accept: application/json
```

### Response Headers

```http
Content-Type: application/json
X-Request-ID: <uuid>
```

---

## Response Format

### Success Response

```json
{
  "data": { /* response data */ },
  "meta": {
    "timestamp": "2026-02-08T22:30:00Z"
  }
}
```

### List Response

```json
{
  "data": [ /* array of items */ ],
  "meta": {
    "total": 42,
    "page": 1,
    "per_page": 20,
    "timestamp": "2026-02-08T22:30:00Z"
  }
}
```

---

## Error Handling

### Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": {
      "field": "name",
      "reason": "Name is required"
    }
  }
}
```

### HTTP Status Codes

| Code | Meaning | Usage |
|------|---------|-------|
| `200` | OK | Successful GET, PUT |
| `201` | Created | Successful POST |
| `204` | No Content | Successful DELETE |
| `400` | Bad Request | Invalid request data |
| `404` | Not Found | Resource doesn't exist |
| `409` | Conflict | Resource conflict (duplicate name, etc.) |
| `422` | Unprocessable Entity | Validation failed |
| `500` | Internal Server Error | Server error |
| `501` | Not Implemented | Endpoint not yet implemented |

### Error Codes

| Code | Description |
|------|-------------|
| `VALIDATION_ERROR` | Request validation failed |
| `NOT_FOUND` | Resource not found |
| `CONFLICT` | Resource conflict |
| `INTERNAL_ERROR` | Unexpected server error |
| `NOT_IMPLEMENTED` | Feature not yet implemented |
| `OPENCLAW_ERROR` | OpenClaw Gateway error |

---

## Rate Limiting

**Current Version:** No rate limiting

**Future Versions:** 
- 100 requests per minute per IP
- 1000 requests per hour per IP
- Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

---

## Endpoints

### Health & Status

#### Get Health Check

```http
GET /health
```

**Response:**

```json
{
  "status": "ok"
}
```

---

#### Get System Status

```http
GET /api/v1/status
```

**Response:**

```json
{
  "version": "1.0.0",
  "status": "running",
  "openclaw": {
    "connected": true,
    "gateway_url": "ws://127.0.0.1:18789",
    "uptime_seconds": 86400
  },
  "database": {
    "connected": true,
    "path": "./data/mission-control.db"
  },
  "stats": {
    "total_agents": 5,
    "active_agents": 2,
    "total_tasks": 42,
    "active_tasks": 3,
    "completed_tasks": 35
  }
}
```

---

### Agents

#### List All Agents

```http
GET /api/v1/agents
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status (idle/working/paused/error) |
| `page` | int | Page number (default: 1) |
| `per_page` | int | Items per page (default: 20, max: 100) |

**Response:**

```json
{
  "data": [
    {
      "id": "jarvis",
      "name": "Jarvis",
      "description": "Personal AI Software Engineer",
      "status": "working",
      "model": "anthropic/claude-sonnet-4-5",
      "workspace_path": "~/.openclaw/workspace-jarvis",
      "current_task_id": "task-123",
      "sub_agents": [
        {
          "id": "subagent-456",
          "name": "Builder",
          "status": "running",
          "task_id": "task-123"
        }
      ],
      "created_at": "2026-02-01T10:00:00Z",
      "updated_at": "2026-02-08T22:30:00Z"
    }
  ],
  "meta": {
    "total": 5,
    "page": 1,
    "per_page": 20
  }
}
```

---

#### Create Agent

```http
POST /api/v1/agents
```

**Minimal Request (Auto-Generated Identity):**

When only `name` and `description` are provided, the system automatically generates identity files incorporating GSD and Ralph principles:

```json
{
  "name": "researcher",
  "description": "Deep research specialist for comprehensive analysis and fact-finding",
  "model": "anthropic/claude-sonnet-4-5",
  "mention_patterns": ["@researcher"]
}
```

The system generates:
- **SOUL.md** - Personality and values based on the description
- **IDENTITY.md** - Role, capabilities, and working style
- **AGENTS.md** - Workspace rules with GSD/Ralph workflow guidelines
- **USER.md** - Placeholder for user information
- **TOOLS.md** - Environment-specific notes
- **HEARTBEAT.md** - Periodic task configuration
- **MEMORY.md** - Initial memory structure

**Full Request (Explicit Identity):**

```json
{
  "name": "researcher",
  "description": "Deep research specialist",
  "model": "anthropic/claude-sonnet-4-5",
  "mention_patterns": ["@researcher"],
  "soul_md": "# SOUL.md\n\nYou are a research specialist...",
  "identity_md": "# IDENTITY.md\n\n- Name: Researcher\n...",
  "agents_md": "# AGENTS.md\n\n...",
  "user_md": "# USER.md\n\n...",
  "tools_md": "# TOOLS.md\n\n...",
  "heartbeat_md": "# HEARTBEAT.md\n\n..."
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": "researcher",
    "name": "Researcher",
    "description": "Deep research specialist for comprehensive analysis and fact-finding",
    "status": "idle",
    "model": "anthropic/claude-sonnet-4-5",
    "workspace_path": "~/.openclaw/workspace-researcher",
    "agent_dir_path": "~/.openclaw/agents/researcher/agent",
    "created_at": "2026-02-08T22:35:00Z",
    "updated_at": "2026-02-08T22:35:00Z"
  }
}
```

**Creates:**
- Workspace at `~/.openclaw/workspace-<name>/`
- Agent directory at `~/.openclaw/agents/<name>/`
- Entry in `~/.openclaw/openclaw.json`
- Git-initialized workspace with config files (auto-generated or explicit)

---

#### Get Agent

```http
GET /api/v1/agents/:id
```

**Response:**

```json
{
  "data": {
    "id": "jarvis",
    "name": "Jarvis",
    "description": "Personal AI Software Engineer",
    "status": "working",
    "model": "anthropic/claude-sonnet-4-5",
    "mention_patterns": ["@jarvis"],
    "workspace_path": "~/.openclaw/workspace-jarvis",
    "agent_dir_path": "~/.openclaw/agents/jarvis/agent",
    "soul_md": "# SOUL.md\n\n...",
    "identity_md": "# IDENTITY.md\n\n...",
    "agents_md": "# AGENTS.md\n\n...",
    "user_md": "# USER.md\n\n...",
    "tools_md": "# TOOLS.md\n\n...",
    "heartbeat_md": "# HEARTBEAT.md\n\n...",
    "memory_md": "# MEMORY.md\n\n...",
    "current_task_id": "task-123",
    "active_session_key": "session-xyz",
    "sub_agents": [...],
    "tasks": [...],
    "created_at": "2026-02-01T10:00:00Z",
    "updated_at": "2026-02-08T22:30:00Z"
  }
}
```

---

#### Update Agent

```http
PUT /api/v1/agents/:id
```

**Request Body:**

```json
{
  "description": "Updated description",
  "model": "anthropic/claude-opus-4-5",
  "soul_md": "# Updated SOUL.md content...",
  "agents_md": "# Updated AGENTS.md content..."
}
```

**Response:** `200 OK`

```json
{
  "data": { /* updated agent */ }
}
```

---

#### Delete Agent

```http
DELETE /api/v1/agents/:id
```

**Response:** `204 No Content`

**Note:** This removes the agent from OpenClaw configuration and deletes its workspace.

---

### Agent Chat Sessions

#### Start Chat Session

Create a new chat session with an agent via OpenClaw's `sessions_spawn`.

```http
POST /api/v1/agents/:id/sessions
```

**Request Body:**

```json
{
  "label": "quick-question",
  "context": "User wants to ask about task progress"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "id": "chat-session-123",
    "agent_id": "jarvis",
    "session_key": "agent:jarvis:chat:chat-session-123",
    "status": "active",
    "started_at": "2026-02-09T20:00:00Z",
    "message_count": 0
  }
}
```

---

#### End Chat Session

Terminate an active chat session.

```http
DELETE /api/v1/agents/:id/sessions/:sessionId
```

**Response:** `200 OK`

```json
{
  "data": {
    "id": "chat-session-123",
    "agent_id": "jarvis",
    "session_key": "agent:jarvis:chat:chat-session-123",
    "status": "ended",
    "started_at": "2026-02-09T20:00:00Z",
    "ended_at": "2026-02-09T20:15:00Z",
    "message_count": 8
  }
}
```

---

#### List Chat Sessions

Get all chat sessions for an agent.

```http
GET /api/v1/agents/:id/sessions
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status (active/ended) |
| `limit` | int | Max sessions to return (default: 20) |
| `offset` | int | Pagination offset |

**Response:**

```json
{
  "data": [
    {
      "id": "chat-session-123",
      "agent_id": "jarvis",
      "session_key": "agent:jarvis:chat:chat-session-123",
      "status": "active",
      "started_at": "2026-02-09T20:00:00Z",
      "ended_at": null,
      "message_count": 8
    },
    {
      "id": "chat-session-122",
      "agent_id": "jarvis",
      "session_key": "agent:jarvis:chat:chat-session-122",
      "status": "ended",
      "started_at": "2026-02-09T18:00:00Z",
      "ended_at": "2026-02-09T18:30:00Z",
      "message_count": 12
    }
  ],
  "meta": {
    "total": 25,
    "limit": 20,
    "offset": 0
  }
}
```

---

#### Get Session Messages

Retrieve message history for a specific chat session.

```http
GET /api/v1/agents/:id/sessions/:sessionId/messages
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | int | Max messages to return (default: 50) |
| `before` | string | Get messages before this timestamp |
| `after` | string | Get messages after this timestamp |

**Response:**

```json
{
  "data": [
    {
      "id": "msg-1",
      "session_id": "chat-session-123",
      "role": "user",
      "content": "Can you review the authentication code we wrote yesterday?",
      "created_at": "2026-02-09T20:01:00Z",
      "token_count": null
    },
    {
      "id": "msg-2",
      "session_id": "chat-session-123",
      "role": "agent",
      "content": "I'll review the auth code. Let me check the files in src/auth/...\n\nI see a few areas we can improve:\n1. Add rate limiting to login endpoint\n2. Implement token refresh mechanism\n3. Add audit logging for auth events",
      "created_at": "2026-02-09T20:02:00Z",
      "model": "anthropic/claude-sonnet-4-5",
      "token_count": 245
    },
    {
      "id": "msg-3",
      "session_id": "chat-session-123",
      "role": "user",
      "content": "Great! Can you create a task for implementing those improvements?",
      "created_at": "2026-02-09T20:05:00Z",
      "token_count": null
    }
  ],
  "meta": {
    "total": 8,
    "session": {
      "id": "chat-session-123",
      "status": "active",
      "started_at": "2026-02-09T20:00:00Z"
    }
  }
}
```

---

#### Send Message

Send a message in an active chat session.

```http
POST /api/v1/agents/:id/sessions/:sessionId/messages
```

**Request Body:**

```json
{
  "content": "What's the status of the dashboard API task?"
}
```

**Response:** `201 Created`

```json
{
  "data": {
    "user_message": {
      "id": "msg-4",
      "session_id": "chat-session-123",
      "role": "user",
      "content": "What's the status of the dashboard API task?",
      "created_at": "2026-02-09T20:10:00Z"
    },
    "agent_message": {
      "id": "msg-5",
      "session_id": "chat-session-123",
      "role": "agent",
      "content": "The dashboard API task is currently in the executing phase. We've completed the agent endpoints and are now working on the task endpoints. Progress is at about 65%.",
      "created_at": "2026-02-09T20:10:15Z",
      "model": "anthropic/claude-sonnet-4-5",
      "token_count": 312
    }
  }
}
```

**Error Responses:**

```json
// Session not found
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Chat session not found"
  }
}

// Session ended
{
  "error": {
    "code": "SESSION_ENDED",
    "message": "Cannot send message to ended session"
  }
}

// Agent busy
{
  "error": {
    "code": "AGENT_BUSY",
    "message": "Agent is currently executing a task"
  }
}
```

---

### Tasks

#### List Tasks

```http
GET /api/v1/tasks
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status |
| `agent_id` | string | Filter by assigned agent |
| `project_id` | string | Filter by project |
| `priority` | int | Filter by priority (1-5) |
| `page` | int | Page number |
| `per_page` | int | Items per page |

**Response:**

```json
{
  "data": [
    {
      "id": "task-123",
      "title": "Build Dashboard API",
      "description": "Create REST API for dashboard with agent and task endpoints",
      "agent_id": "jarvis",
      "project_id": "project-456",
      "status": "executing",
      "priority": 1,
      "git_branch": "feature/dashboard-api",
      "stories_total": 5,
      "stories_passed": 3,
      "created_at": "2026-02-08T20:00:00Z",
      "updated_at": "2026-02-08T22:30:00Z",
      "started_at": "2026-02-08T20:05:00Z",
      "completed_at": null
    }
  ],
  "meta": {
    "total": 42,
    "page": 1,
    "per_page": 20
  }
}
```

---

#### Create Task

```http
POST /api/v1/tasks
```

**Request Body:**

```json
{
  "title": "Build Payment Integration",
  "description": "Integrate Stripe payment processing",
  "agent_id": "jarvis",
  "project_id": "project-456",
  "priority": 1,
  "project_md": "# PROJECT.md\n\n...",
  "requirements_md": "# REQUIREMENTS.md\n\n..."
}
```

**Note:** All tasks automatically use both protocols:
- **GSD** for planning (creates requirements, roadmap, stories)
- **Ralph Loop** for execution (iterates on stories until complete)

**Response:** `201 Created`

```json
{
  "data": { /* created task */ }
}
```

---

#### Get Task

```http
GET /api/v1/tasks/:id
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `include` | string | Comma-separated: phases,stories,events |

**Response:**

```json
{
  "data": {
    "id": "task-123",
    "title": "Build Dashboard API",
    /* ...task fields... */
    "phases": [
      {
        "id": "phase-1",
        "sequence": 1,
        "title": "Research & Planning",
        "status": "done",
        "context_md": "...",
        "research_md": "...",
        "plan_md": "..."
      },
      {
        "id": "phase-2",
        "sequence": 2,
        "title": "Implementation",
        "status": "executing",
        "plan_md": "..."
      }
    ],
    "events": [...]
  }
}
```

---

#### Update Task

```http
PUT /api/v1/tasks/:id
```

**Request Body:**

```json
{
  "title": "Updated title",
  "description": "Updated description",
  "priority": 2,
  "status": "paused"
}
```

**Response:** `200 OK`

---

#### Delete Task

```http
DELETE /api/v1/tasks/:id
```

**Response:** `204 No Content`

---

#### Update Task Status

```http
PUT /api/v1/tasks/:id/status
```

**Request Body:**

```json
{
  "status": "executing"
}
```

**Valid statuses:** `backlog`, `planning`, `discussing`, `executing`, `verifying`, `review`, `done`, `failed`

**Response:** `200 OK`

---

#### Start Task

```http
POST /api/v1/tasks/:id/start
```

**Response:** `200 OK`

```json
{
  "data": {
    "task_id": "task-123",
    "session_key": "session-xyz",
    "status": "executing",
    "started_at": "2026-02-08T22:40:00Z"
  }
}
```

---

#### Stop Task

```http
POST /api/v1/tasks/:id/stop
```

**Response:** `200 OK`

```json
{
  "data": {
    "task_id": "task-123",
    "status": "paused",
    "stopped_at": "2026-02-08T22:45:00Z"
  }
}
```

---

### Phases (GSD)

#### List Phases

```http
GET /api/v1/tasks/:task_id/phases
```

**Response:**

```json
{
  "data": [
    {
      "id": "phase-1",
      "task_id": "task-123",
      "sequence": 1,
      "title": "Research & Planning",
      "description": "Understand requirements and create plan",
      "status": "done",
      "context_md": "# CONTEXT.md\n\n...",
      "research_md": "# RESEARCH.md\n\n...",
      "plan_md": "# PLAN.md\n\n...",
      "summary_md": "# SUMMARY.md\n\n...",
      "verification_result": "passed",
      "created_at": "2026-02-08T20:05:00Z",
      "updated_at": "2026-02-08T20:30:00Z"
    }
  ]
}
```

---

#### Create Phase

```http
POST /api/v1/tasks/:task_id/phases
```

**Request Body:**

```json
{
  "sequence": 2,
  "title": "Implementation",
  "description": "Build the API endpoints",
  "context_md": "# CONTEXT.md\n\n..."
}
```

**Response:** `201 Created`

---

#### Get Phase

```http
GET /api/v1/phases/:id
```

**Response:**

```json
{
  "data": { /* phase object */ }
}
```

---

#### Update Phase

```http
PUT /api/v1/phases/:id
```

**Request Body:**

```json
{
  "status": "executing",
  "plan_md": "# Updated plan..."
}
```

**Response:** `200 OK`

---

### Stories (Ralph)

#### List Stories

```http
GET /api/v1/tasks/:task_id/stories
```

**Response:**

```json
{
  "data": [
    {
      "id": "story-1",
      "task_id": "task-456",
      "sequence": 1,
      "title": "User can login with email and password",
      "description": "Implement login endpoint with JWT auth",
      "passes": true,
      "priority": 1,
      "acceptance_criteria": [
        "POST /api/auth/login accepts email and password",
        "Returns JWT token on success",
        "Returns 401 on invalid credentials"
      ],
      "iterations": 2,
      "last_error": null,
      "created_at": "2026-02-08T21:00:00Z",
      "updated_at": "2026-02-08T21:30:00Z"
    }
  ]
}
```

---

#### Create Story

```http
POST /api/v1/tasks/:task_id/stories
```

**Request Body:**

```json
{
  "sequence": 2,
  "title": "User can logout",
  "description": "Implement logout endpoint",
  "priority": 1,
  "acceptance_criteria": [
    "POST /api/auth/logout invalidates token",
    "Returns 200 on success"
  ]
}
```

**Response:** `201 Created`

---

### Events

#### List Events

```http
GET /api/v1/events
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `task_id` | string | Filter by task |
| `agent_id` | string | Filter by agent |
| `type` | string | Filter by event type |
| `since` | timestamp | Events after this time |
| `limit` | int | Max events to return (default: 50, max: 500) |

**Response:**

```json
{
  "data": [
    {
      "id": "event-789",
      "task_id": "task-123",
      "agent_id": "jarvis",
      "type": "phase_completed",
      "message": "Phase 1 (Research & Planning) completed successfully",
      "details": {
        "phase_id": "phase-1",
        "duration_seconds": 1500
      },
      "created_at": "2026-02-08T20:30:00Z"
    }
  ],
  "meta": {
    "total": 156,
    "limit": 50
  }
}
```

**Event Types:**
- `task_created`
- `task_assigned`
- `task_started`
- `task_completed`
- `task_failed`
- `phase_started`
- `phase_completed`
- `phase_failed`
- `story_passed`
- `story_failed`
- `agent_spawned`
- `execution_error`
- `verification_passed`
- `verification_failed`
- `commit_created`

---

### Settings

#### Get Settings

```http
GET /api/v1/settings
```

**Response:**

```json
{
  "data": {
    "id": "settings-1",
    "openclaw_gateway_url": "ws://127.0.0.1:18789",
    "openclaw_gateway_token": "***",
    "default_model": "anthropic/claude-sonnet-4-5",
    "max_parallel_executions": 3,
    "default_project_directory": "~/projects",
    "gsd_depth": "standard",
    "gsd_mode": "interactive",
    "gsd_research_enabled": true,
    "gsd_plan_check_enabled": true,
    "gsd_verifier_enabled": true,
    "ralph_max_iterations": 10,
    "ralph_auto_commit": true,
    "theme": "dark",
    "updated_at": "2026-02-08T18:00:00Z"
  }
}
```

**Note:** There is no `default_approach` setting. All tasks use:
- **GSD** for planning (research, requirements, roadmap)
- **Ralph Loop** for execution (iterate on stories until complete)

---

#### Update Settings

```http
PUT /api/v1/settings
```

**Request Body:**

```json
{
  "max_parallel_executions": 5,
  "ralph_max_iterations": 15,
  "gsd_depth": "comprehensive"
}
```

**Response:** `200 OK`

---

#### Test Connection

Test OpenClaw Gateway connectivity.

```http
POST /api/v1/settings/test-connection
```

**Response:**

```json
{
  "connected": true,
  "message": "Connected to OpenClaw Gateway"
}
```

**Error Response:**

```json
{
  "connected": false,
  "message": "Connection failed: timeout"
}
```

---

### Projects

#### List Projects

```http
GET /api/v1/projects
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status (active/completed/on-hold) |

**Response:**

```json
{
  "data": [
    {
      "id": "project-456",
      "name": "User Auth System",
      "description": "JWT-based authentication",
      "status": "active",
      "color": "#8b5cf6",
      "created_at": "2026-02-08T10:00:00Z",
      "updated_at": "2026-02-08T10:00:00Z"
    }
  ]
}
```

---

#### Create Project

```http
POST /api/v1/projects
```

**Request Body:**

```json
{
  "name": "User Auth System",
  "description": "JWT-based authentication",
  "status": "active",
  "color": "#8b5cf6"
}
```

**Response:** `201 Created`

---

#### Get Project

```http
GET /api/v1/projects/:id
```

Returns project with task count and done count.

---

#### Update Project

```http
PUT /api/v1/projects/:id
```

Partial updates supported (only non-empty fields updated).

---

#### Delete Project

```http
DELETE /api/v1/projects/:id
```

**Response:** `204 No Content`. Task `project_id` is set to NULL for affected tasks.

---

#### List Project Tasks

```http
GET /api/v1/projects/:id/tasks
```

**Response:** Array of tasks belonging to the project.

---

### Comments

#### List Task Comments

```http
GET /api/v1/tasks/:id/comments
```

**Response:**

```json
{
  "data": [
    {
      "id": "comment-1",
      "task_id": "task-123",
      "author": "user",
      "content": "Consider adding rate limiting here.",
      "created_at": "2026-02-09T14:00:00Z"
    }
  ]
}
```

---

#### Create Comment

```http
POST /api/v1/tasks/:id/comments
```

**Request Body:**

```json
{
  "author": "user",
  "content": "Consider adding rate limiting here."
}
```

**Response:** `201 Created`

---

#### Delete Comment

```http
DELETE /api/v1/comments/:id
```

**Response:** `204 No Content`

---

## WebSocket Events

**Endpoint:** `ws://localhost:8080/ws`

### Connection

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('Connected to Mission Control');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Event:', message);
};
```

### Event Format

```json
{
  "type": "event_type",
  "payload": { /* event data */ },
  "timestamp": "2026-02-08T22:50:00Z"
}
```

### Event Types

#### Agent Status Change

```json
{
  "type": "agent.status",
  "payload": {
    "agent_id": "jarvis",
    "status": "working",
    "current_task_id": "task-123"
  }
}
```

---

#### Task Status Change

```json
{
  "type": "task.status",
  "payload": {
    "task_id": "task-123",
    "status": "executing",
    "progress": 0.45
  }
}
```

---

#### Phase Updated

```json
{
  "type": "phase.updated",
  "payload": {
    "phase_id": "phase-2",
    "task_id": "task-123",
    "status": "executing",
    "progress": 0.65
  }
}
```

---

#### Story Updated

```json
{
  "type": "story.updated",
  "payload": {
    "story_id": "story-3",
    "task_id": "task-456",
    "passes": true,
    "iterations": 2
  }
}
```

---

#### New Event

```json
{
  "type": "event.new",
  "payload": {
    "id": "event-890",
    "task_id": "task-123",
    "type": "phase_completed",
    "message": "Phase 2 completed successfully"
  }
}
```

---

#### Execution Log

```json
{
  "type": "execution.log",
  "payload": {
    "task_id": "task-123",
    "level": "info",
    "message": "Installing npm dependencies...",
    "timestamp": "2026-02-08T22:52:00Z"
  }
}
```

**Log Levels:** `debug`, `info`, `warn`, `error`

---

#### Chat Session Started

```json
{
  "type": "chat.session.started",
  "payload": {
    "session_id": "chat-session-123",
    "agent_id": "jarvis",
    "session_key": "agent:jarvis:chat:chat-session-123",
    "started_at": "2026-02-09T20:00:00Z"
  }
}
```

---

#### Chat Session Ended

```json
{
  "type": "chat.session.ended",
  "payload": {
    "session_id": "chat-session-123",
    "agent_id": "jarvis",
    "ended_at": "2026-02-09T20:15:00Z",
    "duration_seconds": 900,
    "message_count": 8
  }
}
```

---

#### Chat Message Received

```json
{
  "type": "chat.message",
  "payload": {
    "message_id": "msg-5",
    "session_id": "chat-session-123",
    "agent_id": "jarvis",
    "role": "agent",
    "content": "The dashboard API task is currently in the executing phase...",
    "created_at": "2026-02-09T20:10:15Z"
  }
}
```

---

## Agent Self-Reporting

These endpoints are called by executing agents to report their progress.

### Update Phase Progress

```http
POST /api/v1/phases/:id/progress
```

**Request Body:**

```json
{
  "progress": 0.65,
  "message": "Implementing login endpoint..."
}
```

**Response:** `200 OK`

---

### Complete Phase

```http
POST /api/v1/phases/:id/complete
```

**Request Body:**

```json
{
  "summary": "Implemented JWT auth with login/logout endpoints",
  "artifacts": {
    "files_created": ["src/auth/login.go", "src/auth/logout.go"],
    "files_modified": ["src/routes.go"],
    "commit_sha": "abc123def"
  }
}
```

**Response:** `200 OK`

---

### Fail Phase

```http
POST /api/v1/phases/:id/fail
```

**Request Body:**

```json
{
  "error": "Database connection failed",
  "recoverable": true,
  "suggestion": "Check DATABASE_URL environment variable"
}
```

**Response:** `200 OK`

---

### Pass Story

```http
POST /api/v1/stories/:id/pass
```

**Request Body:**

```json
{
  "commit_sha": "abc123def",
  "learnings": "Used bcrypt for password hashing, integrated with existing user model"
}
```

**Response:** `200 OK`

---

### Fail Story

```http
POST /api/v1/stories/:id/fail
```

**Request Body:**

```json
{
  "error": "Test failed: expected 200, got 401",
  "iteration": 3,
  "stack_trace": "..."
}
```

**Response:** `200 OK`

---

### Append Progress Text

```http
POST /api/v1/tasks/:id/progress-txt
```

**Request Body:**

```json
{
  "content": "Iteration 3: Discovered that the auth middleware needs to be applied before the router..."
}
```

**Response:** `200 OK`

---

## Pagination

List endpoints support pagination with the following parameters:

| Parameter | Type | Default | Max |
|-----------|------|---------|-----|
| `page` | int | 1 | - |
| `per_page` | int | 20 | 100 |

Response includes pagination metadata:

```json
{
  "meta": {
    "total": 156,
    "page": 2,
    "per_page": 20,
    "total_pages": 8
  }
}
```

---

## Filtering & Sorting

Many list endpoints support filtering and sorting:

```http
GET /api/v1/tasks?status=executing&priority=1&sort=-created_at
```

Common filters:
- `status` - Filter by status
- `priority` - Filter by priority
- `agent_id` - Filter by agent

Sort:
- `sort` - Field to sort by (prefix with `-` for descending)
- Examples: `created_at`, `-updated_at`, `priority`

---

## Webhooks

**Future Feature:** Webhook support for external integrations

Will support:
- Task completion events
- Agent status changes
- Error notifications

---

## SDK/Client Libraries

**Future Feature:** Official client libraries

Planned:
- Go SDK
- TypeScript/JavaScript SDK
- Python SDK

---

## Changelog

See [CHANGELOG.md](../CHANGELOG.md) (project root) for API version history and breaking changes.
