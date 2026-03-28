# **Version number: V0.x.x**

| Version | time | Revisionist | Remark |
| --- | --- | --- | --- |
| V0.x.x | 2026/03/27 | @Kesya  | Add Weekly Challenge Leaderboard system with privacy-first design |

---

# **1. Overview (Why do it)**

## **1.1 Product Overview and Objectives**

### **1.1.1 Background**

- Source
    
    https://apps.apple.com/us/app/cash-giraffe-surveys-perks/id6740522980?see-all=reviews&platform=iphone
    
    https://www.trustpilot.com/review/cashgiraffe.com
    
    ![Screenshot 2026-03-27 at 20.00.24.png](attachment:1fba972c-118e-4f88-9d9d-ff09c9eb9229:Screenshot_2026-03-27_at_20.00.24.png)
    
    ![Screenshot 2026-03-27 at 19.58.45.png](attachment:d4f7a749-8e54-4fd5-8dd8-0b2975c6d748:Screenshot_2026-03-27_at_19.58.45.png)
    
    ![Screenshot 2026-03-27 at 19.58.54.png](attachment:e8eb161e-05c8-4617-b416-3acf41fdf034:Screenshot_2026-03-27_at_19.58.54.png)
    
    ![Screenshot 2026-03-27 at 20.01.11.png](attachment:5d4224fa-5b8d-4e65-bb25-8da52ed91baa:Screenshot_2026-03-27_at_20.01.11.png)
    

Based on user feedback from platforms such as app store reviews and Trustpilot, a recurring concern is that:

- The amount of gems earned per task is relatively small
- Users need to spend a **significant amount of time playing games** before reaching a payout threshold
- Even with existing features such as daily check-in bonuses, the overall progression still feels slow

As a result, users may feel that:

- the effort required is too high compared to the reward
- long play sessions are not sufficiently motivating
- there is no clear short-term goal to work towards

---

### **1.1.2 Product Overview**

The Weekly Challenge Leaderboard system introduces a **weekly progression layer** that:

- Aggregates user gems into a weekly score
- Ranks users through a leaderboard
- Provides real-time feedback on progress
- Integrates with existing daily check-in
- Supports **rewardable leaderboard campaigns** configurable by admin

The system is designed to be:

- lightweight (no heavy social features)
- privacy-first (masked identities)
- optional (opt-in participation)

It reframes longer play sessions from a cost into a motivating progression and competitive challenge.

---

### **1.1.3 Product Goals**

### Business objectives

- Increase DAU (**target: +10–20% uplift, *to be validated***)
- Improve D7 retention
- Increase weekly engagement (sessions per user per week)
- Increase reward conversion rate
- Increase leaderboard participation rate

---

### User goal

- Understand their progress clearly within a **weekly cycle**
- Feel motivated to continue using the app through **competition and ranking**
- **Perceive longer playtime as meaningful progress rather than effort cost**
- See how close they are to improving their rank or reaching a goal
- Have confidence that their actions lead to visible outcomes

---

### **1.1.4 Target Users**

- Existing Cash Giraffe users earning gems
- Users who engage with daily check-in
- Users who need additional motivation to continue
- Loyal users who spend significant time in the app and aim to maximize rewards

---

## 1.2 Glossary

- **Weekly Challenge** — 7-day cycle (Monday 00:00 UTC → Sunday 23:59 UTC)
- **Weekly Gems** — total gems earned during the active cycle
- **Leaderboard** — global ranked list by Weekly Gems
- **Masked Identity** — obfuscated display name (e.g., `ke****m`)
- **Reward Campaign** — admin-configured rank-based rewards
- **Reward Rule** — rank range + reward type + value

---

## 1.3 Roles and Permissions

**User**

- View weekly challenge progress and rank
- View leaderboard

**Admin**

- Configure Weekly Challenge (start/end time)
- Create/manage Reward Campaigns (rank-based rules)
- Configure claim email for non-gem rewards

---

# 2. Product Description (What to do)

## 2.1 Product Requirements

- Track weekly gems per user (automatic)
- Rank all users on global leaderboard
- Masked identities by default
- Integrate with daily check-in
- Admin-configurable rank-based reward campaigns
- Gem rewards: auto-credit
- Non-gem rewards: send claim email

---

## 2.2 Overall Process

### 2.2.1 Main Flow

`Earn gems → Weekly gems updated → Rank updated → View leaderboard → Week ends → Rewards distributed → New cycle`

