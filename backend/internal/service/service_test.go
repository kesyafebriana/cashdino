package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kesyafebriana/cashdino/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	getTop99Entries         func(ctx context.Context, challengeID string) ([]model.LeaderboardEntry, error)
	getUserEntry            func(ctx context.Context, challengeID, userID string) (*model.LeaderboardEntry, error)
	getLastCompletedChallenge func(ctx context.Context) (*model.WeeklyChallenge, error)
	getTop99Results         func(ctx context.Context, challengeID string) ([]model.WeeklyChallengeResult, error)
	getUserResult           func(ctx context.Context, challengeID, userID string) (*model.WeeklyChallengeResult, error)
	getCampaignByChallenge  func(ctx context.Context, challengeID string) (*model.RewardCampaign, error)
	getRewardTypesByIDs     func(ctx context.Context, ids []string) ([]model.RewardType, error)
	getResultRewards        func(ctx context.Context, challengeID string) (map[string][]model.RewardInfo, error)

	// Admin campaign methods
	listCampaigns            func(ctx context.Context) ([]model.AdminCampaignListItem, error)
	getCampaignByID          func(ctx context.Context, id string) (*model.RewardCampaignFull, error)
	getRewardTypesByCampaign func(ctx context.Context, campaignID string) ([]model.RewardType, error)
	createCampaign           func(ctx context.Context, campaign *model.RewardCampaignFull) (string, error)
	createRewardType         func(ctx context.Context, rt *model.RewardType) (string, error)
	updateCampaignRules      func(ctx context.Context, campaignID string, rules json.RawMessage) error
	updateCampaign           func(ctx context.Context, campaign *model.RewardCampaignFull) error
	deleteRewardTypesByCampaign func(ctx context.Context, campaignID string) error
	getDistributions         func(ctx context.Context, campaignID string) ([]model.AdminDistributionRow, error)
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
func (m *mockRepo) GetTop99Entries(ctx context.Context, challengeID string) ([]model.LeaderboardEntry, error) {
	return m.getTop99Entries(ctx, challengeID)
}
func (m *mockRepo) GetUserEntry(ctx context.Context, challengeID, userID string) (*model.LeaderboardEntry, error) {
	return m.getUserEntry(ctx, challengeID, userID)
}
func (m *mockRepo) GetLastCompletedChallenge(ctx context.Context) (*model.WeeklyChallenge, error) {
	return m.getLastCompletedChallenge(ctx)
}
func (m *mockRepo) GetTop99Results(ctx context.Context, challengeID string) ([]model.WeeklyChallengeResult, error) {
	return m.getTop99Results(ctx, challengeID)
}
func (m *mockRepo) GetUserResult(ctx context.Context, challengeID, userID string) (*model.WeeklyChallengeResult, error) {
	return m.getUserResult(ctx, challengeID, userID)
}
func (m *mockRepo) GetCampaignByChallenge(ctx context.Context, challengeID string) (*model.RewardCampaign, error) {
	return m.getCampaignByChallenge(ctx, challengeID)
}
func (m *mockRepo) GetRewardTypesByIDs(ctx context.Context, ids []string) ([]model.RewardType, error) {
	return m.getRewardTypesByIDs(ctx, ids)
}
func (m *mockRepo) GetResultRewards(ctx context.Context, challengeID string) (map[string][]model.RewardInfo, error) {
	return m.getResultRewards(ctx, challengeID)
}
func (m *mockRepo) ListCampaigns(ctx context.Context) ([]model.AdminCampaignListItem, error) {
	return m.listCampaigns(ctx)
}
func (m *mockRepo) GetCampaignByID(ctx context.Context, id string) (*model.RewardCampaignFull, error) {
	return m.getCampaignByID(ctx, id)
}
func (m *mockRepo) GetRewardTypesByCampaign(ctx context.Context, campaignID string) ([]model.RewardType, error) {
	return m.getRewardTypesByCampaign(ctx, campaignID)
}
func (m *mockRepo) CreateCampaign(ctx context.Context, campaign *model.RewardCampaignFull) (string, error) {
	return m.createCampaign(ctx, campaign)
}
func (m *mockRepo) CreateRewardType(ctx context.Context, rt *model.RewardType) (string, error) {
	return m.createRewardType(ctx, rt)
}
func (m *mockRepo) UpdateCampaignRules(ctx context.Context, campaignID string, rules json.RawMessage) error {
	return m.updateCampaignRules(ctx, campaignID, rules)
}
func (m *mockRepo) UpdateCampaign(ctx context.Context, campaign *model.RewardCampaignFull) error {
	return m.updateCampaign(ctx, campaign)
}
func (m *mockRepo) DeleteRewardTypesByCampaign(ctx context.Context, campaignID string) error {
	return m.deleteRewardTypesByCampaign(ctx, campaignID)
}
func (m *mockRepo) GetDistributions(ctx context.Context, campaignID string) ([]model.AdminDistributionRow, error) {
	return m.getDistributions(ctx, campaignID)
}

// --- Helpers ---

var testEndTime = time.Date(2026, 3, 29, 23, 59, 59, 0, time.UTC)

