package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kesyafebriana/cashdino/backend/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================================
// WeeklyReset tests
// =====================================================================

func TestWeeklyReset_NoActiveChallenge_ReturnsError(t *testing.T) {
	repo := defaultMockRepo()
	repo.completeActiveChallenge = func(_ context.Context) (string, error) {
		return "", fmt.Errorf("completing active challenge: %w", model.ErrNotFound)
	}
	svc := newTestService(repo)

	_, err := svc.WeeklyReset(context.Background())
	assert.Error(t, err)
	assert.True(t, errors.Is(err, model.ErrNotFound))
}

func TestWeeklyReset_SnapshotsResults(t *testing.T) {
	repo := defaultMockRepo()
	now := time.Date(2026, 3, 29, 23, 59, 0, 0, time.UTC)

	repo.getAllLeaderboardEntries = func(_ context.Context, challengeID string) ([]model.LeaderboardEntry, error) {
		assert.Equal(t, "challenge-1", challengeID)
		return []model.LeaderboardEntry{
			{ChallengeID: challengeID, UserID: "user-1", WeeklyGems: 5000, DisplayName: "ja****s"},
			{ChallengeID: challengeID, UserID: "user-2", WeeklyGems: 3000, DisplayName: "bo****b"},
		}, nil
	}

	var insertedResults []*model.WeeklyChallengeResult
	repo.insertChallengeResult = func(_ context.Context, result *model.WeeklyChallengeResult) error {
		insertedResults = append(insertedResults, result)
		return nil
	}

	repo.getCampaignByChallenge = func(_ context.Context, _ string) (*model.RewardCampaign, error) {
		return nil, nil
	}

	svc := newTestService(repo)
	svc.now = func() time.Time { return now }

	resp, err := svc.WeeklyReset(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, 2, resp.ResultsArchived)
	assert.Equal(t, 0, resp.RewardsDistributed)

	require.Len(t, insertedResults, 2)
	assert.Equal(t, 1, insertedResults[0].FinalRank)
	assert.Equal(t, "user-1", insertedResults[0].UserID)
	assert.Equal(t, 5000, insertedResults[0].FinalGems)
	assert.Equal(t, 2, insertedResults[1].FinalRank)
	assert.Equal(t, "user-2", insertedResults[1].UserID)
	assert.Equal(t, 3000, insertedResults[1].FinalGems)
}

func TestWeeklyReset_CreatesNextChallenge(t *testing.T) {
	repo := defaultMockRepo()
	now := time.Date(2026, 3, 29, 23, 59, 0, 0, time.UTC)

	repo.getAllLeaderboardEntries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) {
		return nil, nil
	}
	repo.getCampaignByChallenge = func(_ context.Context, _ string) (*model.RewardCampaign, error) {
		return nil, nil
	}

	var capturedStart, capturedEnd time.Time
	repo.insertWeeklyChallenge = func(_ context.Context, start, end time.Time, status string) (string, error) {
		capturedStart = start
		capturedEnd = end
		assert.Equal(t, "active", status)
		return "new-challenge-1", nil
	}

	svc := newTestService(repo)
	svc.now = func() time.Time { return now }

	resp, err := svc.WeeklyReset(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "new-challenge-1", resp.NewChallengeID)

	expectedStart := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 4, 5, 23, 59, 59, 0, time.UTC)
	assert.Equal(t, expectedStart, capturedStart)
	assert.Equal(t, expectedEnd, capturedEnd)
}

