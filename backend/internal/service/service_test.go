package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kesyafebriana/cashdino/backend/internal/model"
	"github.com/stretchr/testify/assert"
)

// --- Mock Repository ---

type mockRepo struct {
	getUserByID            func(ctx context.Context, userID string) (*model.User, error)
	insertGemHistory       func(ctx context.Context, userID, source string, amount int, gameName *string) error
	getActiveChallenge     func(ctx context.Context) (*model.WeeklyChallenge, error)
	upsertLeaderboardEntry func(ctx context.Context, challengeID, userID, displayName string, amount int) (int, error)
	generateDisplayName    func(ctx context.Context, challengeID, username string) (string, error)
	getTodayCheckin        func(ctx context.Context, today time.Time) (*model.DailyCheckin, error)
	hasCheckedInToday      func(ctx context.Context, userID, checkinID string) (bool, error)
	getLastCheckin         func(ctx context.Context, userID string) (*model.UserDailyCheckin, error)
	getCheckinDate         func(ctx context.Context, checkinID string) (time.Time, error)
	insertUserDailyCheckin func(ctx context.Context, userID, checkinID string, gemsEarned, currentStreak int) error
	runInTx                func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockRepo) GetUserByID(ctx context.Context, userID string) (*model.User, error) {
	return m.getUserByID(ctx, userID)
}
func (m *mockRepo) InsertGemHistory(ctx context.Context, userID, source string, amount int, gameName *string) error {
	return m.insertGemHistory(ctx, userID, source, amount, gameName)
}
func (m *mockRepo) GetActiveChallenge(ctx context.Context) (*model.WeeklyChallenge, error) {
	return m.getActiveChallenge(ctx)
}
func (m *mockRepo) UpsertLeaderboardEntry(ctx context.Context, challengeID, userID, displayName string, amount int) (int, error) {
	return m.upsertLeaderboardEntry(ctx, challengeID, userID, displayName, amount)
}
func (m *mockRepo) GenerateDisplayName(ctx context.Context, challengeID, username string) (string, error) {
	return m.generateDisplayName(ctx, challengeID, username)
}
func (m *mockRepo) GetTodayCheckin(ctx context.Context, today time.Time) (*model.DailyCheckin, error) {
	return m.getTodayCheckin(ctx, today)
}
func (m *mockRepo) HasCheckedInToday(ctx context.Context, userID, checkinID string) (bool, error) {
	return m.hasCheckedInToday(ctx, userID, checkinID)
}
func (m *mockRepo) GetLastCheckin(ctx context.Context, userID string) (*model.UserDailyCheckin, error) {
	return m.getLastCheckin(ctx, userID)
}
func (m *mockRepo) GetCheckinDate(ctx context.Context, checkinID string) (time.Time, error) {
	return m.getCheckinDate(ctx, checkinID)
}
func (m *mockRepo) InsertUserDailyCheckin(ctx context.Context, userID, checkinID string, gemsEarned, currentStreak int) error {
	return m.insertUserDailyCheckin(ctx, userID, checkinID, gemsEarned, currentStreak)
}
func (m *mockRepo) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.runInTx != nil {
		return m.runInTx(ctx, fn)
	}
	return fn(ctx)
}

// --- Helpers ---

func defaultMockRepo() *mockRepo {
	return &mockRepo{
		getUserByID: func(_ context.Context, _ string) (*model.User, error) {
			return &model.User{ID: "user-1", Username: "james", Email: "james@example.com"}, nil
		},
		insertGemHistory: func(_ context.Context, _, _ string, _ int, _ *string) error {
			return nil
		},
		getActiveChallenge: func(_ context.Context) (*model.WeeklyChallenge, error) {
			return &model.WeeklyChallenge{ID: "challenge-1", Status: "active"}, nil
		},
		upsertLeaderboardEntry: func(_ context.Context, _, _, _ string, amount int) (int, error) {
			return amount, nil
		},
		generateDisplayName: func(_ context.Context, _, _ string) (string, error) {
			return "ja****s", nil
		},
		getTodayCheckin: func(_ context.Context, _ time.Time) (*model.DailyCheckin, error) {
			return &model.DailyCheckin{ID: "checkin-1", BaseGems: 100, StreakMultiplier: 1.5, IsActive: true}, nil
		},
		hasCheckedInToday: func(_ context.Context, _, _ string) (bool, error) {
			return false, nil
		},
		getLastCheckin: func(_ context.Context, _ string) (*model.UserDailyCheckin, error) {
			return nil, nil
		},
		getCheckinDate: func(_ context.Context, _ string) (time.Time, error) {
			return time.Now().UTC().AddDate(0, 0, -1), nil
		},
		insertUserDailyCheckin: func(_ context.Context, _, _ string, _, _ int) error {
			return nil
		},
	}
}

