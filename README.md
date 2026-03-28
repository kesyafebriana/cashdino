# CashDino Weekly Challenge Leaderboard

Monorepo containing the backend API (Go/Echo), mobile app (React Native/Expo), and admin dashboard (Next.js).

## Project Structure

```
cashdino/
├── backend/    # Go API (Echo framework)
├── mobile/     # React Native app (Expo)
├── admin/      # Admin dashboard (Next.js)
```

## Prerequisites

- Docker & Docker Compose
- Go 1.23+ (for local backend development)
- Node.js 20+ (for local mobile/admin development)

## Quick Start

```bash
# 1. Clone the repository
git clone <repo-url> cashdino
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

## Makefile Targets

| Target           | Description                          |
| ---------------- | ------------------------------------ |
| `make up`        | Build and start all services         |
| `make down`      | Stop all services                    |
| `make logs`      | Tail logs from all services          |
| `make seed`      | Run the database seed script         |
| `make migrate-up`| Run all pending migrations           |
| `make migrate-down` | Rollback the last migration       |
| `make reset`     | Full reset: down + up + migrate + seed |

## Health Check

```bash
curl http://localhost:8080/api/health
# {"status":"ok"}
```
