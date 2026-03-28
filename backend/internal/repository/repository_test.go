//go:build integration

package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*Repository, func()) {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)

	err = pool.Ping(ctx)
	require.NoError(t, err, "cannot ping test database")

	repo := New(pool)
	return repo, func() { pool.Close() }
}

func TestGetUserByID_ExistingUser_ReturnsUser(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Get first seeded user
	var userID string
	err := repo.pool.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID)
	require.NoError(t, err)

	user, err := repo.GetUserByID(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, userID, user.ID)
	assert.NotEmpty(t, user.Username)
	assert.NotEmpty(t, user.Email)
}

func TestGetUserByID_NonexistentUser_ReturnsError(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	user, err := repo.GetUserByID(context.Background(), "00000000-0000-0000-0000-000000000000")

	assert.Nil(t, user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "getting user by id")
}

func TestGetActiveChallenge_ReturnsActiveChallenge(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	challenge, err := repo.GetActiveChallenge(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "active", challenge.Status)
	assert.NotEmpty(t, challenge.ID)
}

func TestInsertGemHistory_ValidInput_Succeeds(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	var userID string
	err := repo.pool.QueryRow(ctx, `SELECT id FROM users LIMIT 1`).Scan(&userID)
	require.NoError(t, err)

	gameName := "Test Game"
	err = repo.InsertGemHistory(ctx, userID, "gameplay", 100, &gameName)

	assert.NoError(t, err)

	// Verify it was inserted
	var count int
	err = repo.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM gem_history WHERE user_id = $1 AND source = 'gameplay' AND amount = 100 AND game_name = 'Test Game'`,
		userID,
	).Scan(&count)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1)
}

func TestUpsertLeaderboardEntry_NewEntry_CreatesWithGems(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create a fresh user to avoid conflicts with seeded data
	var userID string
	err := repo.pool.QueryRow(ctx,
		`INSERT INTO users (username, email) VALUES ('testuser_upsert', 'testuser_upsert@test.com') RETURNING id`,
	).Scan(&userID)
	require.NoError(t, err)
	defer repo.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)

	challenge, err := repo.GetActiveChallenge(ctx)
	require.NoError(t, err)

	weeklyGems, err := repo.UpsertLeaderboardEntry(ctx, challenge.ID, userID, "te****t", 500)

	assert.NoError(t, err)
	assert.Equal(t, 500, weeklyGems)
}

func TestUpsertLeaderboardEntry_ExistingEntry_IncrementsGems(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()

	var userID string
	err := repo.pool.QueryRow(ctx,
		`INSERT INTO users (username, email) VALUES ('testuser_incr', 'testuser_incr@test.com') RETURNING id`,
	).Scan(&userID)
	require.NoError(t, err)
	defer repo.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)

	challenge, err := repo.GetActiveChallenge(ctx)
	require.NoError(t, err)

	// First upsert
	gems1, err := repo.UpsertLeaderboardEntry(ctx, challenge.ID, userID, "te****r", 300)
	require.NoError(t, err)
	assert.Equal(t, 300, gems1)

	// Second upsert — should increment
	gems2, err := repo.UpsertLeaderboardEntry(ctx, challenge.ID, userID, "te****r", 200)
	assert.NoError(t, err)
	assert.Equal(t, 500, gems2)
}

func TestMaskName_LongName_MasksCorrectly(t *testing.T) {
	assert.Equal(t, "ja****s", maskName("james"))
	assert.Equal(t, "al****r", maskName("alexander"))
}

func TestMaskName_ShortName_MasksCorrectly(t *testing.T) {
	assert.Equal(t, "m****", maskName("mia"))
	assert.Equal(t, "l****", maskName("leo"))
	assert.Equal(t, "e****", maskName("emma"))
}

func TestGetTodayCheckin_ExistingDate_ReturnsConfig(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	today := time.Now().UTC()

	checkin, err := repo.GetTodayCheckin(ctx, today)

	// May or may not exist depending on seed data dates
	assert.NoError(t, err)
	if checkin != nil {
		assert.True(t, checkin.BaseGems > 0)
	}
}

func TestDisplayNameExists_ExistingName_ReturnsTrue(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	challenge, err := repo.GetActiveChallenge(ctx)
	require.NoError(t, err)

	// Get an existing display name from seeded data
	var displayName string
	err = repo.pool.QueryRow(ctx,
		`SELECT display_name FROM leaderboard_entries WHERE challenge_id = $1 LIMIT 1`,
		challenge.ID,
	).Scan(&displayName)
	require.NoError(t, err)

	exists, err := repo.DisplayNameExists(ctx, challenge.ID, displayName)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestDisplayNameExists_NewName_ReturnsFalse(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	challenge, err := repo.GetActiveChallenge(ctx)
	require.NoError(t, err)

	exists, err := repo.DisplayNameExists(ctx, challenge.ID, "zz****zzz999")
	assert.NoError(t, err)
	assert.False(t, exists)
}
