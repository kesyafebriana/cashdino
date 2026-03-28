package repository

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kesyafebriana/cashdino/backend/internal/model"
)

// DBTX is satisfied by both *pgxpool.Pool and pgx.Tx.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txKey struct{}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// getDB returns the transaction from context if present, otherwise the pool.
func (r *Repository) getDB(ctx context.Context) DBTX {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return r.pool
}

// RunInTx executes fn within a database transaction.
// If fn returns an error the transaction is rolled back; otherwise it is committed.
func (r *Repository) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	var u model.User
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT id, username, email, created_at FROM users WHERE id = $1`,
		userID,
	).Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting user by id: %w", model.ErrNotFound)
		}
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	return &u, nil
}

func (r *Repository) InsertGemHistory(ctx context.Context, userID, source string, amount int, gameName *string) error {
	_, err := r.getDB(ctx).Exec(ctx,
		`INSERT INTO gem_history (user_id, source, amount, game_name) VALUES ($1, $2, $3, $4)`,
		userID, source, amount, gameName,
	)
	if err != nil {
		return fmt.Errorf("inserting gem history: %w", err)
	}
	return nil
}

func (r *Repository) GetActiveChallenge(ctx context.Context) (*model.WeeklyChallenge, error) {
	var wc model.WeeklyChallenge
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT id, start_time, end_time, status FROM weekly_challenges WHERE status = 'active' LIMIT 1`,
	).Scan(&wc.ID, &wc.StartTime, &wc.EndTime, &wc.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting active challenge: %w", model.ErrNotFound)
		}
		return nil, fmt.Errorf("getting active challenge: %w", err)
	}
	return &wc, nil
}

// UpsertLeaderboardEntry increments weekly_gems for an existing entry or creates a new one.
// Returns the updated weekly_gems total.
func (r *Repository) UpsertLeaderboardEntry(ctx context.Context, challengeID, userID, displayName string, amount int) (int, error) {
	var weeklyGems int
	err := r.getDB(ctx).QueryRow(ctx,
		`INSERT INTO leaderboard_entries (challenge_id, user_id, weekly_gems, first_gem_earned_at, display_name)
		 VALUES ($1, $2, $3, NOW(), $4)
		 ON CONFLICT (challenge_id, user_id)
		 DO UPDATE SET
		   weekly_gems = leaderboard_entries.weekly_gems + EXCLUDED.weekly_gems,
		   first_gem_earned_at = COALESCE(leaderboard_entries.first_gem_earned_at, NOW())
		 RETURNING weekly_gems`,
		challengeID, userID, amount, displayName,
	).Scan(&weeklyGems)
	if err != nil {
		return 0, fmt.Errorf("upserting leaderboard entry: %w", err)
	}
	return weeklyGems, nil
}

// DisplayNameExists checks if a display_name is already taken for a given challenge.
func (r *Repository) DisplayNameExists(ctx context.Context, challengeID, displayName string) (bool, error) {
	var exists bool
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM leaderboard_entries WHERE challenge_id = $1 AND display_name = $2)`,
		challengeID, displayName,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking display name existence: %w", err)
	}
	return exists, nil
}

// GetTodayCheckin returns today's daily_checkins config, or nil if none exists.
func (r *Repository) GetTodayCheckin(ctx context.Context, today time.Time) (*model.DailyCheckin, error) {
	var dc model.DailyCheckin
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT id, date, base_gems, streak_multiplier, is_active
		 FROM daily_checkins WHERE date = $1`,
		today.Format("2006-01-02"),
	).Scan(&dc.ID, &dc.Date, &dc.BaseGems, &dc.StreakMultiplier, &dc.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting today checkin: %w", err)
	}
	return &dc, nil
}

// HasCheckedInToday checks if a user already has a record for the given checkin_id.
func (r *Repository) HasCheckedInToday(ctx context.Context, userID, checkinID string) (bool, error) {
	var exists bool
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_daily_checkins WHERE user_id = $1 AND checkin_id = $2)`,
		userID, checkinID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking if user checked in today: %w", err)
	}
	return exists, nil
}

// GetLastCheckin returns the user's most recent user_daily_checkins record.
func (r *Repository) GetLastCheckin(ctx context.Context, userID string) (*model.UserDailyCheckin, error) {
	var udc model.UserDailyCheckin
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT id, user_id, checkin_id, gems_earned, current_streak, checked_in_at
		 FROM user_daily_checkins
		 WHERE user_id = $1
		 ORDER BY checked_in_at DESC
		 LIMIT 1`,
		userID,
	).Scan(&udc.ID, &udc.UserID, &udc.CheckinID, &udc.GemsEarned, &udc.CurrentStreak, &udc.CheckedInAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting last checkin: %w", err)
	}
	return &udc, nil
}

