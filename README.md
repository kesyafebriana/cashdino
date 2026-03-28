# CashDino Weekly Challenge Leaderboard

A weekly leaderboard feature for a rewards app (practice project inspired by CashGiraffe). Users earn gems from gameplay, check-ins, and surveys, compete on a weekly leaderboard, and top-ranked users win rewards (gems or gift cards).

## Project Structure

```
cashdino/
├── backend/          # Go API (Echo framework, PostgreSQL)
├── mobile/           # React Native app (Expo)
├── admin/            # Admin dashboard (Next.js, Tailwind CSS)
├── docker-compose.yml
├── .env.example
└── Makefile
```

## Tech Stack

| Component | Technology |
| --------- | ---------- |
| Backend   | Go 1.23, Echo v4, pgx, robfig/cron, gomail |
| Mobile    | React Native, Expo, expo-router |
| Admin     | Next.js 15, App Router, Tailwind CSS v4 |
| Database  | PostgreSQL 16 |
| Hosting   | Docker Compose on VPS |

## Prerequisites

- Docker & Docker Compose
- Go 1.23+ (for local backend development)
- Node.js 20+ (for local mobile/admin development)
- Expo CLI (for mobile development)

## Quick Start

```bash
# 1. Clone the repository
git clone git@github.com:kesyafebriana/cashdino.git
cd cashdino

# 2. Copy environment variables
cp .env.example .env

# 3. Start all services
make up

# 4. Run database migrations
make migrate-up

# 5. Seed the database
make seed
```

The API will be available at `http://localhost:8080` and the admin dashboard at `http://localhost:3000`.

## Makefile Targets (root)

| Target             | Description                            |
| ------------------ | -------------------------------------- |
| `make up`          | Build and start all services           |
| `make down`        | Stop all services                      |
| `make logs`        | Tail logs from all services            |
| `make seed`        | Run the database seed script           |
| `make migrate-up`  | Run all pending migrations             |
| `make migrate-down`| Rollback the last migration            |
| `make reset`       | Full reset: down + up + migrate + seed |

## Makefile Targets (backend)

| Target                          | Description                     |
| ------------------------------- | ------------------------------- |
| `make run`                      | Run the API server locally      |
| `make build`                    | Build the server binary         |
| `make seed`                     | Run the seed script             |
| `make migrate-up`               | Run all pending migrations      |
| `make migrate-down`             | Rollback the last migration     |
| `make migrate-create name=xxx`  | Create a new migration          |
| `make test`                     | Run all tests                   |
| `make lint`                     | Run golangci-lint               |

## Mobile Development

The mobile app runs locally via Expo and connects to the API on your VPS (or localhost).

```bash
cd mobile
npm install
npx expo start
```

Set the API URL in your environment:

```
EXPO_PUBLIC_API_URL=http://<your-vps-ip>:8080
```

No authentication — use the user switcher to pick a test user by ID.

## Health Check

```bash
curl http://localhost:8080/api/health
# {"status":"ok"}
```

## Documentation

Detailed project documentation lives in `.claude/docs/`:

| Document | Description |
| -------- | ----------- |
| [PRD](.claude/docs/PRD.md) | Product requirements — features, business rules, UI specs |
| [System Design](.claude/docs/SYSTEM_DESIGN.md) | Architecture, API contracts, DB schema, cron jobs |

## Claude Code Setup

This project uses [Claude Code](https://docs.anthropic.com/en/docs/claude-code) with project-level configuration in `.claude/`:

- **CLAUDE.md** — Project context, business rules, DB schema, API contracts, code conventions, and guardrails. Claude reads this automatically for every conversation.
- **docs/PRD.md** — Full product requirements document (source of truth for features and business rules).
- **docs/SYSTEM_DESIGN.md** — System design document (source of truth for architecture, APIs, DB schema).
- **skills/** — Reusable prompt templates for common workflows:

| Skill | File | Description |
| ----- | ---- | ----------- |
| TDD Go | `skills/tdd-go.md` | Test-driven development workflow — write tests first, then implementation, per layer |
| Code Reviewer | `skills/code-reviewer.md` | Review checklist for Go, React Native, and Next.js code |
| Figma to Code | `skills/figma-to-code.md` | Convert Figma designs to pixel-accurate code (supports Figma MCP or screenshots) |
| CI/CD Setup | `skills/cicd-setup.md` | GitHub Actions CI + VPS deploy setup instructions |