func TestWeeklyReset_DistributesGemRewards(t *testing.T) {
	repo := defaultMockRepo()
	now := time.Date(2026, 3, 29, 23, 59, 0, 0, time.UTC)

	repo.getAllLeaderboardEntries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) {
		return []model.LeaderboardEntry{
			{ChallengeID: "challenge-1", UserID: "user-1", WeeklyGems: 5000, DisplayName: "ja****s"},
		}, nil
	}

	rules := []model.RewardRule{
		{RankFrom: 1, RankTo: 1, RewardTypeIDs: []string{"rt-gems"}},
	}
	rulesJSON, _ := json.Marshal(rules)

	repo.getCampaignByChallenge = func(_ context.Context, _ string) (*model.RewardCampaign, error) {
		return &model.RewardCampaign{
			ID: "camp-1", ChallengeID: "challenge-1", Name: "Week 14",
			Rules: rulesJSON, Status: "active",
		}, nil
	}

	repo.getCampaignByID = func(_ context.Context, _ string) (*model.RewardCampaignFull, error) {
		return &model.RewardCampaignFull{ID: "camp-1", ChallengeID: "challenge-1"}, nil
	}

	repo.getRewardTypeByID = func(_ context.Context, _ string) (*model.RewardType, error) {
		return &model.RewardType{ID: "rt-gems", Type: "gems", Value: 10000, Name: "10K Gems", Stock: 1}, nil
	}

	var gemInserts []struct{ userID, source string; amount int }
	repo.insertGemHistory = func(_ context.Context, userID, source string, amount int, _ *string) error {
		gemInserts = append(gemInserts, struct{ userID, source string; amount int }{userID, source, amount})
		return nil
	}

	var distributions []*model.RewardDistribution
	repo.insertRewardDistribution = func(_ context.Context, dist *model.RewardDistribution) (string, error) {
		distributions = append(distributions, dist)
		return "dist-1", nil
	}

	var stockDecrements []string
	repo.decrementRewardTypeStock = func(_ context.Context, id string) error {
		stockDecrements = append(stockDecrements, id)
		return nil
	}

	var campaignStatusUpdates []string
	repo.updateCampaignStatus = func(_ context.Context, _, status string) error {
		campaignStatusUpdates = append(campaignStatusUpdates, status)
		return nil
	}

	repo.getUserByID = func(_ context.Context, userID string) (*model.User, error) {
		return &model.User{ID: userID, Username: "james", Email: "james@example.com"}, nil
	}

	svc := newTestService(repo)
	svc.now = func() time.Time { return now }

	resp, err := svc.WeeklyReset(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 1, resp.RewardsDistributed)

	// Gem reward should insert gem_history
	require.Len(t, gemInserts, 1)
	assert.Equal(t, "user-1", gemInserts[0].userID)
	assert.Equal(t, "reward", gemInserts[0].source)
	assert.Equal(t, 10000, gemInserts[0].amount)

	// Should insert a delivered distribution (gems are immediate)
	require.Len(t, distributions, 1)
	assert.Equal(t, "delivered", distributions[0].Status)
	assert.NotNil(t, distributions[0].DeliveredAt)

	require.Len(t, stockDecrements, 1)
	assert.Equal(t, "rt-gems", stockDecrements[0])

	require.Len(t, campaignStatusUpdates, 1)
	assert.Equal(t, "completed", campaignStatusUpdates[0])
}

