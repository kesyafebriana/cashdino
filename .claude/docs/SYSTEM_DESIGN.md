## 1. Stack

- **Backend:** Go (Echo)
- **Mobile App:** React Native (user-facing)
- **Admin Web:** Next.js (admin dashboard)
- **Database:** PostgreSQL
- **Email:** SMTP
- **Cron:** robfig/cron library
- **Hosting:** Railway/Render

---

## 2. Seed Data

Since this is practice with no real app, seed the DB on startup:

- ~50 fake users (use faker or hardcode)
- A few daily_checkins configs for the current week
- 1 active weekly_challenge (this week)
- 1 completed weekly_challenge (last week) with results + reward distributions
- Scatter gem_history records across users so leaderboard has data
- 1 reward_campaign with reward_types + rules for last week (so last week tab has 🎁 icons)

---

## 3. API Endpoints

### User APIs

**GET `/api/challenge/banner`**

- Return: active challenge end_time, user's weekly_gems, rank (or "99+"), gap to next rank

**GET `/api/leaderboard/current`**

- Query leaderboard_entries for active challenge, order by weekly_gems DESC + first_gem_earned_at ASC, limit 99
- Rank = row index in result (not stored)
- Include current user's position (or "99+" if outside top 99)
- Include campaign info if exists: banner_image + reward rules summary with images

**GET `/api/leaderboard/last-week`**

- Query weekly_challenge_results for previous challenge, order by final_rank, limit 99
- Left join reward_distributions + reward_types to attach rewards per user
- Return rewards as array per row (can be empty or multiple)

**POST `/api/gems/earn`**

- Input: source, amount, game_name (optional)
- Insert gem_history
- Upsert leaderboard_entries: increment weekly_gems, set first_gem_earned_at if null
- Return: updated weekly_gems

**POST `/api/checkin`**

- Reject if already checked in today
- Calculate streak: last check-in yesterday → streak + 1, otherwise reset to 1
- gems_earned = base_gems × streak_multiplier
- Insert user_daily_checkins + gem_history (source = 'daily_checkin')
- Upsert leaderboard_entries
- Return: gems_earned, current_streak

### Admin APIs

**POST `/api/admin/campaigns`**

- Input: challenge_id, name, banner_image, reward_types array, rules array, email subject, email body
- Validate: no overlapping rank ranges, stock ≥ users per reward type
- Save reward_types → get IDs → store rules JSONB with reward_type_ids
- Return: created campaign

**PUT `/api/admin/campaigns/:id`**

- Same as create with same validations

**GET `/api/admin/campaigns/:id/distributions`**

- Return: reward_distributions list with user info + status

**POST `/api/admin/reset-week`**

- Manual trigger for weekly reset (since cron timing is inconvenient for testing)
- Same logic as the cron job — lets you test the full flow on demand

---

## 4. Cron Job (robfig/cron)

### Weekly Reset

Runs Sunday 23:59 UTC — but also expose as `POST /api/admin/reset-week` for manual testing.

Steps in order:

1. Mark active challenge → `completed`
2. Snapshot leaderboard into weekly_challenge_results — final_rank = row position from ordered query
3. If reward campaign exists → distribute:
    - For each rule → for each user in rank range → for each reward_type_id:
        - Gems → insert gem_history (source = 'reward') + reward_distributions (delivered)
        - Non-gems → log email to console (or send via Mailtrap) + reward_distributions (delivered or failed)
    - Decrement stock
4. Mark campaign → `completed`
5. Create next challenge (next Monday → Sunday, status = 'active')

### Email Retry

- Every 8 hours (node-cron)
- Query reward_distributions where status = 'failed'
- Retry (max 3 total) → if still fails → console.log alert
- For practice: can just console.log instead of actual email

---

## 5. Key Logic

### Masked Identity

- First 2 chars + `***` + last char (e.g., `james` → `ja****s`)
- ≤ 4 chars → first char + `***`
- Collision → append random digit
- Generate once on leaderboard_entry creation

### Rank

