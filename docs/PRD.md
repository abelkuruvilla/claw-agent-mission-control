# Product Requirements Document (PRD)
# Claw Agent Mission Control

**Version:** 1.0.0  
**Date:** 2026-02-08  
**Author:** Jarvis (AI Software Engineer)  
**Owner:** Abel Kuruvilla

---

## 1. Executive Summary

### 1.1 Vision
Claw Agent Mission Control is a production-grade dashboard for managing AI agents and tasks using OpenClaw Gateway. It implements the GSD (Get Shit Done) and Ralph Loop methodologies for autonomous, context-aware task execution with sub-agent orchestration.

### 1.2 Goals
- Provide a unified interface to create, manage, and monitor AI agents
- Enable task assignment with GSD/Ralph execution approaches
- Support hierarchical agent structures (main agents → sub-agents)
- Deliver a single Golang binary that serves both API and embedded frontend
- Integrate seamlessly with existing OpenClaw configurations

### 1.3 Non-Goals (v1.0)
- Multi-tenant/multi-user authentication (single-user system)
- Mobile-native applications
- Custom LLM provider management (uses OpenClaw's configured providers)
- Real-time collaborative editing

---

## 2. Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Claw Agent Mission Control                       │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                    Single Go Binary                              ││
│  │  ┌──────────────────────┐  ┌──────────────────────────────────┐ ││
│  │  │   Embedded NextJS    │  │         Go HTTP Server           │ ││
│  │  │   (Static Assets)    │  │  ┌────────────┐ ┌─────────────┐  │ ││
│  │  │                      │  │  │  REST API  │ │  WebSocket  │  │ ││
│  │  │  • Dashboard         │  │  │            │ │   Server    │  │ ││
│  │  │  • Agent Management  │  │  │  /api/v1/* │ │  /ws/*      │  │ ││
│  │  │  • Task Board        │  │  └────────────┘ └─────────────┘  │ ││
│  │  │  • Event Log         │  │         │              │         │ ││
│  │  └──────────────────────┘  │         ▼              ▼         │ ││
│  │           │                │  ┌─────────────────────────────┐ │ ││
│  │           │                │  │      Business Logic         │ │ ││
│  │           ▼                │  │  ┌─────────┐ ┌───────────┐  │ │ ││
│  │  ┌──────────────────────┐ │  │  │ GSD     │ │ Ralph     │  │ │ ││
│  │  │     fs.FS Embed      │ │  │  │ Engine  │ │ Engine    │  │ │ ││
│  │  │   (go:embed ui/out)  │ │  │  └─────────┘ └───────────┘  │ │ ││
│  │  └──────────────────────┘ │  └─────────────────────────────┘ │ ││
│  │                            │         │                        │ ││
│  │                            │         ▼                        │ ││
│  │                            │  ┌─────────────────────────────┐ │ ││
│  │                            │  │   OpenClaw Gateway Client   │ │ ││
│  │                            │  │   (WebSocket Connection)    │ │ ││
│  │                            │  └─────────────────────────────┘ │ ││
│  └────────────────────────────┴──────────────────────────────────┘ ││
│                                        │                            │
│                                        ▼                            │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                        SQLite Database                          ││
│  │   agents │ tasks │ phases │ stories │ events │ settings         ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      OpenClaw Gateway                                │
│                    (External Dependency)                             │
│  • Agent Sessions          • sessions_spawn                          │
│  • Task Dispatch           • sessions_send                           │
│  • Real-time Events        • WebSocket Events                        │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Technology Stack

| Layer | Technology | Justification |
|-------|------------|---------------|
| Backend | Go 1.22+ | Performance, single binary, embed support |
| HTTP Framework | Echo v4 | Lightweight, middleware support, WebSocket |
| Database | SQLite + sqlc | Embedded, zero-config, type-safe queries |
| Migrations | golang-migrate | Versioned schema management |
| Frontend | Next.js 14+ | App router, server components, modern React |
| Styling | Tailwind CSS | Utility-first, dark mode support |
| State | Zustand | Lightweight, TypeScript-friendly |
| Real-time | Native WebSocket | Live updates, event streaming |
| Build | Make + npm | Unified build pipeline |

### 2.3 Directory Structure

```
claw-agent-mission-control/
├── cmd/
│   └── server/
│       └── main.go              # Entry point
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP handlers
│   │   ├── middleware/          # Auth, logging, CORS
│   │   └── routes.go            # Route definitions
│   ├── config/
│   │   └── config.go            # Configuration loading
│   ├── db/
│   │   ├── migrations/          # SQL migrations
│   │   ├── queries/             # sqlc queries
│   │   └── sqlite.go            # DB connection
│   ├── executor/
│   │   ├── gsd.go               # GSD execution engine
│   │   ├── ralph.go             # Ralph loop engine
│   │   └── orchestrator.go      # Task orchestration
│   ├── models/
│   │   └── models.go            # Domain models
│   ├── openclaw/
│   │   └── client.go            # OpenClaw Gateway client
│   ├── store/
│   │   └── store.go             # Data access layer
│   └── websocket/
│       └── hub.go               # WebSocket hub
├── ui/
│   ├── src/
│   │   ├── app/                 # Next.js pages
│   │   ├── components/          # React components
│   │   └── lib/                 # Utilities
│   ├── public/
│   ├── tailwind.config.ts
│   ├── next.config.js
│   └── package.json
├── embed.go                     # go:embed directive
├── go.mod
├── go.sum
├── Makefile
├── .env.example
├── docker-compose.yml
├── Dockerfile
└── README.md
```

---

## 3. Data Models

### 3.1 Agent

```go
type Agent struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Type        string    `json:"type"`        // "main" | "sub"
    ParentID    *string   `json:"parent_id"`   // null for main agents
    Status      string    `json:"status"`      // "idle" | "working" | "paused" | "error"
    
    // Configuration
    SoulMD      string    `json:"soul_md"`     // SOUL.md content
    AgentsMD    string    `json:"agents_md"`   // AGENTS.md content
    IdentityMD  string    `json:"identity_md"` // IDENTITY.md content
    
    // OpenClaw Integration
    SessionKey  *string   `json:"session_key"` // Active session
    
    // Metadata
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    
    // Relations (loaded separately)
    CurrentTask *Task     `json:"current_task,omitempty"`
    SubAgents   []Agent   `json:"sub_agents,omitempty"`
}
```

### 3.2 Task

```go
type Task struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    AgentID     *string   `json:"agent_id"`    // Assigned agent
    
    // Execution Approach
    Approach    string    `json:"approach"`    // "gsd" | "ralph"
    
    // Status Tracking
    Status      string    `json:"status"`      
    // "backlog" | "planning" | "discussing" | "executing" | "verifying" | "review" | "done" | "failed"
    
    Priority    int       `json:"priority"`    // 1-5 (1 = highest)
    
    // GSD Artifacts
    ProjectMD      *string `json:"project_md"`
    RequirementsMD *string `json:"requirements_md"`
    RoadmapMD      *string `json:"roadmap_md"`
    StateMD        *string `json:"state_md"`
    
    // Ralph Artifacts
    PrdJSON       *string  `json:"prd_json"`
    ProgressTxt   *string  `json:"progress_txt"`
    
    // Execution Context
    WorkDir       string   `json:"work_dir"`   // Working directory
    GitBranch     *string  `json:"git_branch"`
    
    // Metadata
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    StartedAt     *time.Time `json:"started_at"`
    CompletedAt   *time.Time `json:"completed_at"`
    
    // Relations
    Phases        []Phase   `json:"phases,omitempty"`  // GSD
    Stories       []Story   `json:"stories,omitempty"` // Ralph
    Events        []Event   `json:"events,omitempty"`
}
```

### 3.3 Phase (GSD)

```go
type Phase struct {
    ID          string    `json:"id"`
    TaskID      string    `json:"task_id"`
    Sequence    int       `json:"sequence"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    
    Status      string    `json:"status"`
    // "pending" | "discussing" | "planning" | "executing" | "verifying" | "done" | "failed"
    
    // Artifacts
    ContextMD   *string   `json:"context_md"`
    ResearchMD  *string   `json:"research_md"`
    PlanMD      *string   `json:"plan_md"`
    SummaryMD   *string   `json:"summary_md"`
    UatMD       *string   `json:"uat_md"`
    
    // Verification
    VerificationResult *string `json:"verification_result"`
    
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### 3.4 Story (Ralph)

```go
type Story struct {
    ID          string    `json:"id"`
    TaskID      string    `json:"task_id"`
    Sequence    int       `json:"sequence"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    
    // Ralph Fields
    Passes      bool      `json:"passes"`
    Priority    int       `json:"priority"`
    
    // Acceptance Criteria (JSON array)
    AcceptanceCriteria []string `json:"acceptance_criteria"`
    
    // Execution
    Iterations  int       `json:"iterations"`   // Number of attempts
    LastError   *string   `json:"last_error"`
    
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### 3.5 Event

```go
type Event struct {
    ID        string    `json:"id"`
    TaskID    *string   `json:"task_id"`
    AgentID   *string   `json:"agent_id"`
    
    Type      string    `json:"type"`
    // "task_created" | "task_assigned" | "phase_started" | "phase_completed" |
    // "story_passed" | "story_failed" | "agent_spawned" | "execution_error" |
    // "verification_passed" | "verification_failed" | "commit_created"
    
    Message   string    `json:"message"`
    Details   *string   `json:"details"`  // JSON blob
    
    CreatedAt time.Time `json:"created_at"`
}
```

### 3.6 Settings

```go
type Settings struct {
    ID                    string `json:"id"`
    
    // OpenClaw Connection
    OpenClawGatewayURL    string `json:"openclaw_gateway_url"`
    OpenClawGatewayToken  string `json:"openclaw_gateway_token"`
    
    // Execution Defaults
    DefaultApproach       string `json:"default_approach"`     // "gsd" | "ralph"
    DefaultModel          string `json:"default_model"`
    MaxParallelExecutions int    `json:"max_parallel_executions"`
    
    // GSD Settings
    GSDDepth              string `json:"gsd_depth"`            // "quick" | "standard" | "comprehensive"
    GSDMode               string `json:"gsd_mode"`             // "yolo" | "interactive"
    GSDResearchEnabled    bool   `json:"gsd_research_enabled"`
    GSDPlanCheckEnabled   bool   `json:"gsd_plan_check_enabled"`
    GSDVerifierEnabled    bool   `json:"gsd_verifier_enabled"`
    
    // Ralph Settings
    RalphMaxIterations    int    `json:"ralph_max_iterations"`
    RalphAutoCommit       bool   `json:"ralph_auto_commit"`
    
    // UI Settings
    Theme                 string `json:"theme"`                // "dark" | "light" | "system"
    
    UpdatedAt             time.Time `json:"updated_at"`
}
```

---

## 4. API Specification

### 4.1 REST Endpoints

#### Agents

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/agents` | List all agents |
| GET | `/api/v1/agents/:id` | Get agent by ID |
| POST | `/api/v1/agents` | Create new agent |
| PUT | `/api/v1/agents/:id` | Update agent |
| DELETE | `/api/v1/agents/:id` | Delete agent |
| POST | `/api/v1/agents/:id/spawn-sub` | Spawn sub-agent |
| GET | `/api/v1/agents/:id/history` | Get agent task history |

#### Tasks

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/tasks` | List tasks (with filters) |
| GET | `/api/v1/tasks/:id` | Get task with phases/stories |
| POST | `/api/v1/tasks` | Create new task |
| PUT | `/api/v1/tasks/:id` | Update task |
| DELETE | `/api/v1/tasks/:id` | Delete task |
| POST | `/api/v1/tasks/:id/assign` | Assign task to agent |
| POST | `/api/v1/tasks/:id/start` | Start task execution |
| POST | `/api/v1/tasks/:id/pause` | Pause task execution |
| POST | `/api/v1/tasks/:id/resume` | Resume task execution |
| POST | `/api/v1/tasks/:id/cancel` | Cancel task |
| GET | `/api/v1/tasks/:id/events` | Get task events |

#### Phases (GSD)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/tasks/:id/phases` | List phases |
| POST | `/api/v1/tasks/:id/phases` | Add phase |
| PUT | `/api/v1/phases/:id` | Update phase |
| DELETE | `/api/v1/phases/:id` | Delete phase |
| POST | `/api/v1/phases/:id/discuss` | Start discussion |
| POST | `/api/v1/phases/:id/plan` | Generate plan |
| POST | `/api/v1/phases/:id/execute` | Execute phase |
| POST | `/api/v1/phases/:id/verify` | Verify phase |

#### Stories (Ralph)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/tasks/:id/stories` | List stories |
| POST | `/api/v1/tasks/:id/stories` | Add story |
| PUT | `/api/v1/stories/:id` | Update story |
| DELETE | `/api/v1/stories/:id` | Delete story |
| POST | `/api/v1/stories/:id/execute` | Execute story |

#### Events

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/events` | List all events (paginated) |
| GET | `/api/v1/events/stream` | SSE event stream |

#### Settings

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/settings` | Get settings |
| PUT | `/api/v1/settings` | Update settings |
| POST | `/api/v1/settings/test-connection` | Test OpenClaw connection |

#### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/status` | System status with OpenClaw connection |

### 4.2 WebSocket Events

**Connection:** `ws://localhost:8080/ws`

#### Server → Client Events

```typescript
// Agent status change
{
  "type": "agent.status",
  "payload": {
    "agent_id": "abc123",
    "status": "working",
    "current_task_id": "task123"
  }
}

// Task status change
{
  "type": "task.status",
  "payload": {
    "task_id": "task123",
    "status": "executing",
    "progress": 0.45
  }
}

// Phase/Story update
{
  "type": "phase.updated",
  "payload": { /* Phase object */ }
}