func TestWeeklyReset_DistributesNonGemRewards_SendsEmailAfterCommit(t *testing.T) {
	repo := defaultMockRepo()
	now := time.Date(2026, 3, 29, 23, 59, 0, 0, time.UTC)

	repo.getAllLeaderboardEntries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) {
		return []model.LeaderboardEntry{
			{ChallengeID: "challenge-1", UserID: "user-1", WeeklyGems: 5000, DisplayName: "ja****s"},
		}, nil
	}

	rules := []model.RewardRule{
		{RankFrom: 1, RankTo: 1, RewardTypeIDs: []string{"rt-gc"}},
	}
	rulesJSON, _ := json.Marshal(rules)

	repo.getCampaignByChallenge = func(_ context.Context, _ string) (*model.RewardCampaign, error) {
		return &model.RewardCampaign{
			ID: "camp-1", ChallengeID: "challenge-1",
			Rules: rulesJSON, Status: "active",
		}, nil
	}

	imgURL := "https://img.png/gc.jpg"
	repo.getRewardTypeByID = func(_ context.Context, _ string) (*model.RewardType, error) {
		return &model.RewardType{ID: "rt-gc", Type: "gift_card", Value: 50, Name: "$50 Gift Card", Image: &imgURL, Stock: 1}, nil
	}

	repo.getCampaignByID = func(_ context.Context, _ string) (*model.RewardCampaignFull, error) {
		return &model.RewardCampaignFull{
			ID: "camp-1", ChallengeID: "challenge-1",
			NonGemClaimEmailSubject: "Congrats {{username}}!",
			NonGemClaimEmailBody:    "Rank #{{rank}} wins {{reward_type}} ({{reward_value}}) {{reward_image}}",
		}, nil
	}

	repo.getUserByID = func(_ context.Context, userID string) (*model.User, error) {
		return &model.User{ID: userID, Username: "james", Email: "james@example.com"}, nil
	}

	// Non-gem distributions are inserted as 'pending' during tx
	var distributions []*model.RewardDistribution
	repo.insertRewardDistribution = func(_ context.Context, dist *model.RewardDistribution) (string, error) {
		distributions = append(distributions, dist)
		return "dist-1", nil
	}

	// After tx commit, email succeeds → UpdateDistributionDelivered is called
	var deliveredIDs []string
	repo.updateDistributionDelivered = func(_ context.Context, id string) error {
		deliveredIDs = append(deliveredIDs, id)
		return nil
	}

	repo.decrementRewardTypeStock = func(_ context.Context, _ string) error { return nil }
	repo.updateCampaignStatus = func(_ context.Context, _, _ string) error { return nil }

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	svc := newTestService(repo)
	svc.now = func() time.Time { return now }

	resp, err := svc.WeeklyReset(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, resp.RewardsDistributed)

	// Distribution inserted as pending during tx
	require.Len(t, distributions, 1)
	assert.Equal(t, "pending", distributions[0].Status)

	// Email sent after commit → distribution updated to delivered
	require.Len(t, deliveredIDs, 1)
	assert.Equal(t, "dist-1", deliveredIDs[0])

	// Email was logged (console mode)
	output := buf.String()
	assert.Contains(t, output, "james@example.com")
	assert.Contains(t, output, "Congrats james!")
	assert.Contains(t, output, "Rank #1")
	assert.Contains(t, output, "$50 Gift Card")
}

func TestWeeklyReset_NonGemReward_EmailFails_StatusFailed(t *testing.T) {
	repo := defaultMockRepo()
	now := time.Date(2026, 3, 29, 23, 59, 0, 0, time.UTC)

	repo.getAllLeaderboardEntries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) {
		return []model.LeaderboardEntry{
			{ChallengeID: "challenge-1", UserID: "user-1", WeeklyGems: 5000, DisplayName: "ja****s"},
		}, nil
	}

	rules := []model.RewardRule{
		{RankFrom: 1, RankTo: 1, RewardTypeIDs: []string{"rt-gc"}},
	}
	rulesJSON, _ := json.Marshal(rules)

	repo.getCampaignByChallenge = func(_ context.Context, _ string) (*model.RewardCampaign, error) {
		return &model.RewardCampaign{
			ID: "camp-1", ChallengeID: "challenge-1",
			Rules: rulesJSON, Status: "active",
		}, nil
	}

	repo.getRewardTypeByID = func(_ context.Context, _ string) (*model.RewardType, error) {
		return &model.RewardType{ID: "rt-gc", Type: "gift_card", Value: 50, Name: "$50 Gift Card", Stock: 1}, nil
	}

	repo.getCampaignByID = func(_ context.Context, _ string) (*model.RewardCampaignFull, error) {
		return &model.RewardCampaignFull{
			ID: "camp-1", ChallengeID: "challenge-1",
			NonGemClaimEmailSubject: "Subject", NonGemClaimEmailBody: "Body",
		}, nil
	}

	repo.getUserByID = func(_ context.Context, _ string) (*model.User, error) {
		return &model.User{ID: "user-1", Username: "james", Email: "james@example.com"}, nil
	}

	repo.insertRewardDistribution = func(_ context.Context, _ *model.RewardDistribution) (string, error) {
		return "dist-1", nil
	}

	// After tx commit, email fails → UpdateDistributionFailed is called
	var failedIDs []string
	repo.updateDistributionFailed = func(_ context.Context, id string) error {
		failedIDs = append(failedIDs, id)
		return nil
	}

	repo.decrementRewardTypeStock = func(_ context.Context, _ string) error { return nil }
	repo.updateCampaignStatus = func(_ context.Context, _, _ string) error { return nil }

	// Use SMTP that will fail
	failEmail := NewEmailService("localhost", 19999, "test@test.com", "pass")
	svc := newTestServiceWithEmail(repo, failEmail)
	svc.now = func() time.Time { return now }

	resp, err := svc.WeeklyReset(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, resp.RewardsDistributed)

	// Distribution marked failed after email failure
	require.Len(t, failedIDs, 1)
	assert.Equal(t, "dist-1", failedIDs[0])
}

