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