// New event
{
  "type": "event.new",
  "payload": { /* Event object */ }
}

// Execution log
{
  "type": "execution.log",
  "payload": {
    "task_id": "task123",
    "level": "info",
    "message": "Executing phase 1...",
    "timestamp": "2026-02-08T22:30:00Z"
  }
}
```

---

## 5. User Interface

### 5.1 Design Principles

- **Dark Mode First** - Default to dark theme (can toggle to light)
- **Minimal & Clean** - Asana-inspired, no visual clutter
- **Information Density** - Show relevant data without overwhelming
- **Real-time Updates** - WebSocket-powered live status
- **Keyboard Shortcuts** - Power-user friendly

### 5.2 Color Palette (Dark Mode)

```css
--bg-primary: #0d1117;      /* Main background */
--bg-secondary: #161b22;    /* Cards, panels */
--bg-tertiary: #21262d;     /* Hover states */
--border: #30363d;          /* Borders */
--text-primary: #f0f6fc;    /* Primary text */
--text-secondary: #8b949e;  /* Secondary text */
--accent-blue: #58a6ff;     /* Primary actions */
--accent-green: #3fb950;    /* Success states */
--accent-yellow: #d29922;   /* Warnings */
--accent-red: #f85149;      /* Errors */
--accent-purple: #a371f7;   /* Special/AI */
```

### 5.3 Pages & Components

#### 5.3.1 Dashboard (`/`)

```
┌─────────────────────────────────────────────────────────────────────┐
│  🎮 Mission Control                              [Settings] [?]      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ STATS                                                        │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │   │
│  │  │ 3 Active │ │ 12 Tasks │ │ 8 Done   │ │ 2 Failed │        │   │
│  │  │ Agents   │ │ In Queue │ │ Today    │ │ This Week│        │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘        │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌───────────────────────────────┐ ┌───────────────────────────┐   │
│  │ ACTIVE AGENTS                 │ │ RECENT EVENTS             │   │
│  │                               │ │                           │   │
│  │ ┌───────────────────────────┐ │ │ ● Task "API" completed    │   │
│  │ │ 🤖 Jarvis        Working  │ │ │   2 minutes ago           │   │
│  │ │    Building API endpoint  │ │ │                           │   │
│  │ │    ████████░░ 80%        │ │ │ ● Phase 3 started         │   │
│  │ └───────────────────────────┘ │ │   5 minutes ago           │   │
│  │                               │ │                           │   │
│  │ ┌───────────────────────────┐ │ │ ● Agent "Builder" spawned │   │
│  │ │ 🔧 Builder        Working  │ │ │   10 minutes ago          │   │
│  │ │    Sub-agent of Jarvis    │ │ │                           │   │
│  │ │    Running tests          │ │ │ ● Story 2/5 passed        │   │
│  │ └───────────────────────────┘ │ │   15 minutes ago          │   │
│  │                               │ │                           │   │
│  └───────────────────────────────┘ └───────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