func defaultMockRepo() *mockRepo {
	return &mockRepo{
		getUserByID: func(_ context.Context, _ string) (*model.User, error) {
			return &model.User{ID: "user-1", Username: "james", Email: "james@example.com"}, nil
		},
		insertGemHistory: func(_ context.Context, _, _ string, _ int, _ *string) error {
			return nil
		},
		getActiveChallenge: func(_ context.Context) (*model.WeeklyChallenge, error) {
			return &model.WeeklyChallenge{ID: "challenge-1", StartTime: fixedNow().AddDate(0, 0, -5), EndTime: testEndTime, Status: "active"}, nil
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
		getTop99Entries: func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) {
			return nil, nil
		},
		getUserEntry: func(_ context.Context, _, _ string) (*model.LeaderboardEntry, error) {
			return nil, nil
		},
		getLastCompletedChallenge: func(_ context.Context) (*model.WeeklyChallenge, error) {
			return nil, nil
		},
		getTop99Results: func(_ context.Context, _ string) ([]model.WeeklyChallengeResult, error) {
			return nil, nil
		},
		getUserResult: func(_ context.Context, _, _ string) (*model.WeeklyChallengeResult, error) {
			return nil, nil
		},
		getCampaignByChallenge: func(_ context.Context, _ string) (*model.RewardCampaign, error) {
			return nil, nil
		},
		getRewardTypesByIDs: func(_ context.Context, _ []string) ([]model.RewardType, error) {
			return nil, nil
		},
		getResultRewards: func(_ context.Context, _ string) (map[string][]model.RewardInfo, error) {
			return nil, nil
		},
		listCampaigns: func(_ context.Context) ([]model.AdminCampaignListItem, error) {
			return nil, nil
		},
		getCampaignByID: func(_ context.Context, _ string) (*model.RewardCampaignFull, error) {
			return nil, nil
		},
		getRewardTypesByCampaign: func(_ context.Context, _ string) ([]model.RewardType, error) {
			return nil, nil
		},
		createCampaign: func(_ context.Context, _ *model.RewardCampaignFull) (string, error) {
			return "camp-new", nil
		},
		createRewardType: func(_ context.Context, _ *model.RewardType) (string, error) {
			return "rt-new", nil
		},
		updateCampaignRules: func(_ context.Context, _ string, _ json.RawMessage) error {
			return nil
		},
		updateCampaign: func(_ context.Context, _ *model.RewardCampaignFull) error {
			return nil
		},
		deleteRewardTypesByCampaign: func(_ context.Context, _ string) error {
			return nil
		},
		getDistributions: func(_ context.Context, _ string) ([]model.AdminDistributionRow, error) {
			return nil, nil
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

func sampleTop99(userID string, userRank int) []model.LeaderboardEntry {
	entries := make([]model.LeaderboardEntry, 5)
	gems := []int{5000, 4000, 3000, 2000, 1000}
	names := []string{"ke****m", "gi****1", "mi****a", "xa****7", "lo****8"}
	for i := range entries {
		entries[i] = model.LeaderboardEntry{
			ID: fmt.Sprintf("entry-%d", i+1), ChallengeID: "challenge-1",
			UserID: fmt.Sprintf("other-%d", i+1), WeeklyGems: gems[i], DisplayName: names[i],
		}
	}
	if userRank >= 1 && userRank <= 5 {
		entries[userRank-1].UserID = userID
		entries[userRank-1].DisplayName = "ja****s"
	}
	return entries
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
		UserID: "user-1", Source: "gameplay", Amount: 500, GameName: &gameName,
	})

	assert.NoError(t, err)
	assert.Equal(t, "user-1", resp.UserID)
	assert.Equal(t, 1500, resp.WeeklyGems)
}

func TestEarnGems_EmptyUserID_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{Source: "gameplay", Amount: 100})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "user_id is required")
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestEarnGems_InvalidSource_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	for _, source := range []string{"daily_checkin", "reward", "payout", "invalid", ""} {
		resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{UserID: "user-1", Source: source, Amount: 100})
		assert.Nil(t, resp, "source=%s should fail", source)
		assert.ErrorContains(t, err, "invalid source", "source=%s should fail", source)
		assert.True(t, errors.Is(err, model.ErrValidation), "source=%s should be validation error", source)
	}
}

func TestEarnGems_ZeroAmount_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{UserID: "user-1", Source: "gameplay", Amount: 0})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "amount must be greater than 0")
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestEarnGems_NegativeAmount_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{UserID: "user-1", Source: "gameplay", Amount: -10})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "amount must be greater than 0")
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestEarnGems_UserNotFound_ReturnsNotFoundError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getUserByID = func(_ context.Context, _ string) (*model.User, error) { return nil, model.ErrNotFound }
	svc := newTestService(repo)
	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{UserID: "nonexistent", Source: "gameplay", Amount: 100})
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
	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{UserID: "user-1", Source: "gameplay", Amount: 100})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "getting active challenge")
	assert.False(t, errors.Is(err, model.ErrValidation))
	assert.False(t, errors.Is(err, model.ErrNotFound))
}

func TestEarnGems_AllValidSources_Succeeds(t *testing.T) {
	for _, source := range []string{"gameplay", "survey", "referral", "boost"} {
		svc := newTestService(defaultMockRepo())
		resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{UserID: "user-1", Source: source, Amount: 100})
		assert.NoError(t, err, "source=%s", source)
		assert.NotNil(t, resp, "source=%s", source)
	}
}

func TestEarnGems_GemHistoryInsertFails_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.insertGemHistory = func(_ context.Context, _, _ string, _ int, _ *string) error { return fmt.Errorf("db connection lost") }
	svc := newTestService(repo)
	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{UserID: "user-1", Source: "gameplay", Amount: 100})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "recording gem history")
}

func TestEarnGems_TransactionFails_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.runInTx = func(_ context.Context, _ func(ctx context.Context) error) error { return fmt.Errorf("connection pool exhausted") }
	svc := newTestService(repo)
	resp, err := svc.EarnGems(context.Background(), model.EarnGemsRequest{UserID: "user-1", Source: "gameplay", Amount: 100})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "connection pool exhausted")
}

