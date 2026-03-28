package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
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

	// Leaderboard queries
	GetTop99Entries(ctx context.Context, challengeID string) ([]model.LeaderboardEntry, error)
	GetUserEntry(ctx context.Context, challengeID, userID string) (*model.LeaderboardEntry, error)
	GetLastCompletedChallenge(ctx context.Context) (*model.WeeklyChallenge, error)
	GetTop99Results(ctx context.Context, challengeID string) ([]model.WeeklyChallengeResult, error)
	GetUserResult(ctx context.Context, challengeID, userID string) (*model.WeeklyChallengeResult, error)
	GetCampaignByChallenge(ctx context.Context, challengeID string) (*model.RewardCampaign, error)
	GetRewardTypesByIDs(ctx context.Context, ids []string) ([]model.RewardType, error)
	GetResultRewards(ctx context.Context, challengeID string) (map[string][]model.RewardInfo, error)

	// Admin campaign methods
	ListCampaigns(ctx context.Context) ([]model.AdminCampaignListItem, error)
	GetCampaignByID(ctx context.Context, id string) (*model.RewardCampaignFull, error)
	GetRewardTypesByCampaign(ctx context.Context, campaignID string) ([]model.RewardType, error)
	CreateCampaign(ctx context.Context, campaign *model.RewardCampaignFull) (string, error)
	CreateRewardType(ctx context.Context, rt *model.RewardType) (string, error)
	UpdateCampaignRules(ctx context.Context, campaignID string, rules json.RawMessage) error
	UpdateCampaign(ctx context.Context, campaign *model.RewardCampaignFull) error
	DeleteRewardTypesByCampaign(ctx context.Context, campaignID string) error
	GetDistributions(ctx context.Context, campaignID string) ([]model.AdminDistributionRow, error)

	// Weekly reset methods
	CompleteActiveChallenge(ctx context.Context) (string, error)
	GetAllLeaderboardEntries(ctx context.Context, challengeID string) ([]model.LeaderboardEntry, error)
	InsertChallengeResult(ctx context.Context, result *model.WeeklyChallengeResult) error
	InsertRewardDistribution(ctx context.Context, dist *model.RewardDistribution) error
	DecrementRewardTypeStock(ctx context.Context, rewardTypeID string) error
	UpdateCampaignStatus(ctx context.Context, campaignID, status string) error
	InsertWeeklyChallenge(ctx context.Context, startTime, endTime time.Time, status string) (string, error)
	GetRewardTypeByID(ctx context.Context, id string) (*model.RewardType, error)

	// Email retry methods
	GetFailedDistributions(ctx context.Context) ([]model.FailedDistribution, error)
	UpdateDistributionDelivered(ctx context.Context, id string) error
	IncrementDistributionRetryCount(ctx context.Context, id string) (int, error)
}

var allowedEarnSources = map[string]bool{
	"gameplay": true,
	"survey":   true,
	"referral": true,
	"boost":    true,
}

type Service struct {
	repo  Repository
	email *EmailService
	now   func() time.Time
}

func New(repo Repository, email *EmailService) *Service {
	return &Service{repo: repo, email: email, now: func() time.Time { return time.Now().UTC() }}
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
			return err
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

func (s *Service) GetBanner(ctx context.Context, userID string) (*model.BannerResponse, error) {
	if userID == "" {
		return nil, model.ValidationErr("user_id is required")
	}

	challenge, err := s.repo.GetActiveChallenge(ctx)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, nil // signals "no active challenge"
		}
		return nil, err
	}

	top99, err := s.repo.GetTop99Entries(ctx, challenge.ID)
	if err != nil {
		return nil, fmt.Errorf("getting top 99: %w", err)
	}

	resp := &model.BannerResponse{
		ChallengeID: challenge.ID,
		EndTime:     challenge.EndTime,
		RankDisplay: "99+",
	}

	// Find user in top 99
	for i, entry := range top99 {
		if entry.UserID == userID {
			rank := i + 1
			resp.WeeklyGems = entry.WeeklyGems
			resp.DisplayName = entry.DisplayName
			resp.RankDisplay = fmt.Sprintf("#%d", rank)
			if rank > 1 {
				gap := top99[i-1].WeeklyGems - entry.WeeklyGems + 1
				resp.GapToNext = &gap
			}
			return resp, nil
		}
	}

	// User not in top 99 — get their entry directly
	entry, err := s.repo.GetUserEntry(ctx, challenge.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user entry: %w", err)
	}
	if entry != nil {
		resp.WeeklyGems = entry.WeeklyGems
		resp.DisplayName = entry.DisplayName
	}
	return resp, nil
}