#### 5.3.2 Agents Page (`/agents`)

```
┌─────────────────────────────────────────────────────────────────────┐
│  Agents                                    [+ New Agent]             │
├─────────────────────────────────────────────────────────────────────┤
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │ Search agents...                          [Filter ▼] [Sort ▼] │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                                                                 ││
│  │  ┌─────────────────────────────────────────────────────────┐   ││
│  │  │ 🤖 Jarvis                                    Main Agent │   ││
│  │  │                                                         │   ││
│  │  │ Personal AI Software Engineer                           │   ││
│  │  │                                                         │   ││
│  │  │ Status: ● Working    Current: Building dashboard API    │   ││
│  │  │ Tasks: 5 completed, 2 in progress                       │   ││
│  │  │                                                         │   ││
│  │  │ Sub-agents:                                             │   ││
│  │  │   └─ 🔧 Builder (working)                               │   ││
│  │  │   └─ 🧪 Tester (idle)                                   │   ││
│  │  │                                                         │   ││
│  │  │ [View] [Edit] [Assign Task]                            │   ││
│  │  └─────────────────────────────────────────────────────────┘   ││
│  │                                                                 ││
│  │  ┌─────────────────────────────────────────────────────────┐   ││
│  │  │ 🎨 Designer                                  Main Agent │   ││
│  │  │                                                         │   ││
│  │  │ UI/UX Design Specialist                                 │   ││
│  │  │                                                         │   ││
│  │  │ Status: ○ Idle        Last active: 2 hours ago          │   ││
│  │  │ Tasks: 12 completed                                     │   ││
│  │  │                                                         │   ││
│  │  │ [View] [Edit] [Assign Task]                            │   ││
│  │  └─────────────────────────────────────────────────────────┘   ││
│  │                                                                 ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
```