// GetCheckinDate returns the date for a given checkin_id.
func (r *Repository) GetCheckinDate(ctx context.Context, checkinID string) (time.Time, error) {
	var date time.Time
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT date FROM daily_checkins WHERE id = $1`,
		checkinID,
	).Scan(&date)
	if err != nil {
		return time.Time{}, fmt.Errorf("getting checkin date: %w", err)
	}
	return date, nil
}

func (r *Repository) InsertUserDailyCheckin(ctx context.Context, userID, checkinID string, gemsEarned, currentStreak int) error {
	_, err := r.getDB(ctx).Exec(ctx,
		`INSERT INTO user_daily_checkins (user_id, checkin_id, gems_earned, current_streak)
		 VALUES ($1, $2, $3, $4)`,
		userID, checkinID, gemsEarned, currentStreak,
	)
	if err != nil {
		return fmt.Errorf("inserting user daily checkin: %w", err)
	}
	return nil
}

// GenerateDisplayName creates a masked display name and resolves collisions.
func (r *Repository) GenerateDisplayName(ctx context.Context, challengeID, username string) (string, error) {
	masked := maskName(username)
	for i := 0; i < 10; i++ {
		exists, err := r.DisplayNameExists(ctx, challengeID, masked)
		if err != nil {
			return "", fmt.Errorf("generating display name: %w", err)
		}
		if !exists {
			return masked, nil
		}
		masked = masked + fmt.Sprintf("%d", rand.Intn(10))
	}
	return masked, nil
}

func maskName(name string) string {
	runes := []rune(name)
	if len(runes) <= 4 {
		return string(runes[0:1]) + "****"
	}
	return string(runes[0:2]) + "****" + string(runes[len(runes)-1:])
}

// GetTop99Entries returns the top 99 leaderboard entries for a challenge,
// ordered by weekly_gems DESC, first_gem_earned_at ASC.
func (r *Repository) GetTop99Entries(ctx context.Context, challengeID string) ([]model.LeaderboardEntry, error) {
	rows, err := r.getDB(ctx).Query(ctx,
		`SELECT id, challenge_id, user_id, weekly_gems, first_gem_earned_at, display_name
		 FROM leaderboard_entries
		 WHERE challenge_id = $1 AND weekly_gems > 0
		 ORDER BY weekly_gems DESC, first_gem_earned_at ASC
		 LIMIT 99`,
		challengeID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying top 99 entries: %w", err)
	}
	defer rows.Close()

	var entries []model.LeaderboardEntry
	for rows.Next() {
		var e model.LeaderboardEntry
		if err := rows.Scan(&e.ID, &e.ChallengeID, &e.UserID, &e.WeeklyGems, &e.FirstGemEarnedAt, &e.DisplayName); err != nil {
			return nil, fmt.Errorf("scanning leaderboard entry: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// GetUserEntry returns a specific user's leaderboard entry for a challenge, or nil if not found.
func (r *Repository) GetUserEntry(ctx context.Context, challengeID, userID string) (*model.LeaderboardEntry, error) {
	var e model.LeaderboardEntry
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT id, challenge_id, user_id, weekly_gems, first_gem_earned_at, display_name
		 FROM leaderboard_entries
		 WHERE challenge_id = $1 AND user_id = $2`,
		challengeID, userID,
	).Scan(&e.ID, &e.ChallengeID, &e.UserID, &e.WeeklyGems, &e.FirstGemEarnedAt, &e.DisplayName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user entry: %w", err)
	}
	return &e, nil
}