// =====================================================================
// Checkin tests
// =====================================================================

func TestCheckin_ValidFirstCheckin_ReturnsStreak1(t *testing.T) {
	repo := defaultMockRepo()
	repo.upsertLeaderboardEntry = func(_ context.Context, _, _, _ string, amount int) (int, error) { return amount, nil }
	svc := newTestService(repo)
	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})
	assert.NoError(t, err)
	assert.Equal(t, 150, resp.GemsEarned)
	assert.Equal(t, 1, resp.CurrentStreak)
	assert.Equal(t, 150, resp.WeeklyGems)
}

func TestCheckin_ConsecutiveDay_IncrementsStreak(t *testing.T) {
	repo := defaultMockRepo()
	yesterday := fixedNow().AddDate(0, 0, -1)
	repo.getLastCheckin = func(_ context.Context, _ string) (*model.UserDailyCheckin, error) {
		return &model.UserDailyCheckin{CheckinID: "checkin-yesterday", CurrentStreak: 3, CheckedInAt: yesterday}, nil
	}
	repo.getCheckinDate = func(_ context.Context, _ string) (time.Time, error) { return yesterday, nil }
	repo.upsertLeaderboardEntry = func(_ context.Context, _, _, _ string, amount int) (int, error) { return amount, nil }
	svc := newTestService(repo)
	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})
	assert.NoError(t, err)
	assert.Equal(t, 4, resp.CurrentStreak)
}

func TestCheckin_SkippedDay_ResetsStreak(t *testing.T) {
	repo := defaultMockRepo()
	twoDaysAgo := fixedNow().AddDate(0, 0, -2)
	repo.getLastCheckin = func(_ context.Context, _ string) (*model.UserDailyCheckin, error) {
		return &model.UserDailyCheckin{CheckinID: "checkin-old", CurrentStreak: 5, CheckedInAt: twoDaysAgo}, nil
	}
	repo.getCheckinDate = func(_ context.Context, _ string) (time.Time, error) { return twoDaysAgo, nil }
	repo.upsertLeaderboardEntry = func(_ context.Context, _, _, _ string, amount int) (int, error) { return amount, nil }
	svc := newTestService(repo)
	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.CurrentStreak)
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
	repo.getTodayCheckin = func(_ context.Context, _ time.Time) (*model.DailyCheckin, error) { return nil, nil }
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
	repo.hasCheckedInToday = func(_ context.Context, _, _ string) (bool, error) { return true, nil }
	svc := newTestService(repo)
	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "already checked in today")
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestCheckin_UserNotFound_ReturnsNotFoundError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getUserByID = func(_ context.Context, _ string) (*model.User, error) { return nil, model.ErrNotFound }
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
	repo.upsertLeaderboardEntry = func(_ context.Context, _, _, _ string, amount int) (int, error) { return amount, nil }
	svc := newTestService(repo)
	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})
	assert.NoError(t, err)
	assert.Equal(t, 94, resp.GemsEarned)
}

func TestCheckin_TransactionFails_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.runInTx = func(_ context.Context, _ func(ctx context.Context) error) error { return fmt.Errorf("connection pool exhausted") }
	svc := newTestService(repo)
	resp, err := svc.Checkin(context.Background(), model.CheckinRequest{UserID: "user-1"})
	assert.Nil(t, resp)
	assert.ErrorContains(t, err, "connection pool exhausted")
}

// =====================================================================
// GetBanner tests
// =====================================================================

func TestGetBanner_UserInTop99_ReturnsRankAndGap(t *testing.T) {
	repo := defaultMockRepo()
	entries := sampleTop99("user-1", 3) // user-1 at rank 3, 3000 gems
	repo.getTop99Entries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) { return entries, nil }
	svc := newTestService(repo)

	resp, err := svc.GetBanner(context.Background(), "user-1")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "challenge-1", resp.ChallengeID)
	assert.Equal(t, testEndTime, resp.EndTime)
	assert.Equal(t, 3000, resp.WeeklyGems)
	assert.Equal(t, "#3", resp.RankDisplay)
	assert.Equal(t, "ja****s", resp.DisplayName)
	require.NotNil(t, resp.GapToNext)
	assert.Equal(t, 1001, *resp.GapToNext) // 4000 - 3000 + 1
}

func TestGetBanner_UserRank1_GapIsNil(t *testing.T) {
	repo := defaultMockRepo()
	entries := sampleTop99("user-1", 1) // user-1 at rank 1
	repo.getTop99Entries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) { return entries, nil }
	svc := newTestService(repo)

	resp, err := svc.GetBanner(context.Background(), "user-1")

	require.NoError(t, err)
	assert.Equal(t, "#1", resp.RankDisplay)
	assert.Nil(t, resp.GapToNext)
}

func TestGetBanner_UserNotInTop99_Returns99Plus(t *testing.T) {
	repo := defaultMockRepo()
	repo.getTop99Entries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) { return sampleTop99("", 0), nil }
	repo.getUserEntry = func(_ context.Context, _, _ string) (*model.LeaderboardEntry, error) {
		return &model.LeaderboardEntry{UserID: "user-1", WeeklyGems: 50, DisplayName: "ja****s"}, nil
	}
	svc := newTestService(repo)

	resp, err := svc.GetBanner(context.Background(), "user-1")

	require.NoError(t, err)
	assert.Equal(t, "99+", resp.RankDisplay)
	assert.Equal(t, 50, resp.WeeklyGems)
	assert.Nil(t, resp.GapToNext)
	assert.Equal(t, "ja****s", resp.DisplayName)
}