### 2.2.2 Sub-processes

**A — Gem Attribution**

- Gem earned (games, check-in, surveys, referrals)
- `gem_history` record created
- `leaderboard_entries.weekly_gems` incremented

**B — Weekly Reset**

- Cron at Sunday 23:59 UTC
- Final ranks → `weekly_challenge_results`
- Rewards distributed (if campaign active)
- `weekly_gems` reset to 0
- New cycle auto-created

**C — Reward Distribution**

- Week ends → match users to rank-range rules
- Gems → auto-credit to `gem_history` record with `source = 'reward'`
- Non-gems → claim email sent to `users.email`
- Outside all ranges → no reward

### 2.2.3 State Transitions

- `New Cycle` → *(earns gems)* → `Ranked` → *(week ends)* → `Rewarded` (if in range) → `New Cycle`

---

## 2.3 Global Rules

### 2.3.1 Exception Handling

| Scenario | Behavior |
| --- | --- |
| Network failure | Cached leaderboard + retry button |
| Sync delay | Optimistic local increment |
| Ranking service down | Show user's own gems + "Temporarily unavailable" |
| Claim email fails | Retry 3× / 24 h → notify admin |

### 2.3.2 List Rules

- Sorted by `weekly_gems` DESC
- Tie-break: earliest `first_gem_earned_at`
- Top 99 + current user's sticky footer (if outside Top 99)
- Onscroll pagination
- 0 gems = not shown

### 2.3.3 Interactions

**User highlight:** accent background on own row; if outside Top 99 → sticky footer showing `"99+"`

**Countdown:**

- `≥ 24 h` → `"Xd Xh Xm left"`
- `< 24 h` → `"Xh Xm left"` (orange)
- `< 1 h` → `"Xm left — Final hour!"` (red)

**Rank movement:** ▲ green / ▼ red / — unchanged

---

## 2.4 Feature List (MVP)

| # | Feature | Priority |
| --- | --- | --- |
| F-01 | Weekly gem tracking (automatic, all users) | P0 |
| F-02 | Rank (derived from query row position — not stored) | P0 |
| F-03 | Weekly auto-reset (Monday 00:00 UTC) | P0 |
| F-04 | Leaderboard screen (Top 99 + own position or 99+) | P0 |
| F-05 | Last week leaderboard (archived, with reward icons) | P0 |
| F-06 | Challenge banner on Home | P0 |
| F-06 | Masked display names | P0 |
| F-07 | Countdown timer | P1 |
| F-08 | Rank movement indicators | P1 |
| F-09 | Onboarding modal (one-time tour) | P1 |
| F-10 | Admin: challenge config | P1 |
| F-10 | Admin: reward campaign (rank-based rules) | P0 |
| F-11 | Reward distribution (gems auto-credit) | P0 |
| F-12 | Reward distribution (non-gems → claim email) | P0 |
| F-13 | Reward image (PNG stored in DB per reward type) | P1 |

**Excluded from MVP (future):**

- Custom aliases + profanity filter
- Shareable summary card
- `{{claim_link}}` in email templates
- User reward history screen

---

# 3. Functional Requirements (How to do it)

## 3.1 Weekly Challenge Module

### 3.1.1 Description

- Core engine: gem tracking over 7-day cycle
- All users participate automatically
- Powers banner, leaderboard, countdown

### 3.1.2 User Stories

| ID | Story | Acceptance Criteria |
| --- | --- | --- |
| US-01 | See weekly rank on Home | Banner: rank (or 99+), gems, countdown |
| US-02 | Know gems needed to rank up | "X gems to #Y" on banner + leaderboard (only if in Top 99) |
| US-03 | See when challenge ends | Countdown with urgency styling < 24 h |
| US-04 | Clean weekly reset | Gems → 0, new cycle auto-starts |

### 3.1.3 Interface

**Challenge Banner (Home):**

`🏆 Weekly Challenge              2d 14h left
Your Rank: #12    Weekly Gems: 4,320   ▲3
42 gems to reach #11
                    [ View Leaderboard → ]`

### 3.1.4 Business Logic

- Gem event → increment `weekly_gems`
- Rank is **not stored** — derived from query row position:
`SELECT * FROM leaderboard_entries WHERE challenge_id = ? AND weekly_gems > 0 ORDER BY weekly_gems DESC, first_gem_earned_at ASC LIMIT 99`
- Row index = rank (1, 2, 3…) — assigned by FE or API response
- Current user outside Top 99 → show `"99+"`
- Client polls → updates banner

