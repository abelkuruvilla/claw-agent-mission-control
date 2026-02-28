# Changelog

All notable changes to Claw Agent Mission Control will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2026-02-08

### 🎉 Initial Release

The first production-ready version of Claw Agent Mission Control!

### Added

#### Core Features

- **Orchestrator Agent Management**
  - Create, read, update, delete (CRUD) operations for agents
  - Full OpenClaw integration (creates workspace, agent directory, config entry)
  - Agent identity file editing (SOUL.md, IDENTITY.md, AGENTS.md, USER.md, TOOLS.md, HEARTBEAT.md)
  - Agent status tracking (idle, working, paused, error)
  - Real-time monitoring of agent activity

- **Task Management**
  - Asana-style Kanban board with drag-and-drop (UI)
  - Task lifecycle: Backlog → Planning → Executing → Review → Done
  - Priority levels (1-5)
  - Approach selection (GSD vs Ralph)
  - Task assignment to agents
  - Git branch tracking
  - Working directory configuration

- **GSD Execution Engine**
  - Multi-phase workflow: Research → Planning → Execution → Verification
  - Configurable depth (quick/standard/comprehensive)
  - Interactive and YOLO modes
  - Optional workflow agents (research, plan check, verifier)
  - Artifact storage (CONTEXT.md, RESEARCH.md, PLAN.md, SUMMARY.md, UAT.md)
  - Progress reporting API

- **Ralph Loop Engine**
  - PRD-driven story execution
  - Iterative retry until acceptance criteria pass
  - Max iteration safety limits
  - Auto-commit on successful stories
  - Progress tracking in progress.txt
  - Story pass/fail reporting

- **Sub-Agent Orchestration**
  - Spawn ephemeral sub-agents for parallel work
  - Fresh context per sub-agent (prevents context rot)
  - Automatic session cleanup
  - Sub-agent status monitoring

- **Real-Time Updates**
  - WebSocket server for live event streaming
  - Push notifications for task/agent/phase changes
  - Execution log streaming
  - Hub-based broadcast system

- **Event System**
  - Comprehensive event tracking
  - Event types: task lifecycle, phase changes, story results, errors
  - Queryable event log with filtering
  - Timestamps and detailed metadata

#### API

- **REST API (JSON)**
  - `/api/v1/agents` - Agent CRUD operations
  - `/api/v1/tasks` - Task management
  - `/api/v1/phases` - GSD phase operations
  - `/api/v1/stories` - Ralph story operations
  - `/api/v1/events` - Event log access
  - `/api/v1/settings` - System settings
  - `/health` - Health check endpoint
  - `/api/v1/status` - System status

- **WebSocket API**
  - `/ws` - Real-time event stream
  - Event types: agent.status, task.status, phase.updated, story.updated, event.new, execution.log

- **Agent Self-Reporting**
  - Phase progress reporting
  - Phase completion/failure reporting
  - Story pass/fail reporting
  - Progress text appending

#### Frontend (UI)

- **Dashboard**
  - System stats overview
  - Active agents panel
  - Recent events feed
  - Real-time status updates

- **Agents Page**
  - Agent list with search and filtering
  - Agent detail view
  - Configuration file editor
  - Memory viewer (MEMORY.md)
  - Sub-agent list

- **Tasks Board**
  - Kanban-style board (UI wireframes complete)
  - Task cards with progress indicators
  - Filtering and sorting
  - Task detail modal

- **Settings Page**
  - OpenClaw connection settings
  - Execution defaults
  - GSD configuration
  - Ralph configuration
  - UI preferences

- **Dark Mode**
  - Dark theme by default
  - Light mode toggle support
  - System preference detection

#### Infrastructure

- **Single Binary Deployment**
  - Embedded Next.js frontend in Go binary
  - No external dependencies except SQLite
  - ~14MB binary size
  - Cross-platform support

