package model

import "time"

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
	ID                string     `json:"id"`
	ChallengeID       string     `json:"challenge_id"`
	UserID            string     `json:"user_id"`
	WeeklyGems        int        `json:"weekly_gems"`
	FirstGemEarnedAt  *time.Time `json:"first_gem_earned_at,omitempty"`
	DisplayName       string     `json:"display_name"`
}

type DailyCheckin struct {
	ID               string    `json:"id"`
	Date             time.Time `json:"date"`
	BaseGems         int     `json:"base_gems"`
	StreakMultiplier  float64 `json:"streak_multiplier"`
	IsActive         bool    `json:"is_active"`
}

type UserDailyCheckin struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	CheckinID     string    `json:"checkin_id"`
	GemsEarned    int       `json:"gems_earned"`
	CurrentStreak int       `json:"current_streak"`
	CheckedInAt   time.Time `json:"checked_in_at"`
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