#### 5.3.3 Agent Detail Modal/Page

```
┌─────────────────────────────────────────────────────────────────────┐
│  Agent: Jarvis                                            [X Close] │
├─────────────────────────────────────────────────────────────────────┤
│  [Details] [Configuration] [History] [Sub-agents]                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Name: Jarvis                                                       │
│  Type: Main Agent                                                   │
│  Status: ● Working                                                  │
│                                                                     │
│  ┌─ SOUL.md ───────────────────────────────────────────────────┐   │
│  │ # SOUL.md - Who You Are                                     │   │
│  │                                                             │   │
│  │ You're not a chatbot. You're becoming someone...            │   │
│  │                                                     [Edit]  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ AGENTS.md ─────────────────────────────────────────────────┐   │
│  │ # AGENTS.md - Your Workspace                                │   │
│  │                                                             │   │
│  │ This folder is home. Treat it that way...                   │   │
│  │                                                     [Edit]  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌─ IDENTITY.md ───────────────────────────────────────────────┐   │
│  │ # IDENTITY.md                                               │   │
│  │ - Name: Jarvis                                              │   │
│  │ - Role: AI Software Engineer                                │   │
│  │                                                     [Edit]  │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                     │
│                                      [Save Changes] [Delete Agent]  │
└─────────────────────────────────────────────────────────────────────┘
```

#### 5.3.4 Tasks Page (`/tasks`) - Kanban Board