func TestGetBanner_UserHasNoEntry_Returns99PlusZeroGems(t *testing.T) {
	repo := defaultMockRepo()
	repo.getTop99Entries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) { return nil, nil }
	repo.getUserEntry = func(_ context.Context, _, _ string) (*model.LeaderboardEntry, error) { return nil, nil }
	svc := newTestService(repo)

	resp, err := svc.GetBanner(context.Background(), "user-1")

	require.NoError(t, err)
	assert.Equal(t, "99+", resp.RankDisplay)
	assert.Equal(t, 0, resp.WeeklyGems)
	assert.Nil(t, resp.GapToNext)
}

func TestGetBanner_NoActiveChallenge_ReturnsNil(t *testing.T) {
	repo := defaultMockRepo()
	repo.getActiveChallenge = func(_ context.Context) (*model.WeeklyChallenge, error) { return nil, model.ErrNotFound }
	svc := newTestService(repo)

	resp, err := svc.GetBanner(context.Background(), "user-1")

	assert.NoError(t, err)
	assert.Nil(t, resp)
}

func TestGetBanner_EmptyUserID_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	resp, err := svc.GetBanner(context.Background(), "")
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, model.ErrValidation))
}

// =====================================================================
// GetCurrentLeaderboard tests
// =====================================================================

func TestGetCurrentLeaderboard_UserInTop99_ReturnsLeaderboardAndUser(t *testing.T) {
	repo := defaultMockRepo()
	entries := sampleTop99("user-1", 2) // user-1 at rank 2, 4000 gems
	repo.getTop99Entries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) { return entries, nil }
	svc := newTestService(repo)

	resp, err := svc.GetCurrentLeaderboard(context.Background(), "user-1")

	require.NoError(t, err)
	assert.Equal(t, "challenge-1", resp.Challenge.ID)
	assert.Equal(t, "active", resp.Challenge.Status)
	assert.Len(t, resp.Leaderboard, 5)
	assert.Equal(t, 1, resp.Leaderboard[0].Rank)
	assert.Equal(t, 5000, resp.Leaderboard[0].WeeklyGems)

	require.NotNil(t, resp.CurrentUser)
	require.NotNil(t, resp.CurrentUser.Rank)
	assert.Equal(t, 2, *resp.CurrentUser.Rank)
	assert.Equal(t, "2", resp.CurrentUser.RankDisplay)
	assert.Equal(t, 4000, resp.CurrentUser.WeeklyGems)
	require.NotNil(t, resp.CurrentUser.GapToNext)
	assert.Equal(t, 1001, *resp.CurrentUser.GapToNext) // 5000 - 4000 + 1
}

func TestGetCurrentLeaderboard_UserNotInTop99_Returns99Plus(t *testing.T) {
	repo := defaultMockRepo()
	repo.getTop99Entries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) { return sampleTop99("", 0), nil }
	repo.getUserEntry = func(_ context.Context, _, _ string) (*model.LeaderboardEntry, error) {
		return &model.LeaderboardEntry{UserID: "user-1", WeeklyGems: 50, DisplayName: "ja****s"}, nil
	}
	svc := newTestService(repo)

	resp, err := svc.GetCurrentLeaderboard(context.Background(), "user-1")

	require.NoError(t, err)
	require.NotNil(t, resp.CurrentUser)
	assert.Nil(t, resp.CurrentUser.Rank)
	assert.Equal(t, "99+", resp.CurrentUser.RankDisplay)
	assert.Nil(t, resp.CurrentUser.GapToNext)
}

func TestGetCurrentLeaderboard_NoCampaign_ReturnsNilCampaign(t *testing.T) {
	repo := defaultMockRepo()
	repo.getTop99Entries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) { return nil, nil }
	svc := newTestService(repo)

	resp, err := svc.GetCurrentLeaderboard(context.Background(), "user-1")

	require.NoError(t, err)
	assert.Nil(t, resp.Campaign)
}

func TestGetCurrentLeaderboard_WithCampaign_ReturnsRewardsSummary(t *testing.T) {
	repo := defaultMockRepo()
	repo.getTop99Entries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) { return nil, nil }

	rules, _ := json.Marshal([]model.RewardRule{
		{RankFrom: 1, RankTo: 1, RewardTypeIDs: []string{"rt-1", "rt-2"}},
		{RankFrom: 2, RankTo: 5, RewardTypeIDs: []string{"rt-2"}},
	})
	repo.getCampaignByChallenge = func(_ context.Context, _ string) (*model.RewardCampaign, error) {
		return &model.RewardCampaign{ID: "camp-1", BannerImage: "https://img.png", Rules: rules, Status: "active"}, nil
	}
	repo.getRewardTypesByIDs = func(_ context.Context, _ []string) ([]model.RewardType, error) {
		return []model.RewardType{
			{ID: "rt-1", Name: "10K Gems", Type: "gems", Value: 10000},
			{ID: "rt-2", Name: "Gift Card A", Type: "gift_card", Value: 10, Image: strPtr("https://gc.png")},
		}, nil
	}
	svc := newTestService(repo)

	resp, err := svc.GetCurrentLeaderboard(context.Background(), "user-1")

	require.NoError(t, err)
	require.NotNil(t, resp.Campaign)
	assert.Equal(t, "https://img.png", resp.Campaign.BannerImage)
	require.Len(t, resp.Campaign.RewardsSummary, 2)

	row1 := resp.Campaign.RewardsSummary[0]
	assert.Equal(t, 1, row1.RankFrom)
	assert.Equal(t, 1, row1.RankTo)
	require.Len(t, row1.Rewards, 2)
	assert.Equal(t, "10K Gems", row1.Rewards[0].Name)
	assert.Equal(t, float64(10000), row1.Rewards[0].Value)

	row2 := resp.Campaign.RewardsSummary[1]
	assert.Equal(t, 2, row2.RankFrom)
	assert.Equal(t, 5, row2.RankTo)
	require.Len(t, row2.Rewards, 1)
	assert.Equal(t, "Gift Card A", row2.Rewards[0].Name)
}