func TestWeeklyReset_NoCampaign_SkipsDistribution(t *testing.T) {
	repo := defaultMockRepo()
	now := time.Date(2026, 3, 29, 23, 59, 0, 0, time.UTC)

	repo.getAllLeaderboardEntries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) {
		return []model.LeaderboardEntry{
			{ChallengeID: "challenge-1", UserID: "user-1", WeeklyGems: 5000, DisplayName: "ja****s"},
		}, nil
	}
	repo.getCampaignByChallenge = func(_ context.Context, _ string) (*model.RewardCampaign, error) {
		return nil, nil
	}

	svc := newTestService(repo)
	svc.now = func() time.Time { return now }

	resp, err := svc.WeeklyReset(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, resp.ResultsArchived)
	assert.Equal(t, 0, resp.RewardsDistributed)
}

func TestWeeklyReset_MultipleRulesMultipleRewards(t *testing.T) {
	repo := defaultMockRepo()
	now := time.Date(2026, 3, 29, 23, 59, 0, 0, time.UTC)

	repo.getAllLeaderboardEntries = func(_ context.Context, _ string) ([]model.LeaderboardEntry, error) {
		return []model.LeaderboardEntry{
			{ChallengeID: "challenge-1", UserID: "user-1", WeeklyGems: 5000, DisplayName: "ja****s"},
			{ChallengeID: "challenge-1", UserID: "user-2", WeeklyGems: 3000, DisplayName: "bo****b"},
			{ChallengeID: "challenge-1", UserID: "user-3", WeeklyGems: 1000, DisplayName: "ch****c"},
		}, nil
	}

	rules := []model.RewardRule{
		{RankFrom: 1, RankTo: 1, RewardTypeIDs: []string{"rt-gems", "rt-gc"}},
		{RankFrom: 2, RankTo: 3, RewardTypeIDs: []string{"rt-gems-small"}},
	}
	rulesJSON, _ := json.Marshal(rules)

	repo.getCampaignByChallenge = func(_ context.Context, _ string) (*model.RewardCampaign, error) {
		return &model.RewardCampaign{
			ID: "camp-1", ChallengeID: "challenge-1",
			Rules: rulesJSON, Status: "active",
		}, nil
	}

	repo.getCampaignByID = func(_ context.Context, _ string) (*model.RewardCampaignFull, error) {
		return &model.RewardCampaignFull{
			ID: "camp-1", ChallengeID: "challenge-1",
			NonGemClaimEmailSubject: "Subject", NonGemClaimEmailBody: "Body",
		}, nil
	}

	repo.getRewardTypeByID = func(_ context.Context, id string) (*model.RewardType, error) {
		switch id {
		case "rt-gems":
			return &model.RewardType{ID: "rt-gems", Type: "gems", Value: 10000, Name: "10K Gems", Stock: 1}, nil
		case "rt-gc":
			return &model.RewardType{ID: "rt-gc", Type: "gift_card", Value: 50, Name: "$50 Gift Card", Stock: 1}, nil
		case "rt-gems-small":
			return &model.RewardType{ID: "rt-gems-small", Type: "gems", Value: 5000, Name: "5K Gems", Stock: 2}, nil
		}
		return nil, model.ErrNotFound
	}

	repo.getUserByID = func(_ context.Context, userID string) (*model.User, error) {
		return &model.User{ID: userID, Username: "user", Email: userID + "@example.com"}, nil
	}

	var distCount int
	repo.insertRewardDistribution = func(_ context.Context, _ *model.RewardDistribution) (string, error) {
		distCount++
		return fmt.Sprintf("dist-%d", distCount), nil
	}
	repo.decrementRewardTypeStock = func(_ context.Context, _ string) error { return nil }
	repo.updateCampaignStatus = func(_ context.Context, _, _ string) error { return nil }

	svc := newTestService(repo)
	svc.now = func() time.Time { return now }

	resp, err := svc.WeeklyReset(context.Background())
	require.NoError(t, err)

	// user-1 gets 2 rewards (gems + gc), user-2 and user-3 each get 1 = 4 total
	assert.Equal(t, 4, resp.RewardsDistributed)
	assert.Equal(t, 4, distCount)
}