func fixedNow() time.Time {
	return time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
}

func newTestService(repo *mockRepo) *Service {
	s := New(repo)
	s.now = fixedNow
	return s
}

// =====================================================================
// EarnGems tests
// =====================================================================

func TestEarnGems_ValidGameplay_ReturnsWeeklyGems(t *testing.T) {
	repo := defaultMockRepo()
	repo.upsertLeaderboardEntry = func(_ context.Context, _, _, _ string, amount int) (int, error) {
		return 1500, nil
	}
	svc := newTestService(repo)

	gameName := "Candy Crush"
	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{
		UserID:   "user-1",
		Source:   "gameplay",
		Amount:   500,
		GameName: &gameName,
	})

	assert.NoError(t, err)
	assert.Equal(t, "user-1", resp.UserID)
	assert.Equal(t, 1500, resp.WeeklyGems)
}

func TestEarnGems_EmptyUserID_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())

	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{
		UserID: "",
		Source: "gameplay",
		Amount: 100,
	})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "user_id is required")
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestEarnGems_InvalidSource_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())

	for _, source := range []string{"daily_checkin", "reward", "payout", "invalid", ""} {
		resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{
			UserID: "user-1",
			Source: source,
			Amount: 100,
		})

		assert.Nil(t, resp, "source=%s should fail", source)
		assert.ErrorContains(t, err, "invalid source", "source=%s should fail", source)
		assert.True(t, errors.Is(err, model.ErrValidation), "source=%s should be validation error", source)
	}
}

func TestEarnGems_ZeroAmount_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())

	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{
		UserID: "user-1",
		Source: "gameplay",
		Amount: 0,
	})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "amount must be greater than 0")
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestEarnGems_NegativeAmount_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())

	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{
		UserID: "user-1",
		Source: "gameplay",
		Amount: -10,
	})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "amount must be greater than 0")
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestEarnGems_UserNotFound_ReturnsNotFoundError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getUserByID = func(_ context.Context, _ string) (*model.User, error) {
		return nil, model.ErrNotFound
	}
	svc := newTestService(repo)

	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{
		UserID: "nonexistent",
		Source: "gameplay",
		Amount: 100,
	})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "validating user")
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestEarnGems_NoActiveChallenge_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getActiveChallenge = func(_ context.Context) (*model.WeeklyChallenge, error) {
		return nil, fmt.Errorf("getting active challenge: no rows")
	}
	svc := newTestService(repo)

	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{
		UserID: "user-1",
		Source: "gameplay",
		Amount: 100,
	})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "getting active challenge")
	assert.False(t, errors.Is(err, model.ErrValidation), "should not be validation error")
	assert.False(t, errors.Is(err, model.ErrNotFound), "should not be not-found error")
}

func TestEarnGems_AllValidSources_Succeeds(t *testing.T) {
	for _, source := range []string{"gameplay", "survey", "referral", "boost"} {
		repo := defaultMockRepo()
		svc := newTestService(repo)

		resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{
			UserID: "user-1",
			Source: source,
			Amount: 100,
		})

		assert.NoError(t, err, "source=%s should succeed", source)
		assert.NotNil(t, resp, "source=%s should succeed", source)
	}
}

func TestEarnGems_GemHistoryInsertFails_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.insertGemHistory = func(_ context.Context, _, _ string, _ int, _ *string) error {
		return fmt.Errorf("db connection lost")
	}
	svc := newTestService(repo)

	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{
		UserID: "user-1",
		Source: "gameplay",
		Amount: 100,
	})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "recording gem history")
	assert.False(t, errors.Is(err, model.ErrValidation), "should not be validation error")
}

func TestEarnGems_TransactionFails_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.runInTx = func(_ context.Context, _ func(ctx context.Context) error) error {
		return fmt.Errorf("connection pool exhausted")
	}
	svc := newTestService(repo)

	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{
		UserID: "user-1",
		Source: "gameplay",
		Amount: 100,
	})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "connection pool exhausted")
}

// =====================================================================
// Checkin tests
// =====================================================================