func TestGetCurrentLeaderboard_NoActiveChallenge_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getActiveChallenge = func(_ context.Context) (*model.WeeklyChallenge, error) { return nil, model.ErrNotFound }
	svc := newTestService(repo)

	resp, err := svc.GetCurrentLeaderboard(context.Background(), "user-1")
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestGetCurrentLeaderboard_EmptyUserID_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	resp, err := svc.GetCurrentLeaderboard(context.Background(), "")
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, model.ErrValidation))
}

// =====================================================================
// GetLastWeekLeaderboard tests
// =====================================================================

func TestGetLastWeek_NoCompletedChallenge_ReturnsNilChallenge(t *testing.T) {
	repo := defaultMockRepo()
	repo.getLastCompletedChallenge = func(_ context.Context) (*model.WeeklyChallenge, error) { return nil, nil }
	svc := newTestService(repo)

	resp, err := svc.GetLastWeekLeaderboard(context.Background(), "user-1")

	require.NoError(t, err)
	assert.Nil(t, resp.Challenge)
	assert.Nil(t, resp.Leaderboard)
	assert.Nil(t, resp.CurrentUser)
}

func TestGetLastWeek_UserInResults_ReturnsRankAndRewards(t *testing.T) {
	repo := defaultMockRepo()
	repo.getLastCompletedChallenge = func(_ context.Context) (*model.WeeklyChallenge, error) {
		return &model.WeeklyChallenge{ID: "challenge-0", Status: "completed", EndTime: fixedNow().AddDate(0, 0, -1)}, nil
	}
	repo.getTop99Results = func(_ context.Context, _ string) ([]model.WeeklyChallengeResult, error) {
		return []model.WeeklyChallengeResult{
			{ID: "r1", ChallengeID: "challenge-0", UserID: "user-2", FinalRank: 1, FinalGems: 9000, DisplayName: "ke****m"},
			{ID: "r2", ChallengeID: "challenge-0", UserID: "user-1", FinalRank: 2, FinalGems: 7000, DisplayName: "ja****s"},
		}, nil
	}
	repo.getResultRewards = func(_ context.Context, _ string) (map[string][]model.RewardInfo, error) {
		return map[string][]model.RewardInfo{
			"user-2": {{Name: "10K Gems", Type: "gems", Value: 10000}},
			"user-1": {{Name: "Gift Card", Type: "gift_card", Value: 10, Image: strPtr("https://gc.png")}},
		}, nil
	}
	svc := newTestService(repo)

	resp, err := svc.GetLastWeekLeaderboard(context.Background(), "user-1")

	require.NoError(t, err)
	require.NotNil(t, resp.Challenge)
	assert.Equal(t, "challenge-0", resp.Challenge.ID)

	require.Len(t, resp.Leaderboard, 2)
	assert.Equal(t, 1, resp.Leaderboard[0].Rank)
	assert.Len(t, resp.Leaderboard[0].Rewards, 1)
	assert.Equal(t, "10K Gems", resp.Leaderboard[0].Rewards[0].Name)

	require.NotNil(t, resp.CurrentUser)
	require.NotNil(t, resp.CurrentUser.Rank)
	assert.Equal(t, 2, *resp.CurrentUser.Rank)
	assert.Equal(t, "2", resp.CurrentUser.RankDisplay)
	assert.Equal(t, 7000, resp.CurrentUser.FinalGems)
	assert.Len(t, resp.CurrentUser.Rewards, 1)
}

func TestGetLastWeek_UserNotInResults_Returns99Plus(t *testing.T) {
	repo := defaultMockRepo()
	repo.getLastCompletedChallenge = func(_ context.Context) (*model.WeeklyChallenge, error) {
		return &model.WeeklyChallenge{ID: "challenge-0", Status: "completed"}, nil
	}
	repo.getTop99Results = func(_ context.Context, _ string) ([]model.WeeklyChallengeResult, error) {
		return []model.WeeklyChallengeResult{
			{UserID: "user-2", FinalRank: 1, FinalGems: 9000, DisplayName: "ke****m"},
		}, nil
	}
	repo.getUserResult = func(_ context.Context, _, _ string) (*model.WeeklyChallengeResult, error) {
		return &model.WeeklyChallengeResult{UserID: "user-1", FinalRank: 150, FinalGems: 100, DisplayName: "ja****s"}, nil
	}
	repo.getResultRewards = func(_ context.Context, _ string) (map[string][]model.RewardInfo, error) { return nil, nil }
	svc := newTestService(repo)

	resp, err := svc.GetLastWeekLeaderboard(context.Background(), "user-1")

	require.NoError(t, err)
	require.NotNil(t, resp.CurrentUser)
	assert.Nil(t, resp.CurrentUser.Rank)
	assert.Equal(t, "99+", resp.CurrentUser.RankDisplay)
	assert.Equal(t, 100, resp.CurrentUser.FinalGems)
}

func TestGetLastWeek_EmptyUserID_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	resp, err := svc.GetLastWeekLeaderboard(context.Background(), "")
	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func strPtr(s string) *string { return &s }

// =====================================================================
// ListCampaigns tests
// =====================================================================

