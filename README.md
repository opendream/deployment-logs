# Deployment Logs

Tracks features and fixes deployed to each environment (develop, UAT, production) by analyzing PR merge history from GitHub and GitLab repositories. Triggered via API, it pulls git history, extracts PRs with conventional commit prefixes, and stores deployment logs grouped by date and environment.

## Tech Stack

- **Backend**: Go 1.24 + Fiber v2 + GORM
- **Frontend**: React 18 + Vite + TypeScript
- **Database**: SQLite (default), MSSQL, PostgreSQL, MySQL
- **Git**: go-git/v5 for repository cloning/pulling
- **Deployment**: Docker (multi-stage build)

## Prerequisites

- Go 1.24+
- Node.js 22+
- Git

## Setup

### 1. Clone the repository

```bash
git clone <repo-url>
cd deployment-logs
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` as needed:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_DRIVER` | Database driver (`sqlite`, `postgres`, `mysql`, `mssql`) | `sqlite` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `1433` |
| `DB_USER` | Database user | `sa` |
| `DB_PASSWORD` | Database password | |
| `DB_NAME` | Database name | `deployment_logs` |
| `DB_SCHEMA` | Schema name (MSSQL/Postgres) | `deploymentlogs` |
| `PORT` | Server port | `3000` |
| `APP_PORT` | Exposed port (Docker) | `4900` |
| `REPOS_DIR` | Directory for cloned repos | `./repos` |
| `ADMIN_PASSWORD` | Admin login password | `admin` |
| `JWT_SECRET` | JWT signing secret | |
| `BASE_PATH` | URL base path for reverse proxy setups | |

> **Note:** The API key for the public trigger endpoint is managed via the admin Settings UI (stored in the database), not through environment variables.

### 3. Run locally

**Backend:**

```bash
go run ./cmd/api
```

**Frontend (dev mode with hot reload):**

```bash
cd web
npm install
npm run dev
```

The frontend dev server proxies `/api` requests to the backend on port 3000.

### 4. Run with Docker

```bash
docker compose up --build
```

This starts the app on port 4900 (configurable via `APP_PORT`) with an MSSQL database. For SQLite, set `DB_DRIVER=sqlite` in your `.env`.

## Usage

### Web UI

Open `http://localhost:4900` (Docker) or `http://localhost:5173` (dev mode) to access the config UI where you can:

- Add/edit repository configurations (GitHub or GitLab)
- Configure branch-to-environment mappings
- View deployment logs

### API

All endpoints are under `/api`. Admin endpoints require JWT authentication (login with `ADMIN_PASSWORD`).

**Public endpoints:**

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/auth/login` | None | Login with admin password |
| GET | `/api/public/logs/repo/:repo_name` | None | View logs for a repo |
| POST | `/api/public/trigger` | `X-API-Key` (if set) | Trigger log generation (public if no key configured) |
| GET | `/api/repos` | None | List repositories |
| GET | `/api/logs` | None | List logs (filter: `?repo=X&env=Y`) |
| GET | `/api/logs/:id` | None | Get log with items |

**Admin endpoints (JWT required):**

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/repos` | Create repo config |
| PUT | `/api/repos/:id` | Update repo config |
| DELETE | `/api/repos/:id` | Delete repo config |
| POST | `/api/trigger` | Trigger log generation |
| DELETE | `/api/logs/repo/:repo_name` | Clear logs for a repo |
| GET | `/api/settings` | Get app settings |
| PUT | `/api/settings` | Update app settings (API key) |

**Public trigger example:**

```bash
# First, set the API key via admin settings UI or API
curl -X PUT http://localhost:4900/api/settings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt-token>" \
  -d '{"app_key": "my-secret-key"}'

# Then trigger using the API key
curl -X POST http://localhost:4900/api/public/trigger \
  -H "Content-Type: application/json" \
  -H "X-API-Key: my-secret-key" \
  -d '{"repo_name": "my-repo", "branch": "main"}'
```

### Build & Publish Docker Image

```bash
# One-time setup: create a multi-platform buildx builder
make setup

# Build and push multi-platform image (linux/amd64 + linux/arm64)
make build
```

### Build frontend for production

```bash
cd web && npm run build
```

The built files are served as static assets by the Go binary.

## How It Works

1. A repository config is created with the repo URL, provider (GitHub/GitLab), auth token, and branch mappings
2. When triggered, the service clones or pulls the repository
3. In **PR mode**: fetches merged PRs via GitHub/GitLab API, filtered by conventional commit prefixes (`feat`, `fix`, `hotfix`, `chore`, `refactor`)
4. In **commit mode**: walks the git log and filters commits by the same prefixes
5. Results are saved as a deployment log with individual log items

## License

This project is licensed under the [MIT License](LICENSE).