func (s *Service) GetCurrentLeaderboard(ctx context.Context, userID string) (*model.CurrentLeaderboardResponse, error) {
	if userID == "" {
		return nil, model.ValidationErr("user_id is required")
	}

	challenge, err := s.repo.GetActiveChallenge(ctx)
	if err != nil {
		return nil, err
	}

	top99, err := s.repo.GetTop99Entries(ctx, challenge.ID)
	if err != nil {
		return nil, fmt.Errorf("getting top 99: %w", err)
	}

	// Build leaderboard rows and find user
	rows := make([]model.CurrentLeaderboardRow, len(top99))
	var currentUser *model.CurrentUserInfo
	for i, entry := range top99 {
		rank := i + 1
		rows[i] = model.CurrentLeaderboardRow{
			Rank: rank, DisplayName: entry.DisplayName, WeeklyGems: entry.WeeklyGems,
		}
		if entry.UserID == userID {
			cu := &model.CurrentUserInfo{
				Rank: &rank, RankDisplay: fmt.Sprintf("%d", rank),
				WeeklyGems: entry.WeeklyGems, DisplayName: entry.DisplayName,
			}
			if rank > 1 {
				gap := top99[i-1].WeeklyGems - entry.WeeklyGems + 1
				cu.GapToNext = &gap
			}
			currentUser = cu
		}
	}

	// User not in top 99
	if currentUser == nil {
		entry, err := s.repo.GetUserEntry(ctx, challenge.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("getting user entry: %w", err)
		}
		cu := &model.CurrentUserInfo{RankDisplay: "99+"}
		if entry != nil {
			cu.WeeklyGems = entry.WeeklyGems
			cu.DisplayName = entry.DisplayName
		}
		currentUser = cu
	}

	resp := &model.CurrentLeaderboardResponse{
		Challenge: model.ChallengeInfo{
			ID: challenge.ID, StartTime: challenge.StartTime, EndTime: challenge.EndTime, Status: challenge.Status,
		},
		Leaderboard: rows,
		CurrentUser: currentUser,
	}

	// Campaign info
	campaign, err := s.repo.GetCampaignByChallenge(ctx, challenge.ID)
	if err != nil {
		return nil, fmt.Errorf("getting campaign: %w", err)
	}
	if campaign != nil {
		summary, err := s.buildCampaignSummary(ctx, campaign)
		if err != nil {
			return nil, fmt.Errorf("building campaign summary: %w", err)
		}
		resp.Campaign = summary
	}

	return resp, nil
}

func (s *Service) buildCampaignSummary(ctx context.Context, campaign *model.RewardCampaign) (*model.CampaignSummary, error) {
	var rules []model.RewardRule
	if err := json.Unmarshal(campaign.Rules, &rules); err != nil {
		return nil, fmt.Errorf("parsing campaign rules: %w", err)
	}

	// Collect all reward type IDs
	idSet := make(map[string]bool)
	for _, rule := range rules {
		for _, id := range rule.RewardTypeIDs {
			idSet[id] = true
		}
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	rewardTypes, err := s.repo.GetRewardTypesByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("getting reward types: %w", err)
	}
	rtMap := make(map[string]model.RewardType, len(rewardTypes))
	for _, rt := range rewardTypes {
		rtMap[rt.ID] = rt
	}

	summaryRows := make([]model.RewardsSummaryRow, 0, len(rules))
	for _, rule := range rules {
		rewards := make([]model.RewardInfo, 0, len(rule.RewardTypeIDs))
		for _, rtID := range rule.RewardTypeIDs {
			if rt, ok := rtMap[rtID]; ok {
				rewards = append(rewards, model.RewardInfo{
					Name: rt.Name, Image: rt.Image, Value: rt.Value, Type: rt.Type,
				})
			}
		}
		summaryRows = append(summaryRows, model.RewardsSummaryRow{
			RankFrom: rule.RankFrom, RankTo: rule.RankTo, Rewards: rewards,
		})
	}

	return &model.CampaignSummary{
		BannerImage:    campaign.BannerImage,
		RewardsSummary: summaryRows,
	}, nil
}