func TestListCampaigns_ReturnsCampaigns(t *testing.T) {
	repo := defaultMockRepo()
	repo.listCampaigns = func(_ context.Context) ([]model.AdminCampaignListItem, error) {
		return []model.AdminCampaignListItem{
			{
				ID: "camp-1", ChallengeID: "ch-1", Name: "Week 14", Status: "active",
				ChallengeStart: fixedNow().AddDate(0, 0, -5), ChallengeEnd: testEndTime,
				RewardTypesCount: 3, TotalStock: 11,
			},
			{
				ID: "camp-2", ChallengeID: "ch-0", Name: "Week 13", Status: "completed",
				ChallengeStart: fixedNow().AddDate(0, 0, -12), ChallengeEnd: fixedNow().AddDate(0, 0, -6),
				RewardTypesCount: 2, TotalStock: 6,
			},
		}, nil
	}
	svc := newTestService(repo)

	result, err := svc.ListCampaigns(context.Background())

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "camp-1", result[0].ID)
	assert.Equal(t, 3, result[0].RewardTypesCount)
	assert.Equal(t, 11, result[0].TotalStock)
	assert.Equal(t, "camp-2", result[1].ID)
}

func TestListCampaigns_Empty_ReturnsEmptySlice(t *testing.T) {
	repo := defaultMockRepo()
	repo.listCampaigns = func(_ context.Context) ([]model.AdminCampaignListItem, error) {
		return []model.AdminCampaignListItem{}, nil
	}
	svc := newTestService(repo)

	result, err := svc.ListCampaigns(context.Background())

	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestListCampaigns_RepoError_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.listCampaigns = func(_ context.Context) ([]model.AdminCampaignListItem, error) {
		return nil, fmt.Errorf("db error")
	}
	svc := newTestService(repo)

	result, err := svc.ListCampaigns(context.Background())

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "listing campaigns")
}

// =====================================================================
// GetCampaign tests
// =====================================================================

func TestGetCampaign_ReturnsCampaignWithRewardTypesAndRules(t *testing.T) {
	repo := defaultMockRepo()
	rules, _ := json.Marshal([]model.RewardRule{
		{RankFrom: 1, RankTo: 1, RewardTypeIDs: []string{"rt-1", "rt-2"}},
		{RankFrom: 2, RankTo: 5, RewardTypeIDs: []string{"rt-2"}},
	})
	repo.getCampaignByID = func(_ context.Context, _ string) (*model.RewardCampaignFull, error) {
		return &model.RewardCampaignFull{
			ID: "camp-1", ChallengeID: "ch-1", Name: "Week 14", BannerImage: "https://img.png",
			Rules: rules, Status: "active",
			NonGemClaimEmailSubject: "You won!", NonGemClaimEmailBody: "Hi {{username}}",
		}, nil
	}
	repo.getRewardTypesByCampaign = func(_ context.Context, _ string) ([]model.RewardType, error) {
		return []model.RewardType{
			{ID: "rt-1", CampaignID: "camp-1", Name: "10K Gems", Type: "gems", Value: 10000, Stock: 1},
			{ID: "rt-2", CampaignID: "camp-1", Name: "Gift Card A", Type: "gift_card", Value: 10, Image: strPtr("https://gc.png"), Stock: 5},
		}, nil
	}
	svc := newTestService(repo)

	result, err := svc.GetCampaign(context.Background(), "camp-1")

	require.NoError(t, err)
	assert.Equal(t, "camp-1", result.ID)
	assert.Equal(t, "You won!", result.NonGemClaimEmailSubject)
	assert.Equal(t, "Hi {{username}}", result.NonGemClaimEmailBody)
	require.Len(t, result.RewardTypes, 2)
	assert.Equal(t, "10K Gems", result.RewardTypes[0].Name)
	require.Len(t, result.Rules, 2)
	assert.Equal(t, 1, result.Rules[0].RankFrom)
	assert.Equal(t, 1, result.Rules[0].RankTo)
	require.Len(t, result.Rules[0].RewardNames, 2)
	assert.Equal(t, "10K Gems", result.Rules[0].RewardNames[0])
	assert.Equal(t, "Gift Card A", result.Rules[0].RewardNames[1])
}

func TestGetCampaign_EmptyID_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	result, err := svc.GetCampaign(context.Background(), "")
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestGetCampaign_NotFound_ReturnsNotFoundError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getCampaignByID = func(_ context.Context, _ string) (*model.RewardCampaignFull, error) {
		return nil, nil
	}
	svc := newTestService(repo)

	result, err := svc.GetCampaign(context.Background(), "nonexistent")

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

// =====================================================================
// CreateCampaign tests
// =====================================================================

func validCreateCampaignReq() model.CreateCampaignRequest {
	return model.CreateCampaignRequest{
		ChallengeID: "ch-1",
		Name:        "Week 14 — Easter",
		BannerImage: "https://img.png",
		RewardTypes: []model.CreateRewardTypeInput{
			{Name: "10K Gems", Type: "gems", Value: 10000, Stock: 1},
			{Name: "Gift Card A", Type: "gift_card", Value: 10, Image: strPtr("https://gc.png"), Stock: 5},
		},
		Rules: []model.CreateCampaignRuleInput{
			{RankFrom: 1, RankTo: 1, RewardTypeIndexes: []int{0, 1}},
			{RankFrom: 2, RankTo: 5, RewardTypeIndexes: []int{1}},
		},
		NonGemClaimEmailSubject: "You won!",
		NonGemClaimEmailBody:    "Hi {{username}}, you placed #{{rank}}...",
	}
}

