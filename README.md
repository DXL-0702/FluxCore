<div align="center">

# FluxCore

**Code is Log. Push is Update.**

A developer-first R&D operations console that turns every `git push` into a living project dashboard — zero manual logging, zero context switching.

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react&logoColor=black)](https://react.dev)
[![Vite](https://img.shields.io/badge/Vite-6-646CFF?style=flat-square&logo=vite&logoColor=white)](https://vitejs.dev)
[![SQLite](https://img.shields.io/badge/SQLite-003B57?style=flat-square&logo=sqlite&logoColor=white)](https://sqlite.org)
[![Redis](https://img.shields.io/badge/Redis-DC382D?style=flat-square&logo=redis&logoColor=white)](https://redis.io)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

[English](README.md) · [简体中文](README.zh-CN.md)

</div>

---

## Why FluxCore?

Most project management tools force developers into a **dual-loop workflow** — write code *and then* manually log what you did. FluxCore eliminates the second loop entirely.

| Pain Point | Traditional Tools | FluxCore |
| :--- | :--- | :--- |
| Progress tracking | Manual timesheets / ticket updates | Auto-generated from commits |
| Project status | Stale dashboards, outdated docs | Real-time via WebSocket |
| Multi-project switching | Scattered terminals, lost context | `fluxcore switch` restores everything |
| README as source of truth | Disconnected from dev flow | Parsed & synced on every push |

## Core Philosophy

- **Invisible Logging** — No manual timesheets. Git Hooks + Commit Message parsing auto-generate structured activity logs.
- **Real-Time Feedback** — Jenkins-like immediacy. Push your code, and the Web dashboard refreshes via WebSocket within milliseconds.
- **Context Awareness** — Whether you're in the CLI or the Web UI, FluxCore intelligently detects the current project state and README metadata.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Developer Workflow                       │
│   git commit → post-commit hook → fluxcore CLI → Backend API    │
└────────────────────────────┬────────────────────────────────────┘
                             │
              ┌──────────────▼──────────────┐
              │      Go Backend (Gin)       │
              │  ┌────────┐  ┌───────────┐  │
              │  │Webhook │  │  Commit    │  │
              │  │Receiver│  │  Parser    │  │
              │  └───┬────┘  └─────┬─────┘  │
              │      │             │         │
              │  ┌───▼─────────────▼─────┐  │
              │  │     GORM (SQLite/PG)  │  │
              │  └───────────┬───────────┘  │
              │              │              │
              │  ┌───────────▼───────────┐  │
              │  │  Redis Pub/Sub + WS   │  │
              │  └───────────┬───────────┘  │
              └──────────────┼──────────────┘
                             │
              ┌──────────────▼──────────────┐
              │   React + Vite Dashboard    │
              │  Real-time project cards,   │
              │  activity timeline, CI view  │
              └─────────────────────────────┘
```

## Tech Stack

| Layer | Technology | Role |
| :--- | :--- | :--- |
| **CLI** | Go (Cobra) | Git Hook injection, local context management, project switching |
| **Backend** | Go (Gin) | REST API, Webhook processing, WebSocket broadcast |
| **ORM** | GORM | Data persistence with SQLite/PostgreSQL dynamic switching |
| **Database** | SQLite / PostgreSQL | SQLite for zero-config dev, PostgreSQL for production |
| **Messaging** | Redis | WebSocket Pub/Sub for real-time event distribution |
| **Frontend** | React 19 + Vite 6 | Modern SPA dashboard |
| **Styling** | Tailwind CSS | Utility-first CSS for rapid UI development |

## Workflow Preview

```bash
# 1. Link your project
fluxcore link --project "my-awesome-app"

# 2. Develop and commit as usual
git commit -m "feat: #101 implement payment API"
git push

# 3. Everything else is automatic
#    → CLI: push confirmation in terminal
#    → Web: project card version bumps instantly,
#           timeline shows new log entry,
#           task #101 auto-transitions to "Testing"
```

## Project Structure

```
FluxCore/
├── cli/                  # CLI tool (Go + Cobra)
│   ├── cmd/              #   Command definitions
│   ├── internal/         #   Hook injection, config management
│   └── main.go
├── server/               # Backend (Go + Gin)
│   ├── api/              #   HTTP route handlers
│   ├── model/            #   GORM data models
│   ├── service/          #   Business logic layer
│   ├── ws/               #   WebSocket hub
│   ├── db/               #   Database connection (SQLite/PG)
│   └── main.go
├── web/                  # Frontend (React + Vite)
│   ├── src/
│   │   ├── components/   #   Reusable UI components
│   │   ├── pages/        #   Route pages
│   │   ├── hooks/        #   Custom React hooks
│   │   └── lib/          #   API client, WebSocket client
│   └── index.html
├── migrations/           # Database migration files
├── docker-compose.yml    # Production deployment
└── .fluxcore/            # Local config (git-ignored)
```

## Getting Started

> **Note:** FluxCore is under active development. The following instructions reflect the target setup.

### Prerequisites

- Go 1.23+
- Node.js 20+ & npm
- Redis (for Phase 3+)

### Development

```bash
# Clone the repository
git clone https://github.com/your-username/FluxCore.git
cd FluxCore

# Start the backend
cd server
API_TOKEN=local-dev-token go run .

# Start the frontend (in a new terminal)
cd web
npm install && npm run dev

# Install the CLI
cd cli
go install .

# Initialize FluxCore in any project
cd /path/to/your/project
fluxcore init --server http://127.0.0.1:8080 --token local-dev-token
fluxcore link --project "my-project"
fluxcore status
```

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

1. Fork the repository
2. Create your feature branch (`git checkout -b feat/amazing-feature`)
3. Commit your changes following [Conventional Commits](https://www.conventionalcommits.org/)
4. Push to the branch (`git push origin feat/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the [MIT License](LICENSE).

---

<div align="center">

**FluxCore** — Let your code speak for itself.

</div>
