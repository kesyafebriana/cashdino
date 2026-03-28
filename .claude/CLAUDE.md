# CashDino — Weekly Challenge Leaderboard

## What is this project?
A weekly leaderboard feature for a rewards app (practice project inspired by CashGiraffe). Users earn gems from gameplay/check-ins/surveys, compete on a weekly leaderboard, and top-ranked users win rewards (gems or gift cards).

## Key documents
- PRD: .claude/docs/PRD.md (source of truth for all features and business rules)
- System Design: .claude/docs/SYSTEM_DESIGN.md (source of truth for architecture, APIs, DB schema)
- If a prompt contradicts the PRD or system design, follow the PRD/system design — not the prompt.

## Architecture
- Backend: Go (Echo), PostgreSQL, pgx, robfig/cron, gomail
- Mobile: React Native (Expo)
- Admin: Next.js (App Router, Tailwind CSS)
- Hosting: Docker Compose on VPS
- Email: SMTP (Gmail or console.log for dev)

## Business rules (from PRD — do not deviate)
- All users participate automatically — no opt-in/opt-out
- Leaderboard shows Top 99 only, no search, no pagination beyond 99
- Users outside Top 99 see "99+" — not their exact rank
- Rank is NEVER stored for active week — derived from query row position (ORDER BY weekly_gems DESC, first_gem_earned_at ASC)
- Rank IS stored in weekly_challenge_results at week end (frozen snapshot)
- User balance = SUM(gem_history.amount) — no stored balance column
- Masked identity: first 2 chars + "****" + last char, collision → append random digit
- Weekly cycle: Monday 00:00 UTC → Sunday 23:59 UTC
- Reward rules: rank ranges with multiple rewards per range, no overlapping ranges
- Gem rewards: auto-credit via gem_history (source='reward')
- Non-gem rewards: send claim email via SMTP, template with placeholders
- Stock validation: total users across ranges per reward type must not exceed stock
- Claim emails are transactional (not marketing) — no unsubscribe needed

## DB schema rules (from system design — do not add/remove columns)
- users: id, username, email, created_at (NO total_gems column)
- gem_history: id, user_id, source (enum), amount (+/-), game_name, created_at
- daily_checkins: id, date, base_gems, streak_multiplier, is_active
- user_daily_checkins: id, user_id, checkin_id, gems_earned, current_streak, checked_in_at
- weekly_challenges: id, start_time, end_time, status
- leaderboard_entries: id, challenge_id, user_id, weekly_gems, first_gem_earned_at, display_name
- weekly_challenge_results: id, challenge_id, user_id, final_rank, final_gems, display_name
- reward_campaigns: id, challenge_id, name, banner_image, rules (JSONB), non_gem_claim_email_subject, non_gem_claim_email_body, status
- reward_types: id, campaign_id, name, type, value, image (URL), stock
- reward_distributions: id, campaign_id, user_id, reward_type_id, status, delivered_at, email_sent_at

## API contract (from system design — do not change endpoints or response shapes)
- GET /api/challenge/banner?user_id=
- GET /api/leaderboard/current?user_id=
- GET /api/leaderboard/last-week?user_id=
- POST /api/gems/earn { user_id, source, amount, game_name }
- POST /api/checkin { user_id }
- GET /api/admin/campaigns
- GET /api/admin/campaigns/:id
- POST /api/admin/campaigns
- PUT /api/admin/campaigns/:id
- GET /api/admin/campaigns/:id/distributions
- POST /api/admin/reset-week

## Code conventions
- Go: handler/ → service/ → repository/ layers, no business logic in handlers
- Go: all repository functions take context.Context as first arg
- Go: errors wrapped with fmt.Errorf("doing X: %w", err)
- Go: struct fields have json tags
- React Native: StyleSheet.create, no Tailwind
- Next.js: Tailwind CSS, App Router, 'use client' only where needed
- No authentication in MVP — user_id passed as query param or request body

## What NOT to do
- Do not add features not in the PRD
- Do not create new DB tables or columns not in the schema
- Do not change API endpoint paths or response shapes
- Do not add opt-in/opt-out — all users participate automatically
- Do not store rank for active week — always derive from query
- Do not add a total_gems column to users — always SUM from gem_history
- Do not build search or pagination beyond Top 99
- Do not add authentication/login (MVP shortcut)