func TestCreateCampaign_ValidRequest_ReturnsCampaign(t *testing.T) {
	repo := defaultMockRepo()
	rtIndex := 0
	repo.createRewardType = func(_ context.Context, _ *model.RewardType) (string, error) {
		rtIndex++
		return fmt.Sprintf("rt-%d", rtIndex), nil
	}
	repo.createCampaign = func(_ context.Context, _ *model.RewardCampaignFull) (string, error) {
		return "camp-new", nil
	}
	repo.getCampaignByID = func(_ context.Context, _ string) (*model.RewardCampaignFull, error) {
		rules, _ := json.Marshal([]model.RewardRule{
			{RankFrom: 1, RankTo: 1, RewardTypeIDs: []string{"rt-1", "rt-2"}},
			{RankFrom: 2, RankTo: 5, RewardTypeIDs: []string{"rt-2"}},
		})
		return &model.RewardCampaignFull{
			ID: "camp-new", ChallengeID: "ch-1", Name: "Week 14 — Easter", BannerImage: "https://img.png",
			Rules: rules, Status: "draft",
			NonGemClaimEmailSubject: "You won!", NonGemClaimEmailBody: "Hi {{username}}, you placed #{{rank}}...",
		}, nil
	}
	repo.getRewardTypesByCampaign = func(_ context.Context, _ string) ([]model.RewardType, error) {
		return []model.RewardType{
			{ID: "rt-1", CampaignID: "camp-new", Name: "10K Gems", Type: "gems", Value: 10000, Stock: 1},
			{ID: "rt-2", CampaignID: "camp-new", Name: "Gift Card A", Type: "gift_card", Value: 10, Image: strPtr("https://gc.png"), Stock: 5},
		}, nil
	}
	svc := newTestService(repo)

	result, err := svc.CreateCampaign(context.Background(), validCreateCampaignReq())

	require.NoError(t, err)
	assert.Equal(t, "camp-new", result.ID)
	assert.Equal(t, "Week 14 — Easter", result.Name)
}

func TestCreateCampaign_RankFromGreaterThanRankTo_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	req := validCreateCampaignReq()
	req.Rules = []model.CreateCampaignRuleInput{
		{RankFrom: 5, RankTo: 1, RewardTypeIndexes: []int{0}},
	}

	result, err := svc.CreateCampaign(context.Background(), req)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
	assert.ErrorContains(t, err, "rank_from must be <= rank_to")
}

func TestCreateCampaign_RankFromLessThan1_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	req := validCreateCampaignReq()
	req.Rules = []model.CreateCampaignRuleInput{
		{RankFrom: 0, RankTo: 5, RewardTypeIndexes: []int{0}},
	}

	result, err := svc.CreateCampaign(context.Background(), req)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
	assert.ErrorContains(t, err, "rank_from must be >= 1")
}

func TestCreateCampaign_OverlappingRanges_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	req := validCreateCampaignReq()
	req.Rules = []model.CreateCampaignRuleInput{
		{RankFrom: 1, RankTo: 5, RewardTypeIndexes: []int{0}},
		{RankFrom: 3, RankTo: 10, RewardTypeIndexes: []int{1}},
	}

	result, err := svc.CreateCampaign(context.Background(), req)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
	assert.ErrorContains(t, err, "overlapping rank ranges")
}

func TestCreateCampaign_StockExceeded_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	req := validCreateCampaignReq()
	// Gift Card A (stock 5) used in rank 1-1 (1 user) + rank 2-10 (9 users) = 10 total > 5
	req.Rules = []model.CreateCampaignRuleInput{
		{RankFrom: 1, RankTo: 1, RewardTypeIndexes: []int{0, 1}},
		{RankFrom: 2, RankTo: 10, RewardTypeIndexes: []int{1}},
	}

	result, err := svc.CreateCampaign(context.Background(), req)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
	assert.ErrorContains(t, err, "Gift Card A: need 10, stock is 5")
}

func TestCreateCampaign_InvalidRewardTypeIndex_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	req := validCreateCampaignReq()
	req.Rules = []model.CreateCampaignRuleInput{
		{RankFrom: 1, RankTo: 1, RewardTypeIndexes: []int{0, 5}},
	}

	result, err := svc.CreateCampaign(context.Background(), req)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
	assert.ErrorContains(t, err, "invalid reward_type_index")
}

func TestCreateCampaign_EmptyChallengeID_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	req := validCreateCampaignReq()
	req.ChallengeID = ""

	result, err := svc.CreateCampaign(context.Background(), req)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
	assert.ErrorContains(t, err, "challenge_id is required")
}

func TestCreateCampaign_EmptyName_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	req := validCreateCampaignReq()
	req.Name = ""

	result, err := svc.CreateCampaign(context.Background(), req)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
	assert.ErrorContains(t, err, "name is required")
}

func TestCreateCampaign_NoRewardTypes_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	req := validCreateCampaignReq()
	req.RewardTypes = nil

	result, err := svc.CreateCampaign(context.Background(), req)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
	assert.ErrorContains(t, err, "at least one reward_type is required")
}

func TestCreateCampaign_NoRules_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())
	req := validCreateCampaignReq()
	req.Rules = nil

	result, err := svc.CreateCampaign(context.Background(), req)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
	assert.ErrorContains(t, err, "at least one rule is required")
}

func TestCreateCampaign_TransactionFails_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.runInTx = func(_ context.Context, _ func(ctx context.Context) error) error {
		return fmt.Errorf("connection pool exhausted")
	}
	svc := newTestService(repo)

	result, err := svc.CreateCampaign(context.Background(), validCreateCampaignReq())

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "connection pool exhausted")
}

// =====================================================================
// UpdateCampaign tests
// =====================================================================