func TestCheckin_ValidFirstCheckin_ReturnsStreak1(t *testing.T) {
	repo := defaultMockRepo()
	repo.upsertLeaderboardEntry = func(_ context.Context, _, _, _ string, amount int) (int, error) {
		return amount, nil
	}
	svc := newTestService(repo)

	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})

	assert.NoError(t, err)
	assert.Equal(t, 150, resp.GemsEarned)    // 100 * 1.5
	assert.Equal(t, 1, resp.CurrentStreak)    // first checkin = streak 1
	assert.Equal(t, 150, resp.WeeklyGems)
}

func TestCheckin_ConsecutiveDay_IncrementsStreak(t *testing.T) {
	repo := defaultMockRepo()
	yesterday := fixedNow().AddDate(0, 0, -1)
	repo.getLastCheckin = func(_ context.Context, _ string) (*model.UserDailyCheckin, error) {
		return &model.UserDailyCheckin{
			CheckinID:     "checkin-yesterday",
			CurrentStreak: 3,
			CheckedInAt:   yesterday,
		}, nil
	}
	repo.getCheckinDate = func(_ context.Context, _ string) (time.Time, error) {
		return yesterday, nil
	}
	repo.upsertLeaderboardEntry = func(_ context.Context, _, _, _ string, amount int) (int, error) {
		return amount, nil
	}
	svc := newTestService(repo)

	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})

	assert.NoError(t, err)
	assert.Equal(t, 4, resp.CurrentStreak) // 3 + 1
}

func TestCheckin_SkippedDay_ResetsStreak(t *testing.T) {
	repo := defaultMockRepo()
	twoDaysAgo := fixedNow().AddDate(0, 0, -2)
	repo.getLastCheckin = func(_ context.Context, _ string) (*model.UserDailyCheckin, error) {
		return &model.UserDailyCheckin{
			CheckinID:     "checkin-old",
			CurrentStreak: 5,
			CheckedInAt:   twoDaysAgo,
		}, nil
	}
	repo.getCheckinDate = func(_ context.Context, _ string) (time.Time, error) {
		return twoDaysAgo, nil
	}
	repo.upsertLeaderboardEntry = func(_ context.Context, _, _, _ string, amount int) (int, error) {
		return amount, nil
	}
	svc := newTestService(repo)

	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})

	assert.NoError(t, err)
	assert.Equal(t, 1, resp.CurrentStreak) // reset
}

func TestCheckin_EmptyUserID_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())

	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: ""})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "user_id is required")
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestCheckin_NoCheckinConfig_ReturnsValidationError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getTodayCheckin = func(_ context.Context, _ time.Time) (*model.DailyCheckin, error) {
		return nil, nil
	}
	svc := newTestService(repo)

	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "no check-in available today")
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestCheckin_InactiveCheckin_ReturnsValidationError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getTodayCheckin = func(_ context.Context, _ time.Time) (*model.DailyCheckin, error) {
		return &model.DailyCheckin{ID: "checkin-1", BaseGems: 100, StreakMultiplier: 1.0, IsActive: false}, nil
	}
	svc := newTestService(repo)

	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "no check-in available today")
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestCheckin_AlreadyCheckedIn_ReturnsValidationError(t *testing.T) {
	repo := defaultMockRepo()
	repo.hasCheckedInToday = func(_ context.Context, _, _ string) (bool, error) {
		return true, nil
	}
	svc := newTestService(repo)

	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "already checked in today")
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestCheckin_UserNotFound_ReturnsNotFoundError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getUserByID = func(_ context.Context, _ string) (*model.User, error) {
		return nil, model.ErrNotFound
	}
	svc := newTestService(repo)

	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "nonexistent"})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "validating user")
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestCheckin_GemsRoundedCorrectly(t *testing.T) {
	repo := defaultMockRepo()
	repo.getTodayCheckin = func(_ context.Context, _ time.Time) (*model.DailyCheckin, error) {
		return &model.DailyCheckin{ID: "checkin-1", BaseGems: 75, StreakMultiplier: 1.25, IsActive: true}, nil
	}
	repo.upsertLeaderboardEntry = func(_ context.Context, _, _, _ string, amount int) (int, error) {
		return amount, nil
	}
	svc := newTestService(repo)

	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})

	assert.NoError(t, err)
	assert.Equal(t, 94, resp.GemsEarned) // round(75 * 1.25) = round(93.75) = 94
}

func TestCheckin_TransactionFails_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.runInTx = func(_ context.Context, _ func(ctx context.Context) error) error {
		return fmt.Errorf("connection pool exhausted")
	}
	svc := newTestService(repo)

	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})

	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "connection pool exhausted")
}