// =====================================================================
// RetryFailedEmails tests
// =====================================================================

func TestRetryFailedEmails_NoFailedDistributions(t *testing.T) {
	repo := defaultMockRepo()
	repo.getFailedDistributions = func(_ context.Context) ([]model.FailedDistribution, error) {
		return nil, nil
	}

	svc := newTestService(repo)

	err := svc.RetryFailedEmails(context.Background())
	assert.NoError(t, err)
}

func TestRetryFailedEmails_SuccessfulRetry_MarksDelivered(t *testing.T) {
	repo := defaultMockRepo()
	repo.getFailedDistributions = func(_ context.Context) ([]model.FailedDistribution, error) {
		return []model.FailedDistribution{
			{
				ID: "dist-1", CampaignID: "camp-1", UserID: "user-1",
				Username: "james", Email: "james@example.com",
				RewardTypeName: "$50 Gift Card", RewardTypeValue: 50,
				FinalRank: 1, RetryCount: 0,
				EmailSubject: "Congrats {{username}}!",
				EmailBody:    "You won {{reward_type}}!",
			},
		}, nil
	}

	var deliveredIDs []string
	repo.updateDistributionDelivered = func(_ context.Context, id string) error {
		deliveredIDs = append(deliveredIDs, id)
		return nil
	}

	svc := newTestService(repo)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	err := svc.RetryFailedEmails(context.Background())
	assert.NoError(t, err)

	require.Len(t, deliveredIDs, 1)
	assert.Equal(t, "dist-1", deliveredIDs[0])
}

func TestRetryFailedEmails_FailedRetry_IncrementsCount(t *testing.T) {
	repo := defaultMockRepo()
	repo.getFailedDistributions = func(_ context.Context) ([]model.FailedDistribution, error) {
		return []model.FailedDistribution{
			{
				ID: "dist-1", CampaignID: "camp-1", UserID: "user-1",
				Username: "james", Email: "james@example.com",
				RewardTypeName: "$50 Gift Card", RewardTypeValue: 50,
				FinalRank: 1, RetryCount: 0,
				EmailSubject: "Subject", EmailBody: "Body",
			},
		}, nil
	}

	var incrementedIDs []string
	repo.incrementDistributionRetryCount = func(_ context.Context, id string) (int, error) {
		incrementedIDs = append(incrementedIDs, id)
		return 1, nil
	}

	failEmail := NewEmailService("localhost", 19999, "test@test.com", "pass")
	svc := newTestServiceWithEmail(repo, failEmail)

	err := svc.RetryFailedEmails(context.Background())
	assert.NoError(t, err)

	require.Len(t, incrementedIDs, 1)
	assert.Equal(t, "dist-1", incrementedIDs[0])
}

func TestRetryFailedEmails_ThirdFailure_LogsAlert(t *testing.T) {
	repo := defaultMockRepo()
	repo.getFailedDistributions = func(_ context.Context) ([]model.FailedDistribution, error) {
		return []model.FailedDistribution{
			{
				ID: "dist-1", CampaignID: "camp-1", UserID: "user-1",
				Username: "james", Email: "james@example.com",
				RewardTypeName: "$50 Gift Card", RewardTypeValue: 50,
				FinalRank: 1, RetryCount: 2,
				EmailSubject: "Subject", EmailBody: "Body",
			},
		}, nil
	}

	repo.incrementDistributionRetryCount = func(_ context.Context, _ string) (int, error) {
		return 3, nil
	}

	failEmail := NewEmailService("localhost", 19999, "test@test.com", "pass")
	svc := newTestServiceWithEmail(repo, failEmail)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	err := svc.RetryFailedEmails(context.Background())
	assert.NoError(t, err)

	output := buf.String()
	assert.True(t, strings.Contains(output, "[ALERT] Final email failure for distribution ID: dist-1"))
}