### 3.1.5 Exception Flows

| Scenario | Handling |
| --- | --- |
| No active challenge | Banner: "Next challenge starts in X hours" |
| Sync delay > 60 s | Last known state + "Syncing…" |
| 0 gems | "Earn gems to get ranked!" |
| Challenge ends while in-app | Modal: final rank → transition to new cycle |

---

## 3.2 Leaderboard Module

### 3.2.1 Description

- Global ranked list by weekly gems
- Primary competitive interface

### 3.2.2 User Stories

| ID | Story | Acceptance Criteria |
| --- | --- | --- |
| US-06 | See ranked users | Top 99 with rank, masked name, gems |
| US-07 | See my position | If in Top 99: highlighted row. If outside: sticky footer `"99+"` |

### 3.2.3 Interface

`🏆 Weekly Challenge              2d 14h left
#   Name          Gems      Δ
🥇  ke****m       12,450    ▲2
🥈  gi****21      11,980    —
🥉  mi****a       10,240    ▼1
4   xa****7        9,870    ▲5
...
99  to****3          120    ▼2
┌────────────────────────────────────────┐
│ 99+  YOU ★    80 gems                  │
└────────────────────────────────────────┘`

- Top 3: medal icons
- If user is in Top 99: highlighted row with rank, gems, gap to next
- If user is outside Top 99: sticky footer showing `"99+"` + gems
- No Onscroll pagination after 99 — fixed at 99
- **Prize banner** if campaign active:
    - PNG image (admin uploads via `reward_campaigns.banner_image`)
    - Displayed at top of leaderboard, above the list
    - Tappable → opens **Reward Details Modal**
    - Small "See all rewards →" link below banner (also opens modal)
- Pull-to-refresh

**Reward Details Modal (tap banner or "See all rewards"):**

```jsx
┌──────────────────────────────────────────────┐
│  🏆 This Week's Rewards                 [✕] │
│──────────────────────────────────────────────│
│  #1         [img] 10,000 gems                │
│             [img] $10 Amazon Gift Card        │
│                                              │
│  #2–5       [img] $10 Amazon Gift Card        │
│             [img] 2,000 gems                  │
│                                              │
│  #6–10      [img] $5 Starbucks Gift Card      │
└──────────────────────────────────────────────┘
```

- Lists all reward rules for the active campaign
- Each row: rank range + reward image (from `reward_types.image`) + reward name + value
- Close via ✕ or tap outside

---

## 3.3 Last Week Leaderboard

### 3.3.1 Description

- Archived, read-only view of the previous week's final leaderboard
- Shows final ranks, gems, and reward received per user
- Accessible via tab toggle on leaderboard screen: `[ This Week ] [ Last Week ]`

### 3.3.2 User Stories

| ID | Story | Acceptance Criteria |
| --- | --- | --- |
| US-09 | See last week's final results | Tab shows archived Top 99 and current user with final rank + gems  |
| US-10 | See what rewards were given | Gift icon 🎁 on rows that received a reward; tappable |
| US-11 | See reward details | Tap gift icon → bubble/tooltip shows: reward type, value, image |

### 3.3.3 Interface

`[ This Week ]  [ Last Week ]

Last Week Results — Mar 17–23
#   Name          Gems       🎁
🥇  ke****m       12,450     🎁
🥈  gi****21      11,980     🎁
🥉  mi****a       10,240     🎁
4   xa****7        9,870     🎁
5   lo****8        8,200     🎁
6   pe****1        7,100     🎁
...
99  to****3          120`

**Tap 🎁 on rank #1 → bubble:**

`┌────────────────────────────────────────────┐
│          🏆 Rank #1 Reward                  │
│    [gems image]  |  [gift card image]       │
│  💎 10,000 gems   |   $10 Amazon Gift Card  │
└────────────────────────────────────────────┘`

**Tap 🎁 on rank #2 → bubble:**

`┌─────────────────────────┐
│  🏆 Rank #2 Reward      │
│  [gift card image]       │
│  $10 Amazon Gift Card    │
└─────────────────────────┘`

- 🎁 icon only shown on rows where a reward was distributed
- Rows with no reward → no icon
- Bubble shows: reward type name, value, and image (from `reward_types.image` URL)
- Tap outside bubble to dismiss
- Current user's row still highlighted if in Top 99

### 3.3.4 Data Source