```
┌─────────────────────────────────────────────────────────────────────┐
│  Tasks                           [+ New Task]  [List View] [Board]  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│ │ BACKLOG  │ │ PLANNING │ │EXECUTING │ │ REVIEW   │ │   DONE   │   │
│ │    (5)   │ │    (2)   │ │    (3)   │ │    (1)   │ │   (12)   │   │
│ ├──────────┤ ├──────────┤ ├──────────┤ ├──────────┤ ├──────────┤   │
│ │          │ │          │ │          │ │          │ │          │   │
│ │┌────────┐│ │┌────────┐│ │┌────────┐│ │┌────────┐│ │┌────────┐│   │
│ ││Build   ││ ││API     ││ ││Dashboard││ ││Auth    ││ ││Setup   ││   │
│ ││Payment ││ ││Docs    ││ ││UI      ││ ││Flow    ││ ││Project ││   │
│ ││        ││ ││        ││ ││        ││ ││        ││ ││        ││   │
│ ││GSD     ││ ││Ralph   ││ ││GSD     ││ ││Ralph   ││ ││GSD     ││   │
│ ││P1      ││ ││P2      ││ ││P1      ││ ││P2      ││ ││        ││   │
│ │└────────┘│ │└────────┘│ ││████░░  ││ │└────────┘│ │└────────┘│   │
│ │          │ │          │ ││ 60%    ││ │          │ │          │   │
│ │┌────────┐│ │┌────────┐│ │└────────┘│ │          │ │┌────────┐│   │
│ ││Mobile  ││ ││        ││ │          │ │          │ ││Database││   │
│ ││App     ││ ││        ││ │┌────────┐│ │          │ ││Schema  ││   │
│ ││        ││ ││        ││ ││Testing ││ │          │ │└────────┘│   │
│ ││Ralph   ││ ││        ││ ││Suite   ││ │          │ │          │   │
│ ││P3      ││ ││        ││ ││████████││ │          │ │          │   │
│ │└────────┘│ │└────────┘│ │└────────┘│ │          │ │          │   │
│ │          │ │          │ │          │ │          │ │          │   │
│ └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

#### 5.3.5 Task Detail View

```
┌─────────────────────────────────────────────────────────────────────┐
│  Task: Build Dashboard UI                                  [X Close]│
├─────────────────────────────────────────────────────────────────────┤
│  Status: ● Executing (Phase 3/5)        Approach: GSD              │
│  Agent: 🤖 Jarvis                       Priority: P1               │
│  Started: 2h 30m ago                    Progress: ████████░░ 60%   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  [Overview] [Phases] [Artifacts] [Events]                          │
│                                                                     │
│  ┌─ PHASES ────────────────────────────────────────────────────────┐│
│  │                                                                 ││
│  │  ✓ Phase 1: Project Setup                         [Completed]  ││
│  │     └─ Created Next.js project structure                       ││
│  │                                                                 ││
│  │  ✓ Phase 2: Database Layer                        [Completed]  ││
│  │     └─ SQLite + sqlc integration                               ││
│  │                                                                 ││
│  │  ● Phase 3: REST API                              [Executing]  ││
│  │     └─ Implementing CRUD endpoints                             ││
│  │     └─ Progress: ████████░░ 80%                                ││
│  │     └─ Current: Writing task handlers                          ││
│  │                                                                 ││
│  │  ○ Phase 4: WebSocket Integration                 [Pending]    ││
│  │     └─ Real-time updates                                       ││
│  │                                                                 ││
│  │  ○ Phase 5: UI Components                         [Pending]    ││
│  │     └─ React components + Tailwind                             ││
│  │                                                                 ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
│  [Pause] [Cancel]                                                   │
└─────────────────────────────────────────────────────────────────────┘
```

#### 5.3.6 New Task Modal

```
┌─────────────────────────────────────────────────────────────────────┐
│  Create New Task                                          [X Close] │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  Title *                                                            │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │ Build user authentication system                                ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
│  Description                                                        │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │ Implement JWT-based authentication with login, register,        ││
│  │ password reset, and session management.                         ││
│  │                                                                 ││
│  │                                                                 ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
│  Execution Approach                                                 │
│  ┌───────────────────────┐ ┌───────────────────────┐               │
│  │ ● GSD                 │ │ ○ Ralph               │               │
│  │   Multi-phase with    │ │   Autonomous loop     │               │
│  │   research & planning │ │   until all pass      │               │
│  └───────────────────────┘ └───────────────────────┘               │
│                                                                     │
│  Assign to Agent                                                    │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │ 🤖 Jarvis                                                   ▼  ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
│  Priority                                                           │
│  ○ P1 (Critical)  ● P2 (High)  ○ P3 (Medium)  ○ P4 (Low)           │
│                                                                     │
│  Working Directory                                                  │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │ ~/projects/my-app                                               ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
│                                            [Cancel] [Create Task]   │
└─────────────────────────────────────────────────────────────────────┘
```

#### 5.3.7 Settings Page (`/settings`)

```
┌─────────────────────────────────────────────────────────────────────┐
│  Settings                                                           │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─ OpenClaw Connection ───────────────────────────────────────────┐│
│  │                                                                 ││
│  │  Gateway URL                                                    ││
│  │  ┌─────────────────────────────────────────────────────────────┐││
│  │  │ ws://127.0.0.1:18789                                        │││
│  │  └─────────────────────────────────────────────────────────────┘││
│  │                                                                 ││
│  │  Gateway Token                                                  ││
│  │  ┌─────────────────────────────────────────────────────────────┐││
│  │  │ ••••••••••••••••••••                            [Show]      │││
│  │  └─────────────────────────────────────────────────────────────┘││
│  │                                                                 ││
│  │  Status: ● Connected                      [Test Connection]     ││
│  │                                                                 ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
│  ┌─ Execution Defaults ────────────────────────────────────────────┐│
│  │                                                                 ││
│  │  Default Approach:     ● GSD    ○ Ralph                         ││
│  │  Max Parallel:         [3]                                      ││
│  │  Default Model:        [anthropic/claude-sonnet-4-5      ▼]     ││
│  │                                                                 ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
│  ┌─ GSD Settings ──────────────────────────────────────────────────┐│
│  │                                                                 ││
│  │  Depth:            ○ Quick  ● Standard  ○ Comprehensive         ││
│  │  Mode:             ○ Interactive  ● YOLO (Auto-approve)         ││
│  │  Research:         [✓] Enabled                                  ││
│  │  Plan Checker:     [✓] Enabled                                  ││
│  │  Verifier:         [✓] Enabled                                  ││
│  │                                                                 ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
│  ┌─ Ralph Settings ────────────────────────────────────────────────┐│
│  │                                                                 ││
│  │  Max Iterations:   [10]                                         ││
│  │  Auto Commit:      [✓] Enabled                                  ││
│  │                                                                 ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
│  ┌─ UI ────────────────────────────────────────────────────────────┐│
│  │                                                                 ││
│  │  Theme:            ● Dark  ○ Light  ○ System                    ││
│  │                                                                 ││
│  └─────────────────────────────────────────────────────────────────┘│
│                                                                     │
│                                                    [Save Settings]  │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.4 Components Library

