package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kesyafebriana/cashdino/backend/internal/model"
)

// CompleteActiveChallenge sets the active challenge to 'completed' and returns its ID.
func (r *Repository) CompleteActiveChallenge(ctx context.Context) (string, error) {
	var id string
	err := r.getDB(ctx).QueryRow(ctx,
		`UPDATE weekly_challenges SET status = 'completed' WHERE status = 'active' RETURNING id`,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("completing active challenge: %w", model.ErrNotFound)
		}
		return "", fmt.Errorf("completing active challenge: %w", err)
	}
	return id, nil
}

// GetAllLeaderboardEntries returns all entries with weekly_gems > 0 for a challenge,
// ordered by weekly_gems DESC, first_gem_earned_at ASC.
func (r *Repository) GetAllLeaderboardEntries(ctx context.Context, challengeID string) ([]model.LeaderboardEntry, error) {
	rows, err := r.getDB(ctx).Query(ctx,
		`SELECT id, challenge_id, user_id, weekly_gems, first_gem_earned_at, display_name
		 FROM leaderboard_entries
		 WHERE challenge_id = $1 AND weekly_gems > 0
		 ORDER BY weekly_gems DESC, first_gem_earned_at ASC`,
		challengeID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying all leaderboard entries: %w", err)
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

// InsertChallengeResult inserts a row into weekly_challenge_results.
func (r *Repository) InsertChallengeResult(ctx context.Context, result *model.WeeklyChallengeResult) error {
	_, err := r.getDB(ctx).Exec(ctx,
		`INSERT INTO weekly_challenge_results (challenge_id, user_id, final_rank, final_gems, display_name)
		 VALUES ($1, $2, $3, $4, $5)`,
		result.ChallengeID, result.UserID, result.FinalRank, result.FinalGems, result.DisplayName,
	)
	if err != nil {
		return fmt.Errorf("inserting challenge result: %w", err)
	}
	return nil
}

// InsertRewardDistribution inserts a row into reward_distributions and returns the new ID.
func (r *Repository) InsertRewardDistribution(ctx context.Context, dist *model.RewardDistribution) (string, error) {
	var id string
	err := r.getDB(ctx).QueryRow(ctx,
		`INSERT INTO reward_distributions (campaign_id, user_id, reward_type_id, status, delivered_at, email_sent_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		dist.CampaignID, dist.UserID, dist.RewardTypeID, dist.Status, dist.DeliveredAt, dist.EmailSentAt,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("inserting reward distribution: %w", err)
	}
	return id, nil
}

// DecrementRewardTypeStock decreases the stock of a reward type by 1.
// Returns ErrValidation if stock is already 0.
func (r *Repository) DecrementRewardTypeStock(ctx context.Context, rewardTypeID string) error {
	tag, err := r.getDB(ctx).Exec(ctx,
		`UPDATE reward_types SET stock = stock - 1 WHERE id = $1 AND stock > 0`,
		rewardTypeID,
	)
	if err != nil {
		return fmt.Errorf("decrementing reward type stock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("decrementing reward type stock: %w", model.ValidationErr("reward type stock is 0"))
	}
	return nil
}

// UpdateCampaignStatus updates the status field of a reward campaign.
func (r *Repository) UpdateCampaignStatus(ctx context.Context, campaignID, status string) error {
	_, err := r.getDB(ctx).Exec(ctx,
		`UPDATE reward_campaigns SET status = $1 WHERE id = $2`,
		status, campaignID,
	)
	if err != nil {
		return fmt.Errorf("updating campaign status: %w", err)
	}
	return nil
}

// InsertWeeklyChallenge creates a new weekly challenge and returns its ID.
func (r *Repository) InsertWeeklyChallenge(ctx context.Context, startTime, endTime time.Time, status string) (string, error) {
	var id string
	err := r.getDB(ctx).QueryRow(ctx,
		`INSERT INTO weekly_challenges (start_time, end_time, status) VALUES ($1, $2, $3) RETURNING id`,
		startTime, endTime, status,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("inserting weekly challenge: %w", err)
	}
	return id, nil
}

// GetRewardTypeByID returns a single reward type by ID.
func (r *Repository) GetRewardTypeByID(ctx context.Context, id string) (*model.RewardType, error) {
	var rt model.RewardType
	err := r.getDB(ctx).QueryRow(ctx,
		`SELECT id, campaign_id, name, type, value, image, stock FROM reward_types WHERE id = $1`,
		id,
	).Scan(&rt.ID, &rt.CampaignID, &rt.Name, &rt.Type, &rt.Value, &rt.Image, &rt.Stock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting reward type by id: %w", model.ErrNotFound)
		}
		return nil, fmt.Errorf("getting reward type by id: %w", err)
	}
	return &rt, nil
}

// GetFailedDistributions returns all failed distributions joined with campaign, user, and reward_type info.
func (r *Repository) GetFailedDistributions(ctx context.Context) ([]model.FailedDistribution, error) {
	rows, err := r.getDB(ctx).Query(ctx,
		`SELECT rd.id, rd.campaign_id, rd.user_id, u.username, u.email,
		        rd.reward_type_id, rt.name, rt.value, rt.image,
		        wcr.final_rank,
		        rc.non_gem_claim_email_subject, rc.non_gem_claim_email_body,
		        rd.retry_count
		 FROM reward_distributions rd
		 JOIN users u ON rd.user_id = u.id
		 JOIN reward_types rt ON rd.reward_type_id = rt.id
		 JOIN reward_campaigns rc ON rd.campaign_id = rc.id
		 JOIN weekly_challenge_results wcr ON wcr.challenge_id = rc.challenge_id AND wcr.user_id = rd.user_id
		 WHERE rd.status = 'failed' AND rd.retry_count < 3`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying failed distributions: %w", err)
	}
	defer rows.Close()

	var dists []model.FailedDistribution
	for rows.Next() {
		var d model.FailedDistribution
		if err := rows.Scan(&d.ID, &d.CampaignID, &d.UserID, &d.Username, &d.Email,
			&d.RewardTypeID, &d.RewardTypeName, &d.RewardTypeValue, &d.RewardTypeImage,
			&d.FinalRank,
			&d.EmailSubject, &d.EmailBody,
			&d.RetryCount); err != nil {
			return nil, fmt.Errorf("scanning failed distribution: %w", err)
		}
		dists = append(dists, d)
	}
	return dists, rows.Err()
}

// UpdateDistributionDelivered marks a distribution as delivered with timestamps.
func (r *Repository) UpdateDistributionDelivered(ctx context.Context, id string) error {
	_, err := r.getDB(ctx).Exec(ctx,
		`UPDATE reward_distributions SET status = 'delivered', delivered_at = NOW(), email_sent_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("updating distribution delivered: %w", err)
	}
	return nil
}

// UpdateDistributionFailed marks a distribution as failed.
func (r *Repository) UpdateDistributionFailed(ctx context.Context, id string) error {
	_, err := r.getDB(ctx).Exec(ctx,
		`UPDATE reward_distributions SET status = 'failed' WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("updating distribution failed: %w", err)
	}
	return nil
}

// IncrementDistributionRetryCount increments retry_count and returns the new value.
func (r *Repository) IncrementDistributionRetryCount(ctx context.Context, id string) (int, error) {
	var count int
	err := r.getDB(ctx).QueryRow(ctx,
		`UPDATE reward_distributions SET retry_count = retry_count + 1 WHERE id = $1 RETURNING retry_count`,
		id,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("incrementing distribution retry count: %w", err)
	}
	return count, nil
}
