# 🎮 Claw Agent Mission Control

AI Agent Orchestration Dashboard for OpenClaw Gateway

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Next.js](https://img.shields.io/badge/Next.js-14+-000000?style=flat-square&logo=next.js)](https://nextjs.org/)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)

## Overview

Claw Agent Mission Control is a production-grade dashboard for managing AI agents and tasks. It implements the **GSD (Get Shit Done)** and **Ralph Loop** methodologies for autonomous, context-aware task execution.

### Key Features

- 🤖 **Agent Management** - Create, configure, and monitor AI agents
- 📋 **Task Board** - Asana-style Kanban board for task management
- ⚡ **GSD Execution** - Spec-driven development with research, planning, and verification
- 🔄 **Ralph Loop** - Autonomous iteration until all PRD items pass
- 👥 **Sub-Agent Spawning** - Hierarchical agent orchestration
- 📡 **Real-time Updates** - WebSocket-powered live status
- 🌙 **Dark Mode** - Clean, minimal interface

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+
- OpenClaw Gateway running

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/claw-agent-mission-control.git
cd claw-agent-mission-control

# Copy environment file
cp .env.example .env

# Edit .env with your OpenClaw Gateway details
# OPENCLAW_GATEWAY_URL=ws://127.0.0.1:18789
# OPENCLAW_GATEWAY_TOKEN=your-token

# Build and run
make build
./bin/mission-control
```

Open http://localhost:8080 in your browser.

### Development

```bash
# Start development mode (hot reload)
make dev

# Frontend only (separate terminal)
cd ui && npm run dev

# Backend only (separate terminal)
air
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `OPENCLAW_GATEWAY_URL` | OpenClaw WebSocket URL | `ws://127.0.0.1:18789` |
| `OPENCLAW_GATEWAY_TOKEN` | Authentication token | (required) |
| `DATABASE_PATH` | SQLite database path | `./data/mission-control.db` |
| `DEFAULT_APPROACH` | Default task approach | `gsd` |

See [.env.example](.env.example) for all options.

## Architecture

```
┌─────────────────────────────────────────┐
│         Single Go Binary                │
│  ┌─────────────┐  ┌─────────────────┐  │
│  │  Embedded   │  │   Go HTTP       │  │
│  │  Next.js UI │  │   Server        │  │
│  │             │  │   + WebSocket   │  │
│  └─────────────┘  └─────────────────┘  │
│           │              │              │
│           ▼              ▼              │
│     ┌─────────────────────────┐        │
│     │   SQLite Database       │        │
│     └─────────────────────────┘        │
└─────────────────────────────────────────┘
                    │
                    ▼
          ┌─────────────────┐
          │ OpenClaw Gateway │
          └─────────────────┘
```

## Documentation

- [Product Requirements Document](docs/PRD.md)
- [API Reference](docs/API.md) (coming soon)
- [Development Guide](docs/DEVELOPMENT.md) (coming soon)

## Contributing

Contributions are welcome! Please read our contributing guidelines first.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [OpenClaw](https://github.com/openclaw/openclaw) - AI Agent Runtime
- [GSD](https://github.com/glittercowboy/get-shit-done) - Context Engineering
- [Ralph](https://github.com/snarktank/ralph) - Autonomous Agent Loop
- [Temporal UI Server](https://github.com/temporalio/ui-server) - Embedding Pattern