- **Database**
  - SQLite with WAL mode
  - Type-safe queries with sqlc
  - Migration system (golang-migrate)
  - Foreign key constraints

- **Build System**
  - Makefile with comprehensive targets
  - Multi-stage Docker builds
  - Development hot reload (Air + Next.js dev server)
  - Production optimization

- **Docker Support**
  - Dockerfile with multi-stage build
  - docker-compose.yml for easy deployment
  - Health checks
  - Volume mounting for data persistence

#### Documentation

- **README.md** - Complete project overview with quick start
- **BUILD.md** - Build and deployment guide
- **docs/PRD.md** - Product requirements document
- **docs/DESIGN_SPEC.md** - Architecture and design decisions
- **docs/API.md** - Complete API reference
- **CHANGELOG.md** - Version history (this file)
- **.env.example** - Comprehensive environment variable documentation

#### Configuration

- Environment-based configuration
- Support for `.env` files
- Auto-detection of OpenClaw config
- Sensible defaults for all settings
- Development vs production modes

### Technical Details

#### Stack

- **Backend**: Go 1.22+, Echo v4, SQLite, sqlc
- **Frontend**: Next.js 14, React 18, Tailwind CSS, Zustand
- **Real-time**: Native WebSocket
- **Build**: Make, npm, Docker multi-stage

#### Architecture Patterns

- Embedded filesystem (fs.FS) for UI assets
- WebSocket hub for broadcast messaging
- Repository pattern for data access
- Handler-based API routing
- Type-safe database queries

#### Security

- CORS middleware
- Request recovery
- Input validation
- SQL injection prevention (prepared statements)
- Environment variable isolation

### Known Limitations

- **Single-user system** - No authentication or multi-tenancy (planned for v2.0)
- **No rate limiting** - API is open to all requests (planned for v1.1)
- **Partial UI implementation** - Some screens are wireframes only (in progress)
- **No test coverage** - Unit tests to be added in v1.1
- **Limited error handling** - Some edge cases not yet covered

### Dependencies

