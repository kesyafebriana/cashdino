package model

import (
	"encoding/json"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type GemHistory struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Source    string    `json:"source"`
	Amount    int       `json:"amount"`
	GameName  *string   `json:"game_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type WeeklyChallenge struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}

type LeaderboardEntry struct {
	ID               string     `json:"id"`
	ChallengeID      string     `json:"challenge_id"`
	UserID           string     `json:"user_id"`
	WeeklyGems       int        `json:"weekly_gems"`
	FirstGemEarnedAt *time.Time `json:"first_gem_earned_at,omitempty"`
	DisplayName      string     `json:"display_name"`
}

type DailyCheckin struct {
	ID              string    `json:"id"`
	Date            time.Time `json:"date"`
	BaseGems        int       `json:"base_gems"`
	StreakMultiplier float64   `json:"streak_multiplier"`
	IsActive        bool      `json:"is_active"`
}

type UserDailyCheckin struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	CheckinID     string    `json:"checkin_id"`
	GemsEarned    int       `json:"gems_earned"`
	CurrentStreak int       `json:"current_streak"`
	CheckedInAt   time.Time `json:"checked_in_at"`
}

type WeeklyChallengeResult struct {
	ID          string `json:"id"`
	ChallengeID string `json:"challenge_id"`
	UserID      string `json:"user_id"`
	FinalRank   int    `json:"final_rank"`
	FinalGems   int    `json:"final_gems"`
	DisplayName string `json:"display_name"`
}

type RewardCampaign struct {
	ID          string          `json:"id"`
	ChallengeID string          `json:"challenge_id"`
	Name        string          `json:"name"`
	BannerImage string          `json:"banner_image"`
	Rules       json.RawMessage `json:"rules"`
	Status      string          `json:"status"`
}

type RewardType struct {
	ID         string  `json:"id"`
	CampaignID string  `json:"campaign_id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Value      float64 `json:"value"`
	Image      *string `json:"image"`
	Stock      int     `json:"stock"`
}

type RewardRule struct {
	RankFrom      int      `json:"rank_from"`
	RankTo        int      `json:"rank_to"`
	RewardTypeIDs []string `json:"reward_type_ids"`
}

// Request/response types

type EarnGemsRequest struct {
	UserID   string  `json:"user_id"`
	Source   string  `json:"source"`
	Amount   int     `json:"amount"`
	GameName *string `json:"game_name,omitempty"`
}

type EarnGemsResponse struct {
	UserID     string `json:"user_id"`
	WeeklyGems int    `json:"weekly_gems"`
}

type CheckinRequest struct {
	UserID string `json:"user_id"`
}

type CheckinResponse struct {
	GemsEarned    int `json:"gems_earned"`
	CurrentStreak int `json:"current_streak"`
	WeeklyGems    int `json:"weekly_gems"`
}

// Banner endpoint response
type BannerResponse struct {
	ChallengeID string    `json:"challenge_id"`
	EndTime     time.Time `json:"end_time"`
	WeeklyGems  int       `json:"weekly_gems"`
	RankDisplay string    `json:"rank_display"`
	GapToNext   *int      `json:"gap_to_next"`
	DisplayName string    `json:"display_name"`
}

// Current leaderboard endpoint response
type CurrentLeaderboardResponse struct {
	Challenge   ChallengeInfo          `json:"challenge"`
	Leaderboard []CurrentLeaderboardRow `json:"leaderboard"`
	CurrentUser *CurrentUserInfo       `json:"current_user"`
	Campaign    *CampaignSummary       `json:"campaign"`
}

type ChallengeInfo struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}

type CurrentLeaderboardRow struct {
	Rank        int    `json:"rank"`
	DisplayName string `json:"display_name"`
	WeeklyGems  int    `json:"weekly_gems"`
}

type CurrentUserInfo struct {
	Rank        *int   `json:"rank"`
	RankDisplay string `json:"rank_display"`
	WeeklyGems  int    `json:"weekly_gems"`
	GapToNext   *int   `json:"gap_to_next"`
	DisplayName string `json:"display_name"`
}

type CampaignSummary struct {
	BannerImage    string              `json:"banner_image"`
	RewardsSummary []RewardsSummaryRow `json:"rewards_summary"`
}

type RewardsSummaryRow struct {
	RankFrom int          `json:"rank_from"`
	RankTo   int          `json:"rank_to"`
	Rewards  []RewardInfo `json:"rewards"`
}