// GetLastCompletedChallenge returns the most recent completed challenge, or nil if none.
func (r *Repository) GetLastCompletedChallenge(ctx context.Context) (*model.WeeklyChallenge, error) {
	var wc model.WeeklyChallenge
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT id, start_time, end_time, status
		 FROM weekly_challenges
		 WHERE status = 'completed'
		 ORDER BY end_time DESC
		 LIMIT 1`,
	).Scan(&wc.ID, &wc.StartTime, &wc.EndTime, &wc.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting last completed challenge: %w", err)
	}
	return &wc, nil
}

// GetTop99Results returns the top 99 results for a completed challenge.
func (r *Repository) GetTop99Results(ctx context.Context, challengeID string) ([]model.WeeklyChallengeResult, error) {
	rows, err := r.getDB(ctx).Query(ctx,
		`SELECT id, challenge_id, user_id, final_rank, final_gems, display_name
		 FROM weekly_challenge_results
		 WHERE challenge_id = $1
		 ORDER BY final_rank ASC
		 LIMIT 99`,
		challengeID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying top 99 results: %w", err)
	}
	defer rows.Close()

	var results []model.WeeklyChallengeResult
	for rows.Next() {
		var res model.WeeklyChallengeResult
		if err := rows.Scan(&res.ID, &res.ChallengeID, &res.UserID, &res.FinalRank, &res.FinalGems, &res.DisplayName); err != nil {
			return nil, fmt.Errorf("scanning challenge result: %w", err)
		}
		results = append(results, res)
	}
	return results, rows.Err()
}

// GetUserResult returns a specific user's result for a challenge, or nil if not found.
func (r *Repository) GetUserResult(ctx context.Context, challengeID, userID string) (*model.WeeklyChallengeResult, error) {
	var result model.WeeklyChallengeResult
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT id, challenge_id, user_id, final_rank, final_gems, display_name
		 FROM weekly_challenge_results
		 WHERE challenge_id = $1 AND user_id = $2`,
		challengeID, userID,
	).Scan(&result.ID, &result.ChallengeID, &result.UserID, &result.FinalRank, &result.FinalGems, &result.DisplayName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting user result: %w", err)
	}
	return &result, nil
}

// GetCampaignByChallenge returns the active/scheduled reward campaign for a challenge, or nil.
func (r *Repository) GetCampaignByChallenge(ctx context.Context, challengeID string) (*model.RewardCampaign, error) {
	var c model.RewardCampaign
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT id, challenge_id, name, banner_image, rules, status
		 FROM reward_campaigns
		 WHERE challenge_id = $1 AND status IN ('active', 'scheduled')
		 LIMIT 1`,
		challengeID,
	).Scan(&c.ID, &c.ChallengeID, &c.Name, &c.BannerImage, &c.Rules, &c.Status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting campaign by challenge: %w", err)
	}
	return &c, nil
}

// GetRewardTypesByIDs returns reward types matching the given IDs.
func (r *Repository) GetRewardTypesByIDs(ctx context.Context, ids []string) ([]model.RewardType, error) {
	rows, err := r.getDB(ctx).Query(ctx,
		`SELECT id, campaign_id, name, type, value, image, stock
		 FROM reward_types
		 WHERE id = ANY($1)`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("querying reward types: %w", err)
	}
	defer rows.Close()

	var types []model.RewardType
	for rows.Next() {
		var rt model.RewardType
		if err := rows.Scan(&rt.ID, &rt.CampaignID, &rt.Name, &rt.Type, &rt.Value, &rt.Image, &rt.Stock); err != nil {
			return nil, fmt.Errorf("scanning reward type: %w", err)
		}
		types = append(types, rt)
	}
	return types, rows.Err()
}

// GetResultRewards returns reward distributions grouped by user_id for a given challenge.
func (r *Repository) GetResultRewards(ctx context.Context, challengeID string) (map[string][]model.RewardInfo, error) {
	rows, err := r.getDB(ctx).Query(ctx,
		`SELECT rd.user_id, rt.name, rt.type, rt.value, rt.image
		 FROM reward_distributions rd
		 JOIN reward_types rt ON rd.reward_type_id = rt.id
		 JOIN reward_campaigns rc ON rd.campaign_id = rc.id
		 WHERE rc.challenge_id = $1 AND rd.status = 'delivered'`,
		challengeID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying result rewards: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]model.RewardInfo)
	for rows.Next() {
		var userID string
		var ri model.RewardInfo
		if err := rows.Scan(&userID, &ri.Name, &ri.Type, &ri.Value, &ri.Image); err != nil {
			return nil, fmt.Errorf("scanning reward distribution: %w", err)
		}
		result[userID] = append(result[userID], ri)
	}
	return result, rows.Err()
}
