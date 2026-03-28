package service

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"strings"
	"time"

	"github.com/kesyafebriana/cashdino/backend/internal/model"
)

// pendingEmail holds info needed to send an email after the DB transaction commits.
type pendingEmail struct {
	distributionID string
	to             string
	subject        string
	body           string
}

// WeeklyReset performs the end-of-week processing:
// 1. Lock current challenge
// 2. Snapshot results
// 3. Distribute rewards (emails deferred until after commit)
// 4. Create next challenge
func (s *Service) WeeklyReset(ctx context.Context) (*model.WeeklyResetResponse, error) {
	var resp model.WeeklyResetResponse
	var pendingEmails []pendingEmail

	err := s.repo.RunInTx(ctx, func(ctx context.Context) error {
		// Step 1: Lock current challenge
		challengeID, err := s.repo.CompleteActiveChallenge(ctx)
		if err != nil {
			return fmt.Errorf("locking challenge: %w", err)
		}

		// Step 2: Snapshot results
		entries, err := s.repo.GetAllLeaderboardEntries(ctx, challengeID)
		if err != nil {
			return fmt.Errorf("getting leaderboard entries: %w", err)
		}

		for i, entry := range entries {
			result := &model.WeeklyChallengeResult{
				ChallengeID: challengeID,
				UserID:      entry.UserID,
				FinalRank:   i + 1,
				FinalGems:   entry.WeeklyGems,
				DisplayName: entry.DisplayName,
			}
			if err := s.repo.InsertChallengeResult(ctx, result); err != nil {
				return fmt.Errorf("inserting challenge result: %w", err)
			}
		}
		resp.ResultsArchived = len(entries)

		// Step 3: Distribute rewards (collect pending emails for after commit)
		distributed, emails, err := s.distributeRewards(ctx, challengeID, entries)
		if err != nil {
			return fmt.Errorf("distributing rewards: %w", err)
		}
		resp.RewardsDistributed = distributed
		pendingEmails = emails

		// Step 4: Create next challenge
		now := s.now()
		nextStart, nextEnd := nextWeekBounds(now)
		newChallengeID, err := s.repo.InsertWeeklyChallenge(ctx, nextStart, nextEnd, "active")
		if err != nil {
			return fmt.Errorf("creating next challenge: %w", err)
		}
		resp.NewChallengeID = newChallengeID
		resp.Status = "completed"

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Send emails outside the transaction
	s.sendPendingEmails(ctx, pendingEmails)

	return &resp, nil
}

// sendPendingEmails sends emails collected during the transaction and updates distribution status.
func (s *Service) sendPendingEmails(ctx context.Context, emails []pendingEmail) {
	for _, pe := range emails {
		emailErr := s.email.SendEmail(pe.to, pe.subject, pe.body)
		if emailErr == nil {
			if err := s.repo.UpdateDistributionDelivered(ctx, pe.distributionID); err != nil {
				log.Printf("failed to mark distribution %s as delivered: %v", pe.distributionID, err)
			}
		} else {
			log.Printf("email send failed for distribution %s: %v", pe.distributionID, emailErr)
			if err := s.repo.UpdateDistributionFailed(ctx, pe.distributionID); err != nil {
				log.Printf("failed to mark distribution %s as failed: %v", pe.distributionID, err)
			}
		}
	}
}

// distributeRewards handles Step 3 of the weekly reset.
// Non-gem rewards are inserted as 'pending' and returned as pendingEmails for post-commit sending.
func (s *Service) distributeRewards(ctx context.Context, challengeID string, entries []model.LeaderboardEntry) (int, []pendingEmail, error) {
	campaign, err := s.repo.GetCampaignByChallenge(ctx, challengeID)
	if err != nil {
		return 0, nil, fmt.Errorf("getting campaign: %w", err)
	}
	if campaign == nil {
		return 0, nil, nil
	}

	var rules []model.RewardRule
	if err := json.Unmarshal(campaign.Rules, &rules); err != nil {
		return 0, nil, fmt.Errorf("parsing campaign rules: %w", err)
	}

	// Get campaign full info for email templates
	campaignFull, err := s.repo.GetCampaignByID(ctx, campaign.ID)
	if err != nil {
		return 0, nil, fmt.Errorf("getting campaign details: %w", err)
	}

	// Build rank→user map from entries (already rank-ordered)
	type rankedUser struct {
		userID      string
		rank        int
		displayName string
	}
	rankedUsers := make([]rankedUser, len(entries))
	for i, e := range entries {
		rankedUsers[i] = rankedUser{userID: e.UserID, rank: i + 1, displayName: e.DisplayName}
	}

	distributed := 0
	var emails []pendingEmail
	for _, rule := range rules {
		// Get users in rank range
		var usersInRange []rankedUser
		for _, ru := range rankedUsers {
			if ru.rank >= rule.RankFrom && ru.rank <= rule.RankTo {
				usersInRange = append(usersInRange, ru)
			}
		}

		for _, user := range usersInRange {
			for _, rewardTypeID := range rule.RewardTypeIDs {
				rt, err := s.repo.GetRewardTypeByID(ctx, rewardTypeID)
				if err != nil {
					return distributed, emails, fmt.Errorf("getting reward type %s: %w", rewardTypeID, err)
				}

				if rt.Type == "gems" {
					if err := s.distributeGemReward(ctx, campaign.ID, user.userID, rt); err != nil {
						return distributed, emails, err
					}
				} else {
					pe, err := s.prepareNonGemReward(ctx, campaign.ID, user.userID, user.rank, rt, campaignFull)
					if err != nil {
						return distributed, emails, err
					}
					emails = append(emails, *pe)
				}

				if err := s.repo.DecrementRewardTypeStock(ctx, rewardTypeID); err != nil {
					return distributed, emails, fmt.Errorf("decrementing stock: %w", err)
				}

				distributed++
			}
		}
	}

	// Mark campaign completed
	if err := s.repo.UpdateCampaignStatus(ctx, campaign.ID, "completed"); err != nil {
		return distributed, emails, fmt.Errorf("updating campaign status: %w", err)
	}

	return distributed, emails, nil
}

func (s *Service) distributeGemReward(ctx context.Context, campaignID, userID string, rt *model.RewardType) error {
	if err := s.repo.InsertGemHistory(ctx, userID, "reward", int(rt.Value), nil); err != nil {
		return fmt.Errorf("inserting gem reward: %w", err)
	}

	now := s.now()
	dist := &model.RewardDistribution{
		CampaignID:   campaignID,
		UserID:       userID,
		RewardTypeID: rt.ID,
		Status:       "delivered",
		DeliveredAt:  &now,
	}
	if _, err := s.repo.InsertRewardDistribution(ctx, dist); err != nil {
		return fmt.Errorf("inserting gem distribution: %w", err)
	}
	return nil
}

// prepareNonGemReward inserts a 'pending' distribution and returns a pendingEmail for post-commit sending.
func (s *Service) prepareNonGemReward(ctx context.Context, campaignID, userID string, rank int, rt *model.RewardType, campaign *model.RewardCampaignFull) (*pendingEmail, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user for email: %w", err)
	}

	subject := replacePlaceholders(campaign.NonGemClaimEmailSubject, user.Username, rank, rt)
	body := replacePlaceholders(campaign.NonGemClaimEmailBody, user.Username, rank, rt)

	dist := &model.RewardDistribution{
		CampaignID:   campaignID,
		UserID:       userID,
		RewardTypeID: rt.ID,
		Status:       "pending",
	}

	distID, err := s.repo.InsertRewardDistribution(ctx, dist)
	if err != nil {
		return nil, fmt.Errorf("inserting pending distribution: %w", err)
	}

	return &pendingEmail{
		distributionID: distID,
		to:             user.Email,
		subject:        subject,
		body:           body,
	}, nil
}

func replacePlaceholders(template, username string, rank int, rt *model.RewardType) string {
	s := template
	s = strings.ReplaceAll(s, "{{username}}", username)
	s = strings.ReplaceAll(s, "{{rank}}", fmt.Sprintf("%d", rank))
	s = strings.ReplaceAll(s, "{{reward_type}}", rt.Name)
	s = strings.ReplaceAll(s, "{{reward_value}}", fmt.Sprintf("%g", rt.Value))
	s = strings.ReplaceAll(s, "{{reward_image}}", buildImageTag(rt.Image))

	return s
}

// nextWeekBounds calculates the next Monday 00:00 UTC and Sunday 23:59:59 UTC.
func nextWeekBounds(now time.Time) (time.Time, time.Time) {
	// Find next Monday
	daysUntilMonday := (8 - int(now.Weekday())) % 7
	if daysUntilMonday == 0 {
		daysUntilMonday = 7
	}
	nextMonday := time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 0, 0, 0, 0, time.UTC)
	nextSunday := time.Date(nextMonday.Year(), nextMonday.Month(), nextMonday.Day()+6, 23, 59, 59, 0, time.UTC)
	return nextMonday, nextSunday
}

