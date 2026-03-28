You are a DevOps engineer setting up simple CI/CD for a practice project. Keep it minimal — no over-engineering.

Project context:
- Monorepo: backend (Go), mobile (React Native/Expo), admin (Next.js)
- Hosted on a VPS with Docker Compose
- No Kubernetes, no AWS services, no Terraform
- Budget: $0 (use free tiers only)

CI tool: GitHub Actions (free for public repos, 2000 min/month for private)

Create these workflow files:

.github/workflows/backend.yml — triggers on push to backend/**
- Job 1: lint
  - golangci-lint
- Job 2: test (depends on lint passing)
  - Start postgres service container (postgres:16)
  - Run migrations
  - Run go test ./... -v
- Job 3: build
  - go build — verify it compiles
- Only runs if files in backend/ changed

.github/workflows/admin.yml — triggers on push to admin/**
- Job 1: lint + typecheck
  - npm ci
  - next lint
  - tsc --noEmit
- Job 2: build
  - next build — verify it builds
- Only runs if files in admin/ changed

.github/workflows/deploy.yml — triggers on push to main branch ONLY
- SSH into VPS
- git pull
- docker-compose up -d --build
- Run migrations
- Uses GitHub Secrets: VPS_HOST, VPS_USER, VPS_SSH_KEY

DO NOT create:
- Separate staging/production environments
- Docker registry (build on VPS directly)
- Helm charts or K8s manifests
- Complex branch strategies (just main + feature branches)
- Mobile CI (Expo builds are handled by EAS, skip for now)

Add to README.md:
- CI/CD section explaining what runs on push
- How to set up GitHub Secrets for deploy
- How to manually deploy (ssh + docker-compose)

Add to root Makefile:
  make deploy        # ssh into VPS and run deploy steps (for manual deploy)
  make ci-local      # run all CI checks locally before pushing