| Component | Description |
|-----------|-------------|
| `AgentCard` | Displays agent info, status, current task |
| `AgentForm` | Create/edit agent with markdown editors |
| `TaskCard` | Kanban task card with status badge |
| `TaskForm` | Create/edit task modal |
| `TaskDetail` | Full task view with phases/stories |
| `PhaseList` | GSD phases with progress indicators |
| `StoryList` | Ralph stories with pass/fail status |
| `EventLog` | Real-time event stream sidebar |
| `StatusBadge` | Colored status indicator |
| `ProgressBar` | Task/phase progress visualization |
| `MarkdownEditor` | Monaco-based markdown editor |
| `ConfirmDialog` | Confirmation modal |
| `Toast` | Notification toasts |
| `Sidebar` | Navigation sidebar |
| `Header` | Top navigation bar |

---

## 6. Execution Engines

### 6.1 GSD Engine

The GSD (Get Shit Done) engine implements spec-driven development with fresh context per phase.

#### Workflow

```
1. Task Created
   │
   ▼
2. Initialization (Auto)
   ├─ Questions → Build PROJECT.md
   ├─ Research → Domain knowledge
   ├─ Requirements → REQUIREMENTS.md
   └─ Roadmap → ROADMAP.md, Phases
   │
   ▼
3. For Each Phase:
   │
   ├─► Discuss Phase
   │   └─ AI asks clarifying questions
   │   └─ Creates CONTEXT.md
   │
   ├─► Plan Phase
   │   └─ Research implementation
   │   └─ Create atomic PLAN.md files
   │   └─ Verify plans
   │
   ├─► Execute Phase
   │   └─ Spawn fresh sub-agent session
   │   └─ Run plans in parallel waves
   │   └─ Atomic commits per task
   │   └─ Update STATE.md
   │
   └─► Verify Phase
       └─ Check deliverables
       └─ Create SUMMARY.md
       └─ Mark phase done or create fix plans
   │
   ▼
4. All Phases Complete → Task Done
```

#### Implementation

```go
type GSDEngine struct {
    openclawClient *openclaw.Client
    store          *store.Store
    hub            *websocket.Hub
}

func (e *GSDEngine) ExecutePhase(ctx context.Context, phase *models.Phase) error {
    // 1. Spawn fresh sub-agent session
    session, err := e.openclawClient.SpawnSession(ctx, &openclaw.SpawnRequest{
        Task:    phase.BuildPrompt(),
        AgentID: phase.Task.Agent.ID,
        Label:   fmt.Sprintf("phase-%s", phase.ID),
    })
    
    // 2. Monitor execution via WebSocket
    // 3. Update phase status
    // 4. Collect artifacts
    // 5. Create verification
}
```

### 6.2 Ralph Engine

The Ralph engine implements autonomous looping until all PRD items pass.

#### Workflow