#### Backend
- `github.com/labstack/echo/v4` - HTTP framework
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/golang-migrate/migrate/v4` - Database migrations
- Other standard Go libraries

#### Frontend
- `next` ^14.0.0 - React framework
- `react` ^18.0.0 - UI library
- `tailwindcss` - Styling
- `zustand` - State management

### Migration Notes

**First Installation:**
- Run `make build` to create the binary
- Copy `.env.example` to `.env` and configure
- Run `./bin/mission-control` to start
- Database will be created automatically

**Future Upgrades:**
- Database migrations will run automatically on startup
- Breaking changes will be documented in this file

---

## [1.0.1] - 2026-02-09

### 🔧 API Fixes, UI Rebuild & Identity Generation

### Added

- **Agent Identity Auto-Generation** - New feature that automatically generates agent identity files based on description
  - Uses GSD (Get Shit Done) and Ralph Loop principles
  - Generates SOUL.md, IDENTITY.md, AGENTS.md, USER.md, TOOLS.md, HEARTBEAT.md, MEMORY.md
  - Incorporates structured workflow (Research → Planning → Execution → Verification)
  - Includes autonomous iteration principles and context management
  - Falls back to explicit identity files if provided

### Fixed

- **Events API** - Implemented `GET /api/v1/events` endpoint that was returning 501 Not Implemented
  - Supports query params: `task_id`, `agent_id`, `limit`
  - Returns paginated event list with proper formatting

- **Settings API** - Implemented `GET /api/v1/settings` endpoint
  - Returns current settings or defaults if not configured
  - Proper handling of nullable database fields

- **Test Connection API** - Added `POST /api/v1/settings/test-connection` endpoint
  - Allows UI to verify OpenClaw Gateway connectivity
  - Returns connection status and message

### Changed

- **UI Rebuilt from Scratch** - Complete frontend rewrite using Figma design reference
  - NextJS 14 with App Router
  - ShadCN component library (17 components)
  - react-dnd for drag-and-drop Kanban
  - Zustand for state management
  - Dark theme by default (#0d1117 background)

- **Modal-based Detail Views** - Changed from separate pages to slide-out sheets
  - Agent details now shown in side sheet (AgentDetailSheet)
  - Project details now shown in side sheet (ProjectDetailSheet)
  - Enables static export for Go binary embedding

- **Static Export** - Re-enabled `output: 'export'` in next.config.ts
  - Required for Go binary embedding
  - Removed dynamic routes that were incompatible

### Added

- New UI pages:
  - `/` - Dashboard with stats, active agents, recent events
  - `/agents` - Agent list with detail sheet
  - `/projects` - Project list with detail sheet
  - `/tasks` - 8-column Kanban board (backlog→failed)
  - `/events` - Event log with filtering
  - `/settings` - Configuration page

---

## [Unreleased]

### Added (implementation summaries consolidated)

- **Projects API** — Full CRUD: `GET/POST /api/v1/projects`, `GET/PUT/DELETE /api/v1/projects/:id`, `GET /api/v1/projects/:id/tasks`. DB table `projects` (id, name, description, status, color). Tasks can be linked via `project_id`.
- **Comments API** — Task comments: `GET/POST /api/v1/tasks/:id/comments`, `DELETE /api/v1/comments/:id`. DB table `comments` (id, task_id, author, content, created_at). Cascade delete when task is deleted.
- **Task extensions** — `project_id` and `parent_task_id` on tasks; `ListTasksByProject`, `ListSubtasks` queries. Migrations 000008 (projects), 000009 (comments), 000010 (extend tasks).
- **Chat API** — Agent chat sessions and messages: `POST/GET /api/v1/agents/:id/sessions`, `DELETE .../sessions/:sessionId`, `GET/POST .../sessions/:sessionId/messages`. OpenClaw integration for spawn/send.
- **Tasks page rebuild** — Kanban aligned to Figma: 5 columns (Inbox, Assigned, In Progress, Review, Done), status-colored dots, approach badges (GSD/RALPH), progress and assignee on cards.

### Planned for v1.1

- [ ] Unit test coverage (target: 80%)
- [ ] Integration tests for API endpoints
- [ ] Rate limiting on API endpoints
- [ ] User authentication (optional, for multi-user setups)
- [ ] Complete UI implementation (all wireframes → real components)
- [ ] API client SDK (Go, TypeScript)
- [ ] Metrics and observability (Prometheus endpoint)
- [ ] Webhook support for external integrations
- [ ] Task templates
- [ ] Agent cloning
- [ ] Bulk operations (archive, delete, reassign)

### Planned for v2.0

- [ ] Multi-user support with RBAC
- [ ] Team collaboration features
- [ ] Agent marketplace (share agent configurations)
- [ ] Custom execution engines (plugin system)
- [ ] Task scheduling and recurring tasks
- [ ] Advanced analytics and reporting
- [ ] Mobile app (React Native)
- [ ] Cloud deployment options (AWS, GCP, Azure)

---

## Version History

- **[1.0.0]** - 2026-02-08 - Initial release
- **[0.1.0]** - 2026-02-01 - Internal prototype

---

## Notes

### Versioning Policy

- **Major version (X.0.0)**: Breaking API changes, major feature additions
- **Minor version (1.X.0)**: New features, backward compatible
- **Patch version (1.0.X)**: Bug fixes, performance improvements

### Deprecation Policy

- Features marked as deprecated will be maintained for at least one minor version
- Breaking changes will be announced in advance
- Migration guides will be provided

### Support

- Latest major version: Full support
- Previous major version: Security fixes only
- Older versions: Community support only

---

**Legend:**
- `Added` - New features
- `Changed` - Changes to existing functionality
- `Deprecated` - Soon-to-be removed features
- `Removed` - Removed features
- `Fixed` - Bug fixes
- `Security` - Security improvements