func (s *Service) GetLastWeekLeaderboard(ctx context.Context, userID string) (*model.LastWeekResponse, error) {
	if userID == "" {
		return nil, model.ValidationErr("user_id is required")
	}

	challenge, err := s.repo.GetLastCompletedChallenge(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting last completed challenge: %w", err)
	}
	if challenge == nil {
		return &model.LastWeekResponse{Challenge: nil}, nil
	}

	results, err := s.repo.GetTop99Results(ctx, challenge.ID)
	if err != nil {
		return nil, fmt.Errorf("getting top 99 results: %w", err)
	}

	rewards, err := s.repo.GetResultRewards(ctx, challenge.ID)
	if err != nil {
		return nil, fmt.Errorf("getting result rewards: %w", err)
	}

	rows := make([]model.LastWeekRow, len(results))
	var currentUser *model.LastWeekUserInfo
	for i, result := range results {
		userRewards := rewards[result.UserID]
		if userRewards == nil {
			userRewards = []model.RewardInfo{}
		}
		rows[i] = model.LastWeekRow{
			Rank: result.FinalRank, DisplayName: result.DisplayName,
			FinalGems: result.FinalGems, Rewards: userRewards,
		}
		if result.UserID == userID {
			rank := result.FinalRank
			currentUser = &model.LastWeekUserInfo{
				Rank: &rank, RankDisplay: fmt.Sprintf("%d", rank),
				FinalGems: result.FinalGems, Rewards: userRewards,
			}
		}
	}

	// User not in top 99 results — check if they have a result at all
	if currentUser == nil {
		userResult, err := s.repo.GetUserResult(ctx, challenge.ID, userID)
		if err != nil {
			return nil, fmt.Errorf("getting user result: %w", err)
		}
		if userResult != nil {
			userRewards := rewards[userResult.UserID]
			if userRewards == nil {
				userRewards = []model.RewardInfo{}
			}
			currentUser = &model.LastWeekUserInfo{
				RankDisplay: "99+", FinalGems: userResult.FinalGems, Rewards: userRewards,
			}
		}
	}

	return &model.LastWeekResponse{
		Challenge: &model.LastWeekChallengeInfo{
			ID: challenge.ID, StartTime: challenge.StartTime, EndTime: challenge.EndTime,
		},
		Leaderboard: rows,
		CurrentUser: currentUser,
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
			return err
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

// =====================================================================
// Admin campaign methods
// =====================================================================

func (s *Service) ListCampaigns(ctx context.Context) ([]model.AdminCampaignListItem, error) {
	items, err := s.repo.ListCampaigns(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing campaigns: %w", err)
	}
	return items, nil
}

func (s *Service) GetCampaign(ctx context.Context, id string) (*model.AdminCampaignDetail, error) {
	if id == "" {
		return nil, model.ValidationErr("campaign id is required")
	}

	campaign, err := s.repo.GetCampaignByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting campaign: %w", err)
	}
	if campaign == nil {
		return nil, model.NotFoundErr("campaign not found")
	}

	rewardTypes, err := s.repo.GetRewardTypesByCampaign(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting reward types: %w", err)
	}

	// Build reward type lookup
	rtMap := make(map[string]model.RewardType, len(rewardTypes))
	for _, rt := range rewardTypes {
		rtMap[rt.ID] = rt
	}

	// Parse rules and resolve reward names
	var rawRules []model.RewardRule
	if campaign.Rules != nil {
		if err := json.Unmarshal(campaign.Rules, &rawRules); err != nil {
			return nil, fmt.Errorf("parsing campaign rules: %w", err)
		}
	}

	ruleDetails := make([]model.AdminCampaignRuleDetail, 0, len(rawRules))
	for _, rule := range rawRules {
		names := make([]string, 0, len(rule.RewardTypeIDs))
		rewards := make([]model.RewardInfo, 0, len(rule.RewardTypeIDs))
		for _, rtID := range rule.RewardTypeIDs {
			if rt, ok := rtMap[rtID]; ok {
				names = append(names, rt.Name)
				rewards = append(rewards, model.RewardInfo{
					Name: rt.Name, Image: rt.Image, Value: rt.Value, Type: rt.Type,
				})
			}
		}
		ruleDetails = append(ruleDetails, model.AdminCampaignRuleDetail{
			RankFrom: rule.RankFrom, RankTo: rule.RankTo,
			RewardNames: names, RewardTypes: rewards,
		})
	}

	return &model.AdminCampaignDetail{
		ID:                      campaign.ID,
		ChallengeID:             campaign.ChallengeID,
		Name:                    campaign.Name,
		BannerImage:             campaign.BannerImage,
		Status:                  campaign.Status,
		NonGemClaimEmailSubject: campaign.NonGemClaimEmailSubject,
		NonGemClaimEmailBody:    campaign.NonGemClaimEmailBody,
		RewardTypes:             rewardTypes,
		Rules:                   ruleDetails,
	}, nil
}

// validateCampaignRequest validates the shared rules for create/update campaign.
func validateCampaignRequest(req model.CreateCampaignRequest) error {
	if req.ChallengeID == "" {
		return model.ValidationErr("challenge_id is required")
	}
	if req.Name == "" {
		return model.ValidationErr("name is required")
	}
	if len(req.RewardTypes) == 0 {
		return model.ValidationErr("at least one reward_type is required")
	}
	if len(req.Rules) == 0 {
		return model.ValidationErr("at least one rule is required")
	}

	// Validate reward_type_indexes and rank constraints
	for _, rule := range req.Rules {
		if rule.RankFrom < 1 {
			return model.ValidationErr("rank_from must be >= 1")
		}
		if rule.RankFrom > rule.RankTo {
			return model.ValidationErr("rank_from must be <= rank_to")
		}
		for _, idx := range rule.RewardTypeIndexes {
			if idx < 0 || idx >= len(req.RewardTypes) {
				return model.ValidationErr(fmt.Sprintf("invalid reward_type_index: %d", idx))
			}
		}
	}

	// Check overlapping ranges: sort by rank_from, then check each rank_from > previous rank_to
	sorted := make([]model.CreateCampaignRuleInput, len(req.Rules))
	copy(sorted, req.Rules)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].RankFrom < sorted[j].RankFrom })
	for i := 1; i < len(sorted); i++ {
		if sorted[i].RankFrom <= sorted[i-1].RankTo {
			return model.ValidationErr("overlapping rank ranges")
		}
	}

	// Stock validation: for each reward_type, sum total users across all ranges that use it
	usersPerRewardType := make(map[int]int) // index → total users
	for _, rule := range req.Rules {
		usersInRange := rule.RankTo - rule.RankFrom + 1
		for _, idx := range rule.RewardTypeIndexes {
			usersPerRewardType[idx] += usersInRange
		}
	}
	for idx, totalUsers := range usersPerRewardType {
		rt := req.RewardTypes[idx]
		if totalUsers > rt.Stock {
			return model.ValidationErr(fmt.Sprintf("%s: need %d, stock is %d", rt.Name, totalUsers, rt.Stock))
		}
	}

	return nil
}