- Reads from `weekly_challenge_results` (final rank + gems)
- Joins with `reward_distributions` → `reward_types` to get reward info + image
- No writes — fully read-only

---

## 3.4 Reward Campaign Module

### 3.4.1 Description

- Admin creates rank-range rules per weekly challenge
- Auto-distributed at week end

### 3.4.2 Reward Rule Format

**Admin defines reward types first, then assigns them to rank ranges.**

**Step 1 — Define reward types:**

| Reward Type | Image | Value | Stock |
| --- | --- | --- | --- |
| Gems | 💎 (default icon) | 10000 | 1 |
| Gems | 💎 (default icon) | 2000 | 4 |
| Gift Card A $10 | uploaded PNG | $10 | 2 |
| Gift Card B $5 | uploaded PNG | $5 | 1 |
- Each reward type: `name`, `image` (PNG URL), `type`, `value`, `stock`
- `stock` = total number of this reward available for the campaign

**Step 2 — Assign to rank ranges:**

`1-1    reward: Gift Card A ($10) + Gems (10,000)
2-5    reward: Gift Card A ($10)
6-10   reward: Gift Card B ($5)`

- Rank 1 → $10 Gift Card A (claim manually via email) + 10K gems (auto-credited)
- Rank 2–5 → $10 Gift Card A (claim manually via email) + 2K gems (auto-credited, must have 4 stock 2000 gems)
- Rank 6–10 → $5 Gift Card B (claim manually via email)
- Rank 11+ → no reward
- No overlapping ranges (validated on save)
- System validates stock: total users in range must not exceed reward stock (e.g., Gift Card A stock = 5 covers rank 1 + ranks 2–5)

### 3.3.3 Admin Campaign Fields

- `Campaign name` — internal label
- `Linked challenge` — select week
- `Banner image` — PNG URL displayed on leaderboard (tappable → reward details modal)
- `Reward types` — repeatable:
    - `Name` — e.g., "Gift Card A"
    - `Type` — gems / gift_card / cash / other
    - `Value` — amount
    - `Image` — PNG URL
    - `Stock` — total available
- `Reward rules` — repeatable rows:
    - `Rank from` — INT
    - `Rank to` — INT
    - `Reward types` — select one or more reward types for this range
- `Claim email subject` — for non-gem rewards
- `Claim email body` — supports `{{username}}`, `{{rank}}`, `{{reward_type}}`, `{{reward_value}}`, `{{reward_image}}`
- `Status` — draft / scheduled / active / completed

### 3.4.4 Distribution Flow

For each rank range rule, for each reward type in that range:

**If reward type = gems:**

1. Final ranks locked
2. Match users by rank range
3. Write `gem_history` with `source = 'reward'`
4. Create `reward_distributions` record → `status = 'delivered'`

**If reward type ≠ gems:**

1. Final ranks locked
2. Match users by rank range
3. Send claim email to `users.email` using admin template
4. Replace `{{username}}`, `{{rank}}`, `{{reward_type}}`, `{{reward_value}}`, `{{reward_image}}`
5. Create `reward_distributions` record → `status = 'delivered'`
6. If email fails → `status = 'failed'` → retry 3× → notify admin

**A single user can receive multiple `reward_distributions` records** (one per reward type in their rank range).

**Users outside all rank ranges → no records created.**

### 3.4.5 Stock Validation

- On campaign save: system calculates total users per reward type across all ranges
    - e.g., Gift Card A used in range 1–1 (1 user) + range 2–5 (4 users) = 5 needed → stock must be ≥ 5
- If stock insufficient → admin shown validation error, save blocked
- Stock decremented on distribution

---

## 3.5 Onboarding Module

### 3.4.1 Description

- One-time modal shown to all existing users after app update
- Introduces the Weekly Challenge Leaderboard
- Guides user to visit the leaderboard

### 3.4.2 Trigger

- Shown once on first app open after feature launch
- Flag `onboarding_seen` stored locally (device) — no DB needed
- If dismissed, never shown again

### 3.4.3 Flow

**Screen 1 — What's New**

- Headline: "Introducing Weekly Challenges"
- Body: "Every gem you earn now counts toward a weekly ranking. Reach the top and earn rewards!"
- Illustration: leaderboard preview

**Screen 2 — How it Works**

- Headline: "How it Works"
- Body: "Gems from games, check-ins, and surveys all count. Rankings reset every Monday."
- Illustration: gem sources → leaderboard

**Screen 3 — CTA**