type RewardInfo struct {
	Name  string  `json:"name"`
	Image *string `json:"image"`
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// Last week leaderboard endpoint response
type LastWeekResponse struct {
	Challenge   *LastWeekChallengeInfo `json:"challenge"`
	Leaderboard []LastWeekRow         `json:"leaderboard,omitempty"`
	CurrentUser *LastWeekUserInfo     `json:"current_user,omitempty"`
}

type LastWeekChallengeInfo struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type LastWeekRow struct {
	Rank        int          `json:"rank"`
	DisplayName string       `json:"display_name"`
	FinalGems   int          `json:"final_gems"`
	Rewards     []RewardInfo `json:"rewards"`
}

type LastWeekUserInfo struct {
	Rank        *int         `json:"rank"`
	RankDisplay string       `json:"rank_display"`
	FinalGems   int          `json:"final_gems"`
	Rewards     []RewardInfo `json:"rewards"`
}

// Admin campaign request/response types

type CreateCampaignRequest struct {
	ChallengeID            string                   `json:"challenge_id"`
	Name                   string                   `json:"name"`
	BannerImage            string                   `json:"banner_image"`
	RewardTypes            []CreateRewardTypeInput   `json:"reward_types"`
	Rules                  []CreateCampaignRuleInput `json:"rules"`
	NonGemClaimEmailSubject string                  `json:"non_gem_claim_email_subject"`
	NonGemClaimEmailBody    string                  `json:"non_gem_claim_email_body"`
}

type CreateRewardTypeInput struct {
	Name  string  `json:"name"`
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Image *string `json:"image"`
	Stock int     `json:"stock"`
}

type CreateCampaignRuleInput struct {
	RankFrom          int   `json:"rank_from"`
	RankTo            int   `json:"rank_to"`
	RewardTypeIndexes []int `json:"reward_type_indexes"`
}

type AdminCampaignListItem struct {
	ID               string    `json:"id"`
	ChallengeID      string    `json:"challenge_id"`
	Name             string    `json:"name"`
	BannerImage      string    `json:"banner_image"`
	Status           string    `json:"status"`
	ChallengeStart   time.Time `json:"challenge_start"`
	ChallengeEnd     time.Time `json:"challenge_end"`
	RewardTypesCount int       `json:"reward_types_count"`
	TotalStock       int       `json:"total_stock"`
}

type AdminCampaignDetail struct {
	ID                      string                    `json:"id"`
	ChallengeID             string                    `json:"challenge_id"`
	Name                    string                    `json:"name"`
	BannerImage             string                    `json:"banner_image"`
	Status                  string                    `json:"status"`
	NonGemClaimEmailSubject string                    `json:"non_gem_claim_email_subject"`
	NonGemClaimEmailBody    string                    `json:"non_gem_claim_email_body"`
	RewardTypes             []RewardType              `json:"reward_types"`
	Rules                   []AdminCampaignRuleDetail `json:"rules"`
}

type AdminCampaignRuleDetail struct {
	RankFrom    int          `json:"rank_from"`
	RankTo      int          `json:"rank_to"`
	RewardNames []string     `json:"reward_names"`
	RewardTypes []RewardInfo `json:"reward_types"`
}

type AdminDistributionRow struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	DisplayName    string     `json:"display_name"`
	MaskedEmail    string     `json:"masked_email"`
	RewardName     string     `json:"reward_name"`
	RewardType     string     `json:"reward_type"`
	RewardValue    float64    `json:"reward_value"`
	RewardImage    *string    `json:"reward_image"`
	Status         string     `json:"status"`
	DeliveredAt    *time.Time `json:"delivered_at"`
	EmailSentAt    *time.Time `json:"email_sent_at"`
	FinalRank      int        `json:"final_rank"`
}

// RewardCampaignFull includes email template fields not in the base RewardCampaign
type RewardCampaignFull struct {
	ID                      string          `json:"id"`
	ChallengeID             string          `json:"challenge_id"`
	Name                    string          `json:"name"`
	BannerImage             string          `json:"banner_image"`
	Rules                   json.RawMessage `json:"rules"`
	Status                  string          `json:"status"`
	NonGemClaimEmailSubject string          `json:"non_gem_claim_email_subject"`
	NonGemClaimEmailBody    string          `json:"non_gem_claim_email_body"`
}