func (s *Service) CreateCampaign(ctx context.Context, req model.CreateCampaignRequest) (*model.AdminCampaignDetail, error) {
	if err := validateCampaignRequest(req); err != nil {
		return nil, err
	}

	var campaignID string
	err := s.repo.RunInTx(ctx, func(ctx context.Context) error {
		// 1. Create campaign record (rules will be updated after reward_types are created)
		campaign := &model.RewardCampaignFull{
			ChallengeID:             req.ChallengeID,
			Name:                    req.Name,
			BannerImage:             req.BannerImage,
			Status:                  "draft",
			NonGemClaimEmailSubject: req.NonGemClaimEmailSubject,
			NonGemClaimEmailBody:    req.NonGemClaimEmailBody,
		}
		var err error
		campaignID, err = s.repo.CreateCampaign(ctx, campaign)
		if err != nil {
			return fmt.Errorf("creating campaign: %w", err)
		}

		// 2. Create reward_types, collect UUIDs
		rtIDs := make([]string, len(req.RewardTypes))
		for i, rtInput := range req.RewardTypes {
			rt := &model.RewardType{
				CampaignID: campaignID,
				Name:       rtInput.Name,
				Type:       rtInput.Type,
				Value:      rtInput.Value,
				Image:      rtInput.Image,
				Stock:      rtInput.Stock,
			}
			rtIDs[i], err = s.repo.CreateRewardType(ctx, rt)
			if err != nil {
				return fmt.Errorf("creating reward type: %w", err)
			}
		}

		// 3. Map indexes to UUIDs and build rules JSONB
		rules := make([]model.RewardRule, len(req.Rules))
		for i, ruleInput := range req.Rules {
			rewardTypeIDs := make([]string, len(ruleInput.RewardTypeIndexes))
			for j, idx := range ruleInput.RewardTypeIndexes {
				rewardTypeIDs[j] = rtIDs[idx]
			}
			rules[i] = model.RewardRule{
				RankFrom:      ruleInput.RankFrom,
				RankTo:        ruleInput.RankTo,
				RewardTypeIDs: rewardTypeIDs,
			}
		}

		// 4. Update campaign with final rules JSONB
		rulesJSON, err := json.Marshal(rules)
		if err != nil {
			return fmt.Errorf("marshaling rules: %w", err)
		}
		if err := s.repo.UpdateCampaignRules(ctx, campaignID, rulesJSON); err != nil {
			return fmt.Errorf("updating campaign rules: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetCampaign(ctx, campaignID)
}

func (s *Service) UpdateCampaign(ctx context.Context, id string, req model.CreateCampaignRequest) (*model.AdminCampaignDetail, error) {
	if id == "" {
		return nil, model.ValidationErr("campaign id is required")
	}

	if err := validateCampaignRequest(req); err != nil {
		return nil, err
	}

	// Verify campaign exists
	existing, err := s.repo.GetCampaignByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting campaign: %w", err)
	}
	if existing == nil {
		return nil, model.NotFoundErr("campaign not found")
	}

	// Prevent updating campaigns that already have distributions (would cascade-delete them)
	dists, err := s.repo.GetDistributions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("checking distributions: %w", err)
	}
	if len(dists) > 0 {
		return nil, model.ValidationErr("cannot update campaign that has reward distributions")
	}

	err = s.repo.RunInTx(ctx, func(ctx context.Context) error {
		// Delete existing reward_types
		if err := s.repo.DeleteRewardTypesByCampaign(ctx, id); err != nil {
			return fmt.Errorf("deleting reward types: %w", err)
		}

		// Create new reward_types
		rtIDs := make([]string, len(req.RewardTypes))
		for i, rtInput := range req.RewardTypes {
			rt := &model.RewardType{
				CampaignID: id,
				Name:       rtInput.Name,
				Type:       rtInput.Type,
				Value:      rtInput.Value,
				Image:      rtInput.Image,
				Stock:      rtInput.Stock,
			}
			var err error
			rtIDs[i], err = s.repo.CreateRewardType(ctx, rt)
			if err != nil {
				return fmt.Errorf("creating reward type: %w", err)
			}
		}

		// Build rules JSONB
		rules := make([]model.RewardRule, len(req.Rules))
		for i, ruleInput := range req.Rules {
			rewardTypeIDs := make([]string, len(ruleInput.RewardTypeIndexes))
			for j, idx := range ruleInput.RewardTypeIndexes {
				rewardTypeIDs[j] = rtIDs[idx]
			}
			rules[i] = model.RewardRule{
				RankFrom:      ruleInput.RankFrom,
				RankTo:        ruleInput.RankTo,
				RewardTypeIDs: rewardTypeIDs,
			}
		}
		rulesJSON, err := json.Marshal(rules)
		if err != nil {
			return fmt.Errorf("marshaling rules: %w", err)
		}

		// Update campaign fields
		campaign := &model.RewardCampaignFull{
			ID:                      id,
			ChallengeID:             req.ChallengeID,
			Name:                    req.Name,
			BannerImage:             req.BannerImage,
			Rules:                   rulesJSON,
			Status:                  existing.Status,
			NonGemClaimEmailSubject: req.NonGemClaimEmailSubject,
			NonGemClaimEmailBody:    req.NonGemClaimEmailBody,
		}
		if err := s.repo.UpdateCampaign(ctx, campaign); err != nil {
			return fmt.Errorf("updating campaign: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetCampaign(ctx, id)
}

func (s *Service) GetDistributions(ctx context.Context, campaignID string) ([]model.AdminDistributionRow, error) {
	if campaignID == "" {
		return nil, model.ValidationErr("campaign_id is required")
	}

	rows, err := s.repo.GetDistributions(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("getting distributions: %w", err)
	}
	return rows, nil
}