func TestUpdateCampaign_ValidRequest_ReturnsCampaign(t *testing.T) {
	repo := defaultMockRepo()
	repo.getCampaignByID = func(_ context.Context, id string) (*model.RewardCampaignFull, error) {
		if id == "camp-1" {
			rules, _ := json.Marshal([]model.RewardRule{
				{RankFrom: 1, RankTo: 1, RewardTypeIDs: []string{"rt-1", "rt-2"}},
				{RankFrom: 2, RankTo: 5, RewardTypeIDs: []string{"rt-2"}},
			})
			return &model.RewardCampaignFull{
				ID: "camp-1", ChallengeID: "ch-1", Name: "Week 14 — Easter", BannerImage: "https://img.png",
				Rules: rules, Status: "draft",
				NonGemClaimEmailSubject: "You won!", NonGemClaimEmailBody: "Hi {{username}}",
			}, nil
		}
		return nil, nil
	}
	rtIndex := 0
	repo.createRewardType = func(_ context.Context, _ *model.RewardType) (string, error) {
		rtIndex++
		return fmt.Sprintf("rt-new-%d", rtIndex), nil
	}
	repo.getRewardTypesByCampaign = func(_ context.Context, _ string) ([]model.RewardType, error) {
		return []model.RewardType{
			{ID: "rt-new-1", CampaignID: "camp-1", Name: "10K Gems", Type: "gems", Value: 10000, Stock: 1},
			{ID: "rt-new-2", CampaignID: "camp-1", Name: "Gift Card A", Type: "gift_card", Value: 10, Image: strPtr("https://gc.png"), Stock: 5},
		}, nil
	}
	svc := newTestService(repo)

	result, err := svc.UpdateCampaign(context.Background(), "camp-1", validCreateCampaignReq())

	require.NoError(t, err)
	assert.Equal(t, "camp-1", result.ID)
}

func TestUpdateCampaign_NotFound_ReturnsNotFoundError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getCampaignByID = func(_ context.Context, _ string) (*model.RewardCampaignFull, error) {
		return nil, nil
	}
	svc := newTestService(repo)

	result, err := svc.UpdateCampaign(context.Background(), "nonexistent", validCreateCampaignReq())

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestUpdateCampaign_EmptyID_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())

	result, err := svc.UpdateCampaign(context.Background(), "", validCreateCampaignReq())

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestUpdateCampaign_OverlappingRanges_ReturnsValidationError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getCampaignByID = func(_ context.Context, _ string) (*model.RewardCampaignFull, error) {
		return &model.RewardCampaignFull{ID: "camp-1", Status: "draft"}, nil
	}
	svc := newTestService(repo)
	req := validCreateCampaignReq()
	req.Rules = []model.CreateCampaignRuleInput{
		{RankFrom: 1, RankTo: 5, RewardTypeIndexes: []int{0}},
		{RankFrom: 3, RankTo: 10, RewardTypeIndexes: []int{1}},
	}

	result, err := svc.UpdateCampaign(context.Background(), "camp-1", req)

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
	assert.ErrorContains(t, err, "overlapping rank ranges")
}

func TestUpdateCampaign_HasDistributions_ReturnsValidationError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getCampaignByID = func(_ context.Context, _ string) (*model.RewardCampaignFull, error) {
		return &model.RewardCampaignFull{ID: "camp-1", Status: "completed"}, nil
	}
	repo.getDistributions = func(_ context.Context, _ string) ([]model.AdminDistributionRow, error) {
		return []model.AdminDistributionRow{
			{ID: "dist-1", UserID: "user-1", Status: "delivered", FinalRank: 1},
		}, nil
	}
	svc := newTestService(repo)

	result, err := svc.UpdateCampaign(context.Background(), "camp-1", validCreateCampaignReq())

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
	assert.ErrorContains(t, err, "cannot update campaign that has reward distributions")
}

// =====================================================================
// GetDistributions tests
// =====================================================================

func TestGetDistributions_ReturnsDistributions(t *testing.T) {
	repo := defaultMockRepo()
	delivered := fixedNow()
	repo.getDistributions = func(_ context.Context, _ string) ([]model.AdminDistributionRow, error) {
		return []model.AdminDistributionRow{
			{
				ID: "dist-1", UserID: "user-1", DisplayName: "ja****s", MaskedEmail: "ja****@example.com",
				RewardName: "10K Gems", RewardType: "gems", RewardValue: 10000,
				Status: "delivered", DeliveredAt: &delivered, FinalRank: 1,
			},
			{
				ID: "dist-2", UserID: "user-2", DisplayName: "ke****m", MaskedEmail: "ke****@example.com",
				RewardName: "Gift Card A", RewardType: "gift_card", RewardValue: 10, RewardImage: strPtr("https://gc.png"),
				Status: "delivered", DeliveredAt: &delivered, FinalRank: 2,
			},
		}, nil
	}
	svc := newTestService(repo)

	result, err := svc.GetDistributions(context.Background(), "camp-1")

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "dist-1", result[0].ID)
	assert.Equal(t, 1, result[0].FinalRank)
	assert.Equal(t, "dist-2", result[1].ID)
	assert.Equal(t, 2, result[1].FinalRank)
}

func TestGetDistributions_EmptyID_ReturnsValidationError(t *testing.T) {
	svc := newTestService(defaultMockRepo())

	result, err := svc.GetDistributions(context.Background(), "")

	assert.Nil(t, result)
	assert.True(t, errors.Is(err, model.ErrValidation))
}

func TestGetDistributions_RepoError_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.getDistributions = func(_ context.Context, _ string) ([]model.AdminDistributionRow, error) {
		return nil, fmt.Errorf("db error")
	}
	svc := newTestService(repo)

	result, err := svc.GetDistributions(context.Background(), "camp-1")

	assert.Nil(t, result)
	assert.ErrorContains(t, err, "getting distributions")
}
