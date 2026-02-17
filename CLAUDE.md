# Deployment Logs Service

## Overview
Tracks features/fixes deployed to each environment (develop, uat, production) by analyzing PR merge history from GitHub/GitLab repositories. Triggered via API, it pulls git history, extracts PRs with conventional commit prefixes, and stores deployment logs grouped by date and environment.

## Tech Stack
- **Backend**: Go 1.23 + Fiber v2 + GORM (auto-migration)
- **Frontend**: React 18 + Vite + TypeScript (config UI only)
- **Database**: SQLite (default), MSSQL, PostgreSQL, MySQL via GORM dialects
- **Git**: go-git/v5 for repository cloning/pulling
- **Deployment**: Docker (multi-stage build), no persistent volume required for production with external DB

## Project Structure
```
deployment-logs/
├── main.go                        # Entry point, route wiring, CORS, static serving
├── config/config.go               # Env-based config loading (.env via godotenv)
├── database/database.go           # GORM multi-driver setup + auto-migrate
├── models/
│   ├── repository_config.go       # RepositoryConfig model
│   ├── deployment_log.go          # DeploymentLog model (has many LogItems)
│   └── log_item.go                # LogItem model (individual PR/commit entry)
├── handlers/
│   ├── config.go                  # CRUD /api/configs
│   ├── trigger.go                 # POST /api/trigger
│   ├── logs.go                    # GET /api/logs, /api/logs/:id
│   └── settings.go                # GET/PUT /api/settings
├── services/
│   ├── git_service.go             # Clone/pull repos using go-git
│   ├── github_service.go          # GitHub PR fetching via REST API
│   ├── gitlab_service.go          # GitLab MR fetching via REST API
│   ├── commit_service.go          # Commit-based log extraction (fallback)
│   └── log_generator.go           # Core orchestration: analyze merges → generate logs
├── middleware/auth.go             # X-API-Key auth (skip if APP_KEY empty)
├── web/                           # React frontend
│   ├── src/
│   │   ├── App.tsx                # Routes + nav bar
│   │   ├── pages/ConfigList.tsx   # List/delete repo configs
│   │   ├── pages/ConfigForm.tsx   # Create/edit repo config
│   │   └── pages/Settings.tsx     # App key management
│   ├── vite.config.ts             # Proxy /api to backend in dev
│   └── package.json
├── Dockerfile                     # Multi-stage: node → go → alpine
├── docker-compose.yml
└── .env.example
```

## Database Schema
Three tables, auto-migrated by GORM:
- **repository_configs** — repo name, URL, provider (github/gitlab), mode (pr/commit), auth token, branches (JSON string)
- **deployment_logs** — links to repo config, environment name, generation timestamp
- **log_items** — PR/commit title, number, SHA, merged_at, category (feat/fix/hotfix/chore/refactor)

## API Endpoints
| Method | Path | Description |
|--------|------|-------------|
| GET | /api/configs | List all repo configs |
| POST | /api/configs | Create repo config |
| PUT | /api/configs/:id | Update repo config |
| DELETE | /api/configs/:id | Delete repo config |
| POST | /api/trigger | Trigger log generation `{repo_name, environment}` |
| GET | /api/logs | List logs `?repo=X&env=Y` |
| GET | /api/logs/:id | Get log with items |
| GET | /api/settings | Get app settings |
| PUT | /api/settings | Update app settings |

Auth: `X-API-Key` header required when `APP_KEY` env var is set.

## Core Logic (log_generator.go)
1. Find repo config by name
2. Clone or pull repo (re-clones if missing — container-safe)
3. Determine "since" from last deployment log (default: 30 days ago)
4. PR mode: fetch merged PRs via GitHub/GitLab API, filter by conventional commit pattern
5. Commit mode: walk git log, filter by `^(feat|fix|hotfix|chore|refactor)(\(.*\))?:`
6. Save DeploymentLog + LogItems in a transaction
7. Return log with items

## Environment Variables
See `.env.example`. Key vars:
- `DB_DRIVER` — sqlite|postgres|mysql|mssql
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SCHEMA`
- `APP_KEY` — API authentication key (empty = no auth)
- `PORT` — server port (default 3000)
- `REPOS_DIR` — local clone directory (ephemeral in containers)

## Development Commands
```bash
# Backend
go run main.go

# Frontend dev (with hot reload, proxies /api to :3000)
cd web && npm run dev

# Build frontend for production
cd web && npm run build

# Docker
docker compose up --build
```

## Architecture Notes
- Handlers use closure pattern: `func ListConfigs(db *gorm.DB) fiber.Handler` returning a `func(c *fiber.Ctx) error`
- Frontend stores API key in localStorage, sends via `X-API-Key` header on all requests
- Git repos are cloned to `REPOS_DIR/{repo_name}`, safe to lose on container restart
- MSSQL/Postgres support schema isolation via `DB_SCHEMA`
- Frontend is built and served as static files by the Go binary in production