- Headline: "See Where You Stand"
- Body: "Your name is masked for privacy. Check the leaderboard now!"
- CTA: **[ View Leaderboard → ]** (navigates to leaderboard screen)
- Secondary: **[ Got it ]** (dismisses modal, stays on Home)

### 3.4.4 Rules

- Max 3 screens, swipeable
- Skip button on every screen
- No login or action required — purely informational
- Does not block app usage

---

## 3.6 Masked Identity

### Logic

- First 2 chars + `***` + last char (e.g., `james` → `ja****s`)
- Username ≤ 4 chars → first char + `***`
- Random digit appended on collision
- Generated once at leaderboard entry creation

---

# 4. Non-functional Requirements

## 4.1 Security & Compliance

- **No opt-in required** — gem data already consented at registration (GDPR Art. 6(1)(a) + Art. 6(1)(b)); leaderboard is a new display of existing data
- **Masked names by default** — no PII exposed via leaderboard API
- **Data minimization** — API returns only `display_name`, `weekly_gems`, `rank`, `rank_delta`
- **Claim emails are transactional** — exempt from marketing opt-out; no promotional content
- **Anti-fraud** — flag accounts > 3σ gem rate; exclude confirmed fraud

---

## 4.2 Performance

| Requirement | Target |
| --- | --- |
| Leaderboard load | < 1.5 s (P95) |
| Weekly reset (1M users) | < 5 min |
| API response | < 200 ms (P95) |

---

## 4.3 Database Design

### A — Base Tables (simulated app data, might be simpler)

### `users`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | PK |
| `username` | VARCHAR(30) | Unique |
| `email` | VARCHAR(255) | Unique |
| `created_at` | TIMESTAMP | — |

### `gem_history`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | PK |
| `user_id` | UUID | FK → `users` |
| `source` | ENUM | `gameplay` / `daily_checkin` / `survey` / `referral` / `boost` / `reward` / `payout` |
| `amount` | INT | Positive = earned, negative = spent/withdrawn |
| `game_name` | VARCHAR(100) | Nullable; only if source = `gameplay` |
| `created_at` | TIMESTAMP | — |
- **Index:** `(user_id, created_at DESC)`

### `daily_checkins`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | PK |
| `date` | DATE | Calendar date |
| `base_gems` | INT | Gems for that day |
| `streak_multiplier` | DECIMAL(3,2) | e.g., 1.00, 1.50, 2.00 |
| `is_active` | BOOLEAN | Admin toggle |

### `user_daily_checkins`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | PK |
| `user_id` | UUID | FK → `users` |
| `checkin_id` | UUID | FK → `daily_checkins` |
| `gems_earned` | INT | base_gems × streak_multiplier |
| `current_streak` | INT | Consecutive days |
| `checked_in_at` | TIMESTAMP | — |
- **Unique:** `(user_id, checkin_id)`

---

### B — Weekly Challenge Tables

### `weekly_challenges`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | PK |
| `start_time` | TIMESTAMP | Monday 00:00 UTC |
| `end_time` | TIMESTAMP | Sunday 23:59 UTC |
| `status` | ENUM | scheduled / active / completed |

### `leaderboard_entries`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | PK |
| `challenge_id` | UUID | FK → `weekly_challenges` |
| `user_id` | UUID | FK → `users` |
| `weekly_gems` | INT | Default 0 |
| `first_gem_earned_at` | TIMESTAMP | Nullable, tie-break |
| `display_name` | VARCHAR(20) | Masked username |
- **Index:** `(challenge_id, weekly_gems DESC, first_gem_earned_at ASC)`
- **Unique:** `(challenge_id, user_id)`

### `weekly_challenge_results`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | PK |
| `challenge_id` | UUID | FK |
| `user_id` | UUID | FK |
| `final_rank` | INT | — |
| `final_gems` | INT | — |

---

### C — Reward Tables

### `reward_types`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | PK |
| `campaign_id` | UUID | FK → `reward_campaigns` |
| `name` | VARCHAR(100) | e.g., "Gift Card A", "10K Gems" |
| `type` | ENUM | gems / gift_card / cash / other |
| `value` | DECIMAL | Amount |
| `image` | VARCHAR(500) | URL to PNG; nullable for gems |
| `stock` | INT | Total available for this campaign |