```
1. Task Created with prd.json
   │
   ▼
2. Ralph Loop:
   │
   ├─► Check: All stories pass?
   │   └─ YES → Complete, exit loop
   │   └─ NO → Continue
   │
   ├─► Select highest priority story where passes=false
   │
   ├─► Spawn fresh agent session
   │   └─ Context: prd.json + progress.txt + git history
   │   └─ Task: Implement this single story
   │
   ├─► Agent works:
   │   └─ Implement story
   │   └─ Run quality checks (typecheck, tests)
   │   └─ If checks pass → Commit
   │   └─ Update prd.json (passes: true)
   │   └─ Append learnings to progress.txt
   │
   └─► Iteration complete → Loop back
   │
   ▼
3. Max iterations reached OR all stories pass → Task Complete
```

#### Implementation

```go
type RalphEngine struct {
    openclawClient *openclaw.Client
    store          *store.Store
    hub            *websocket.Hub
    maxIterations  int
}

func (e *RalphEngine) Run(ctx context.Context, task *models.Task) error {
    for iteration := 0; iteration < e.maxIterations; iteration++ {
        // 1. Check if all stories pass
        if e.allStoriesPass(task) {
            return e.markComplete(ctx, task)
        }
        
        // 2. Get next story
        story := e.getNextStory(task)
        
        // 3. Spawn fresh session
        result, err := e.openclawClient.SpawnSession(ctx, &openclaw.SpawnRequest{
            Task:    e.buildPrompt(task, story),
            AgentID: task.Agent.ID,
            Label:   fmt.Sprintf("ralph-%s-iter-%d", task.ID, iteration),
        })
        
        // 4. Update story status based on result
        // 5. Append to progress.txt
        // 6. Broadcast update
    }
    
    return ErrMaxIterationsReached
}
```

### 6.3 Sub-Agent Spawning

Both engines can spawn sub-agents for parallel execution:

```go
func (e *Orchestrator) SpawnSubAgent(ctx context.Context, parent *models.Agent, task string) (*models.Agent, error) {
    // 1. Create sub-agent record
    subAgent := &models.Agent{
        ID:       uuid.New().String(),
        Name:     fmt.Sprintf("%s-worker-%d", parent.Name, count),
        Type:     "sub",
        ParentID: &parent.ID,
        SoulMD:   parent.SoulMD,   // Inherit config
        AgentsMD: parent.AgentsMD,
    }
    
    // 2. Register with OpenClaw
    session, err := e.openclawClient.SpawnSession(ctx, &openclaw.SpawnRequest{
        Task:    task,
        AgentID: parent.ID,  // Spawn under parent
    })
    
    subAgent.SessionKey = &session.Key
    
    // 3. Save and return
    return e.store.CreateAgent(ctx, subAgent)
}
```

---

## 7. Configuration

### 7.1 Environment Variables

```bash
# .env.example

# Server Configuration
PORT=8080
HOST=0.0.0.0
ENV=development  # development | production

# Database
DATABASE_PATH=./data/mission-control.db

# OpenClaw Gateway
OPENCLAW_GATEWAY_URL=ws://127.0.0.1:18789
OPENCLAW_GATEWAY_TOKEN=your-token-here

# Optional: Read from OpenClaw config
OPENCLAW_CONFIG_PATH=~/.openclaw/openclaw.json

# Execution Defaults
DEFAULT_APPROACH=gsd
DEFAULT_MODEL=anthropic/claude-sonnet-4-5
MAX_PARALLEL_EXECUTIONS=3

# GSD Defaults
GSD_DEPTH=standard
GSD_MODE=interactive
GSD_RESEARCH_ENABLED=true
GSD_PLAN_CHECK_ENABLED=true
GSD_VERIFIER_ENABLED=true

# Ralph Defaults
RALPH_MAX_ITERATIONS=10
RALPH_AUTO_COMMIT=true

# UI
THEME=dark
```

### 7.2 Configuration Loading Priority

1. Environment variables (highest)
2. `.env` file in working directory
3. OpenClaw config (`~/.openclaw/openclaw.json`) for gateway URL/token
4. Defaults (lowest)

---

## 8. Build & Deployment

### 8.1 Build Process

```makefile
# Makefile

.PHONY: all build clean dev

# Build frontend and embed into Go binary
all: build

# Development mode with hot reload
dev:
	@echo "Starting frontend dev server..."
	cd ui && npm run dev &
	@echo "Starting backend with Air..."
	air

# Production build
build: build-ui build-server

build-ui:
	cd ui && npm ci && npm run build

build-server: build-ui
	go build -o bin/mission-control ./cmd/server

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf ui/.next
	rm -rf ui/out

# Run production binary
run: build
	./bin/mission-control

# Docker build
docker:
	docker build -t claw-agent-mission-control .
```

### 8.2 Embedding Frontend

```go
// embed.go
package main

import (
    "embed"
    "io/fs"
)

//go:embed ui/out/*
var embeddedUI embed.FS

func UIAssets() (fs.FS, error) {
    return fs.Sub(embeddedUI, "ui/out")
}
```

### 8.3 Dockerfile