// RetryFailedEmails retries sending emails for failed distributions.
func (s *Service) RetryFailedEmails(ctx context.Context) error {
	failed, err := s.repo.GetFailedDistributions(ctx)
	if err != nil {
		return fmt.Errorf("getting failed distributions: %w", err)
	}

	for _, dist := range failed {
		subject := replacePlaceholdersFromDist(dist.EmailSubject, dist)
		body := replacePlaceholdersFromDist(dist.EmailBody, dist)

		emailErr := s.email.SendEmail(dist.Email, subject, body)
		if emailErr == nil {
			if err := s.repo.UpdateDistributionDelivered(ctx, dist.ID); err != nil {
				log.Printf("failed to mark distribution %s as delivered: %v", dist.ID, err)
			}
		} else {
			retryCount, err := s.repo.IncrementDistributionRetryCount(ctx, dist.ID)
			if err != nil {
				log.Printf("failed to increment retry count for distribution %s: %v", dist.ID, err)
				continue
			}
			if retryCount >= 3 {
				log.Printf("[ALERT] Final email failure for distribution ID: %s", dist.ID)
			}
		}
	}
	return nil
}

func replacePlaceholdersFromDist(template string, dist model.FailedDistribution) string {
	s := template
	s = strings.ReplaceAll(s, "{{username}}", dist.Username)
	s = strings.ReplaceAll(s, "{{rank}}", fmt.Sprintf("%d", dist.FinalRank))
	s = strings.ReplaceAll(s, "{{reward_type}}", dist.RewardTypeName)
	s = strings.ReplaceAll(s, "{{reward_value}}", fmt.Sprintf("%g", dist.RewardTypeValue))
	s = strings.ReplaceAll(s, "{{reward_image}}", buildImageTag(dist.RewardTypeImage))

	return s
}

func buildImageTag(imageURL *string) string {
	if imageURL == nil {
		return ""
	}
	return fmt.Sprintf("<img src='%s' width='100'>", html.EscapeString(*imageURL))
}