### `reward_campaigns`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | PK |
| `challenge_id` | UUID | FK |
| `name` | VARCHAR(100) | Internal label |
| `banner_image` | VARCHAR(500) | URL to prize banner PNG; shown on leaderboard |
| `rules` | JSONB | Rank-range rules referencing reward_types (see below) |
| `non_gem_claim_email_subject` | VARCHAR(255) | Admin-configurable |
| `non_gem_claim_email_body` | TEXT | Supports `{{username}}`, `{{rank}}`, `{{reward_type}}`, `{{reward_value}}`, `{{reward_image}}` |
| `status` | ENUM | draft / scheduled / active / completed |

**`rules` JSONB — supports multiple rewards per range:**

json

`[
  {
    "rank_from": 1,
    "rank_to": 1,
    "reward_type_ids": ["uuid-10k-gems", "uuid-gift-card-a"]
  },
  {
    "rank_from": 2,
    "rank_to": 5,
    "reward_type_ids": ["uuid-2k-gems", "uuid-gift-card-a"]
  },
  {
    "rank_from": 6,
    "rank_to": 10,
    "reward_type_ids": ["uuid-gift-card-b"]
  }
]`

- `reward_type_ids` is an **array** — one or more reward types per range
- No overlapping rank ranges (validated on save)
- Stock validated on save (total users across ranges ≤ stock per reward type)

### `reward_distributions`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | UUID | PK |
| `campaign_id` | UUID | FK |
| `user_id` | UUID | FK |
| `reward_type_id` | UUID | FK → `reward_types` |
| `status` | ENUM | pending / delivered / failed |
| `delivered_at` | TIMESTAMP | Nullable |
| `email_sent_at` | TIMESTAMP | Nullable; non-gem only |

---

## 4.5 System Integration

| Integration | Direction | Purpose |
| --- | --- | --- |
| Gem tracking system | Inbound | Gem events → update `weekly_gems` |
| Push notification service | Outbound | Weekly results, rewards |
| Email service (new) | Outbound | Claim emails for non-gem rewards |

---

# 5. Appendix

## 5.1 Acceptance Criteria & Test Points

| # | Test Case | Expected Result |
| --- | --- | --- |
| T-01 | Earn gems from gameplay (Add mock feature inside task page to add gems) | `weekly_gems` increments |
| T-02 | Complete daily check-in | Check-in gems added to weekly total |
| T-03 | View leaderboard | Sorted correctly; own row highlighted; masked names |
| T-04 | Countdown accuracy | Matches server time ±1 min |
| T-05 | Weekly reset | Gems → 0; new challenge created; results archived |
| T-06 | Tie-breaking | Earlier `first_gem_earned_at` ranks higher |
| T-07 | Masked display name | Correct format; no collisions |
| T-08 | Outside Top 99 | Sticky footer showing "99+" + gems |
| T-09 | Gem reward | Rank in range → gems auto-credited + gem_history |
| T-10 | Non-gem reward | Rank in range → claim email with correct template + reward image |
| T-11 | Outside all ranges | No reward record; no email; no gems |
| T-12 | Overlapping ranges | Admin save rejected with validation error |
| T-13 | Email failure | 3× retry → admin notified |
| T-14 | 0 gems user | Not shown on leaderboard |
| T-15 | Reward type image | PNG URL stored; displayed in reward bubble + modal + email |
| T-16 | Prize banner | PNG displayed at top of leaderboard when campaign active |
| T-17 | Prize banner tap | Opens reward details modal with all rank ranges + images |
| T-18 | No active campaign | No prize banner shown |
| T-19 | Last week leaderboard | Tab shows archived Top 99 with final ranks + gems |
| T-17 | Last week reward icon | 🎁 shown only on rows with reward; not shown on others |
| T-18 | Reward bubble | Tap 🎁 → bubble shows reward type, value, image; tap outside dismisses |
| T-19 | Onboarding — first open | Modal shown once after update; 3 screens |
| T-20 | Onboarding — dismiss | Never shown again after skip or "Got it" |
| T-21 | Onboarding — CTA | "View Leaderboard" navigates to leaderboard screen |

---

## 5.2 Future Development

| Feature | Notes |
| --- | --- |
| Toast feedback | Real-time toast after gem events ("+X gems · You're now #Y") |
| Custom aliases + profanity filter | User-chosen display names with moderation |
| Shareable summary card | End-of-week image with rank + gems |
| `{{claim_link}}` in email | Deep link or custom claim page |
| User reward history screen | View past rewards in-app |
| Multiple or grouped leaderboards | Based on user activity level/volume, so users can compete within smaller and more balanced groups |