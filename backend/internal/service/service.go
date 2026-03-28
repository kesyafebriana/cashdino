package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/kesyafebriana/cashdino/backend/internal/model"
)

type Repository interface {
	GetUserByID(ctx context.Context, userID string) (*model.User, error)
	InsertGemHistory(ctx context.Context, userID, source string, amount int, gameName *string) error
	GetActiveChallenge(ctx context.Context) (*model.WeeklyChallenge, error)
	UpsertLeaderboardEntry(ctx context.Context, challengeID, userID, displayName string, amount int) (int, error)
	GenerateDisplayName(ctx context.Context, challengeID, username string) (string, error)
	GetTodayCheckin(ctx context.Context, today time.Time) (*model.DailyCheckin, error)
	HasCheckedInToday(ctx context.Context, userID, checkinID string) (bool, error)
	GetLastCheckin(ctx context.Context, userID string) (*model.UserDailyCheckin, error)
	GetCheckinDate(ctx context.Context, checkinID string) (time.Time, error)
	InsertUserDailyCheckin(ctx context.Context, userID, checkinID string, gemsEarned, currentStreak int) error
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

var allowedEarnSources = map[string]bool{
	"gameplay": true,
	"survey":   true,
	"referral": true,
	"boost":    true,
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func New(repo Repository) *Service {
	return &Service{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) EarnGems(ctx context.Context, req model.EarnGemsRequest) (*model.EarnGemsResponse, error) {
	if req.UserID == "" {
		return nil, model.ValidationErr("user_id is required")
	}
	if !allowedEarnSources[req.Source] {
		return nil, model.ValidationErr("invalid source: must be one of gameplay, survey, referral, boost")
	}
	if req.Amount <= 0 {
		return nil, model.ValidationErr("amount must be greater than 0")
	}

	// Verify user exists and get username for display name
	user, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("validating user: %w", err)
	}

	// All writes in a single transaction
	var weeklyGems int
	err = s.repo.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.InsertGemHistory(ctx, req.UserID, req.Source, req.Amount, req.GameName); err != nil {
			return fmt.Errorf("recording gem history: %w", err)
		}

		challenge, err := s.repo.GetActiveChallenge(ctx)
		if err != nil {
			return fmt.Errorf("getting active challenge: %w", err)
		}

		displayName, err := s.repo.GenerateDisplayName(ctx, challenge.ID, user.Username)
		if err != nil {
			return fmt.Errorf("generating display name: %w", err)
		}

		weeklyGems, err = s.repo.UpsertLeaderboardEntry(ctx, challenge.ID, req.UserID, displayName, req.Amount)
		if err != nil {
			return fmt.Errorf("updating leaderboard: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &model.EarnGemsResponse{
		UserID:     req.UserID,
		WeeklyGems: weeklyGems,
	}, nil
}

func (s *Service) Checkin(ctx context.Context, req model.CheckinRequest) (*model.CheckinResponse, error) {
	if req.UserID == "" {
		return nil, model.ValidationErr("user_id is required")
	}

	// Verify user exists
	user, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("validating user: %w", err)
	}

	// Get today's checkin config
	now := s.now()
	todayCheckin, err := s.repo.GetTodayCheckin(ctx, now)
	if err != nil {
		return nil, fmt.Errorf("getting today checkin config: %w", err)
	}
	if todayCheckin == nil || !todayCheckin.IsActive {
		return nil, model.ValidationErr("no check-in available today")
	}

	// Check if already checked in today
	alreadyCheckedIn, err := s.repo.HasCheckedInToday(ctx, req.UserID, todayCheckin.ID)
	if err != nil {
		return nil, fmt.Errorf("checking existing checkin: %w", err)
	}
	if alreadyCheckedIn {
		return nil, model.ValidationErr("already checked in today")
	}

	// Calculate streak
	currentStreak := 1
	lastCheckin, err := s.repo.GetLastCheckin(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("getting last checkin: %w", err)
	}
	if lastCheckin != nil {
		lastDate, err := s.repo.GetCheckinDate(ctx, lastCheckin.CheckinID)
		if err != nil {
			return nil, fmt.Errorf("getting last checkin date: %w", err)
		}
		yesterday := now.AddDate(0, 0, -1)
		if lastDate.Format("2006-01-02") == yesterday.Format("2006-01-02") {
			currentStreak = lastCheckin.CurrentStreak + 1
		}
	}

	// Calculate gems earned
	gemsEarned := int(math.Round(float64(todayCheckin.BaseGems) * todayCheckin.StreakMultiplier))

	// All writes in a single transaction
	var weeklyGems int
	err = s.repo.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.repo.InsertUserDailyCheckin(ctx, req.UserID, todayCheckin.ID, gemsEarned, currentStreak); err != nil {
			return fmt.Errorf("recording checkin: %w", err)
		}

		if err := s.repo.InsertGemHistory(ctx, req.UserID, "daily_checkin", gemsEarned, nil); err != nil {
			return fmt.Errorf("recording gem history for checkin: %w", err)
		}

		challenge, err := s.repo.GetActiveChallenge(ctx)
		if err != nil {
			return fmt.Errorf("getting active challenge: %w", err)
		}

		displayName, err := s.repo.GenerateDisplayName(ctx, challenge.ID, user.Username)
		if err != nil {
			return fmt.Errorf("generating display name: %w", err)
		}

		weeklyGems, err = s.repo.UpsertLeaderboardEntry(ctx, challenge.ID, req.UserID, displayName, gemsEarned)
		if err != nil {
			return fmt.Errorf("updating leaderboard for checkin: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &model.CheckinResponse{
		GemsEarned:    gemsEarned,
		CurrentStreak: currentStreak,
		WeeklyGems:    weeklyGems,
	}, nil
}