```dockerfile
# Build stage - Frontend
FROM node:20-alpine AS frontend
WORKDIR /app/ui
COPY ui/package*.json ./
RUN npm ci
COPY ui/ ./
RUN npm run build

# Build stage - Backend
FROM golang:1.22-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/ui/out ./ui/out
RUN CGO_ENABLED=1 go build -o mission-control ./cmd/server

# Runtime stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite
WORKDIR /app
COPY --from=backend /app/mission-control .
EXPOSE 8080
CMD ["./mission-control"]
```

### 8.4 Running

```bash
# Development
make dev

# Production (single binary)
./mission-control

# With custom config
PORT=3000 OPENCLAW_GATEWAY_URL=ws://192.168.1.100:18789 ./mission-control

# Docker
docker run -p 8080:8080 \
  -e OPENCLAW_GATEWAY_URL=ws://host.docker.internal:18789 \
  -e OPENCLAW_GATEWAY_TOKEN=your-token \
  -v ./data:/app/data \
  claw-agent-mission-control
```

---

## 9. Testing Strategy

### 9.1 Backend Tests

| Type | Coverage Target | Tools |
|------|-----------------|-------|
| Unit | 80% | Go testing, testify |
| Integration | Key flows | testcontainers-go |
| API | All endpoints | httptest |

### 9.2 Frontend Tests

| Type | Coverage Target | Tools |
|------|-----------------|-------|
| Unit | Components | Jest, React Testing Library |
| E2E | Critical paths | Playwright |

### 9.3 Test Commands

```bash
# Backend tests
go test ./...

# Frontend tests
cd ui && npm test

# E2E tests
cd ui && npm run test:e2e

# All tests
make test
```

---

## 10. Security Considerations

### 10.1 Authentication
- v1.0: Single-user, no authentication required
- Future: Optional basic auth or API key

### 10.2 Data Protection
- SQLite database in configurable location
- Sensitive tokens stored only in environment/config
- No cloud storage or external data transmission

### 10.3 OpenClaw Integration
- Token-based authentication to gateway
- WebSocket connection over localhost by default
- Support for TLS in production deployments

---

## 11. Milestones & Timeline

### Phase 1: Foundation (MVP)
**Duration:** 1-2 weeks

- [ ] Project setup (Go + Next.js structure)
- [ ] Database schema and migrations
- [ ] Basic REST API (CRUD for agents, tasks)
- [ ] OpenClaw Gateway client
- [ ] Frontend scaffolding with Tailwind
- [ ] Basic dashboard page
- [ ] Agent list and create/edit

### Phase 2: Task Management
**Duration:** 1-2 weeks

- [ ] Kanban board UI
- [ ] Task CRUD and assignment
- [ ] Phase/Story data models
- [ ] Task detail views
- [ ] Drag-and-drop functionality

### Phase 3: Execution Engines
**Duration:** 2-3 weeks

- [ ] GSD Engine implementation
- [ ] Ralph Engine implementation
- [ ] Sub-agent spawning
- [ ] WebSocket real-time updates
- [ ] Event logging

### Phase 4: Polish & Production
**Duration:** 1 week

- [ ] Error handling and recovery
- [ ] Settings page
- [ ] Docker support
- [ ] Documentation
- [ ] Testing

---

## 12. Success Metrics

| Metric | Target |
|--------|--------|
| Task completion rate | > 85% |
| Average task execution time | Track baseline |
| Sub-agent spawn success | > 95% |
| WebSocket connection stability | > 99% uptime |
| UI responsiveness | < 100ms for interactions |
| Build time (full) | < 3 minutes |
| Binary size | < 50MB |

---

## 13. Future Considerations (v2.0+)

- Multi-user authentication
- Team collaboration features
- Task templates
- Custom execution workflows
- Integration with external tools (Jira, Linear)
- Metrics and analytics dashboard
- Mobile-responsive design
- Plugin system for custom engines
- AI-powered task suggestions

---

## 14. Glossary

| Term | Definition |
|------|------------|
| GSD | Get Shit Done - spec-driven development methodology |
| Ralph | Autonomous loop pattern for iterative task completion |
| Phase | A stage in GSD workflow (discuss → plan → execute → verify) |
| Story | A single item in Ralph's prd.json |
| Main Agent | Top-level OpenClaw agent |
| Sub-Agent | Agent spawned by main agent for parallel work |
| OpenClaw Gateway | The AI runtime that executes agent sessions |

---

## 15. Appendix

### A. Reference Implementations

- [Temporal UI Server](https://github.com/temporalio/ui-server) - Go + embedded frontend pattern
- [crshdn/mission-control](https://github.com/crshdn/mission-control) - OpenClaw task dashboard
- [GSD](https://github.com/glittercowboy/get-shit-done) - Context engineering methodology
- [Ralph](https://github.com/snarktank/ralph) - Autonomous agent loop

### B. OpenClaw API Reference

Key functions used:
- `sessions_spawn` - Create isolated agent session
- `sessions_send` - Send message to session
- `sessions_list` - List active sessions
- `sessions_history` - Get session history

---

*This PRD is a living document and will be updated as requirements evolve.*