- This week: row index from query — never stored
- Last week: final_rank from weekly_challenge_results
- Outside Top 99 → "99+"

### Gap to Next Rank

- Rank N (N > 1): gap = gems of rank N-1 minus current user's gems + 1
- Rank 1: "You're in the lead"
- 99+: no gap shown

### Stock Validation (on campaign save)

- For each reward_type, sum total users across all ranges using it
- Must not exceed stock → reject with error if insufficient

### Rank Range Overlap Validation

- Sort rules by rank_from → each rank_from must be > previous rank_to
- Overlap → reject with error

---

## 6. Email (SMTP — free)

Use Gmail SMTP (free, 500 emails/day) or any personal email SMTP.

**Gmail SMTP setup:**

- Host: `smtp.gmail.com`
- Port: `587`
- Username: your Gmail address
- Password: App Password (generate in Google Account → Security → App Passwords)
- TLS: enabled

**How it works:**

- Go sends email via `net/smtp` or `gomail` library
- Emails actually deliver to real inboxes — good for testing end-to-end
- Replace `{{username}}`, `{{rank}}`, `{{reward_type}}`, `{{reward_value}}`, `{{reward_image}}` in template
- `{{reward_image}}` → `<img>` tag with URL

**Fallback:** If you don't want real emails, just console.log the rendered email and mark as delivered

---

## 7. Frontend Pages

### Mobile App (React Native)

- **Home screen** — challenge banner (rank, gems, countdown, "View Leaderboard" button)
- **Leaderboard screen** — two tabs:
    - **This Week** — top 99 list, sticky footer if 99+, prize banner at top (tappable → modal), "See all rewards" link
    - **Last Week** — archived top 99, 🎁 icon on rewarded rows (tap → bubble with reward details)
- **Reward Details Modal** — lists all rewards per rank range with images
- **Reward Bubble** — tap 🎁 on last week → shows that user's rewards (can be multiple)
- **Onboarding Modal** — 3 screens (what's new, how it works, CTA to leaderboard), shown once, stored in AsyncStorage
- **Mock Gem Earn button** — somewhere on the page to simulate earning gems (calls POST /api/gems/earn)
- **Daily Check-in button** — calls POST /api/checkin

### Admin Web (React / Next.js)

- **Campaign creator page** — form to create campaign: name, banner image URL, reward types (name, type, value, image URL, stock), rank-range rules (assign multiple reward types per range), email subject + body template
- **Campaign list page** — see all campaigns with status
- **Distribution report page** — see reward distributions per campaign
- **Manual reset button** — calls POST /api/admin/reset-week for testing

---

## 9. Dev Shortcuts (for practice)

- Mobile app: use `?user_id=xxx` param or a dropdown to switch between users
- Seed script that populates realistic data on startup (`./seed` or `go run cmd/seed/main.go`)
- Manual reset endpoint so you don't have to wait until Sunday
- Console.log or Mailhog for emails
- Skip authentication — no login needed, just pick a user
- Images: use placeholder URLs (e.g., picsum.photos or static files)
- React Native: use Expo for fast setup, point API_BASE_URL to VPS IP

---

## 8. Docker Compose Structure

**Services:**

- `api` — Go binary (builds from Dockerfile with multi-stage build, small image)
- `db` — postgres:16 (with volume for persistence)
- `admin` — Next.js (or static build served by nginx)

**Not in Docker:**

- React Native mobile app — runs locally via Expo or emulator, connects to VPS API

**Volumes:**

- `pgdata` — persists PostgreSQL data

**Ports exposed on VPS:**

- `8080` — Go API
- `3000` — Admin web (Next.js)
- `5432` — PostgreSQL (optional, only if you want direct DB access)

**Environment variables (.env):**

- `DATABASE_URL` — postgres connection string
- `SMTP_HOST` — smtp.gmail.com (or your email provider)
- `SMTP_PORT` — 587
- `SMTP_USER` — your email address
- `SMTP_PASS` — app password
- `API_PORT` — 8080