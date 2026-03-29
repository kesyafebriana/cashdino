package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var usernames = []string{
	"james", "olivia", "william", "emma", "benjamin",
	"sophia", "lucas", "mia", "henry", "charlotte",
	"alexander", "amelia", "daniel", "harper", "matthew",
	"evelyn", "jackson", "abigail", "sebastian", "ella",
	"aiden", "scarlett", "owen", "grace", "samuel",
	"lily", "ryan", "chloe", "nathan", "zoey",
	"caleb", "penelope", "christian", "layla", "hunter",
	"riley", "connor", "nora", "eli", "stella",
	"aaron", "lucy", "landon", "emilia", "adrian",
	"maya", "miles", "aria", "leo", "ellie",
	// 53 additional users (total: 103)
	"ethan", "victoria", "logan", "hannah", "mason",
	"addison", "jacob", "aurora", "liam", "savannah",
	"noah", "brooklyn", "jack", "leah", "luke",
	"natalie", "gabriel", "hazel", "julian", "violet",
	"carter", "eleanor", "jayden", "claire", "dylan",
	"skylar", "isaac", "bella", "andrew", "alice",
	"thomas", "madelyn", "joshua", "audrey", "christopher",
	"paisley", "theodore", "sadie", "david", "ruby",
	"joseph", "eva", "charles", "naomi", "wyatt",
	"quinn", "max", "ivy", "finn", "lyla",
	"oscar", "freya", "felix",
}

var gameNames = []string{
	"Candy Crush", "Bubble Pop", "Word Hunt", "Puzzle Quest",
	"Gem Miner", "Tower Stack", "Lucky Spin", "Card Match",
}

type rewardRule struct {
	RankFrom      int      `json:"rank_from"`
	RankTo        int      `json:"rank_to"`
	RewardTypeIDs []string `json:"reward_type_ids"`
}

type lastWeekUser struct {
	index int
	gems  int
}

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("connected to database, running seed...")

	// Use a transaction for the entire seed
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	// -----------------------------------------------------------
	// 1. Create 103 users
	// -----------------------------------------------------------
	log.Println("creating users...")
	userIDs := make([]string, len(usernames))
	for i, name := range usernames {
		email := fmt.Sprintf("%s@example.com", name)
		err := tx.QueryRow(ctx,
			`INSERT INTO users (username, email) VALUES ($1, $2) RETURNING id`,
			name, email,
		).Scan(&userIDs[i])
		if err != nil {
			log.Fatalf("failed to create user %s: %v", name, err)
		}
	}
	log.Printf("created %d users", len(userIDs))

	// -----------------------------------------------------------
	// 2. Create daily_checkins for past 14 days
	// -----------------------------------------------------------
	log.Println("creating daily checkins...")
	multipliers := []string{"1.00", "1.25", "1.50", "1.75", "2.00"}
	checkinIDs := make([]string, 14)
	now := time.Now().UTC()
	for i := 0; i < 14; i++ {
		date := now.AddDate(0, 0, -(13 - i))
		baseGems := 50 + rand.Intn(51) // 50-100
		mult := multipliers[i%5]
		err := tx.QueryRow(ctx,
			`INSERT INTO daily_checkins (date, base_gems, streak_multiplier, is_active)
			 VALUES ($1, $2, $3, true) RETURNING id`,
			date.Format("2006-01-02"), baseGems, mult,
		).Scan(&checkinIDs[i])
		if err != nil {
			log.Fatalf("failed to create daily checkin: %v", err)
		}
	}
	log.Printf("created %d daily checkins", len(checkinIDs))

	// -----------------------------------------------------------
	// 3. Create weekly challenges (last week + this week)
	// -----------------------------------------------------------
	log.Println("creating weekly challenges...")

	// Find the Monday of this week
	today := now
	weekday := int(today.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	thisMonday := today.AddDate(0, 0, -(weekday - 1))
	thisMonday = time.Date(thisMonday.Year(), thisMonday.Month(), thisMonday.Day(), 0, 0, 0, 0, time.UTC)
	thisSunday := thisMonday.AddDate(0, 0, 6)
	thisSunday = time.Date(thisSunday.Year(), thisSunday.Month(), thisSunday.Day(), 23, 59, 59, 0, time.UTC)

	lastMonday := thisMonday.AddDate(0, 0, -7)
	lastSunday := thisMonday.AddDate(0, 0, -1)
	lastSunday = time.Date(lastSunday.Year(), lastSunday.Month(), lastSunday.Day(), 23, 59, 59, 0, time.UTC)

	var lastChallengeID string
	err = tx.QueryRow(ctx,
		`INSERT INTO weekly_challenges (start_time, end_time, status)
		 VALUES ($1, $2, 'completed') RETURNING id`,
		lastMonday, lastSunday,
	).Scan(&lastChallengeID)
	if err != nil {
		log.Fatalf("failed to create last week challenge: %v", err)
	}

	var activeChallengeID string
	err = tx.QueryRow(ctx,
		`INSERT INTO weekly_challenges (start_time, end_time, status)
		 VALUES ($1, $2, 'active') RETURNING id`,
		thisMonday, thisSunday,
	).Scan(&activeChallengeID)
	if err != nil {
		log.Fatalf("failed to create active challenge: %v", err)
	}
	log.Printf("created challenges: last=%s, active=%s", lastChallengeID, activeChallengeID)

	// -----------------------------------------------------------
	// 4. Generate masked display names
	// -----------------------------------------------------------
	displayNames := generateDisplayNames(usernames)

	// -----------------------------------------------------------
	// 5. Seed active week: gem_history + leaderboard_entries
	// -----------------------------------------------------------
	log.Println("seeding active week gem_history and leaderboard_entries...")

	totalUsers := len(userIDs)
	for i, userID := range userIDs {
		var targetGems int
		switch {
		case i < 3: // top 3: 10K-15K
			targetGems = 10000 + rand.Intn(5001)
		case i < 10: // top 4-10: 5K-10K
			targetGems = 5000 + rand.Intn(5001)
		case i < 25: // middle 11-25: 3K-5K
			targetGems = 3000 + rand.Intn(2001)
		case i < 50: // lower-middle 26-50: 500-3K
			targetGems = 500 + rand.Intn(2501)
		case i < 75: // low 51-75: 100-500
			targetGems = 100 + rand.Intn(401)
		case i < 99: // very low 76-99: 10-100
			targetGems = 10 + rand.Intn(91)
		default: // zero gems 100-103: 0 gems (no activity)
			targetGems = 0
		}

		if targetGems == 0 {
			// No gem_history or leaderboard entry for zero-gem users
			continue
		}

		// Scatter gems across multiple gem_history records
		remaining := targetGems
		var firstEarned time.Time
		recordCount := 5 + rand.Intn(16) // 5-20 records per user

		for j := 0; j < recordCount && remaining > 0; j++ {
			amount := remaining / (recordCount - j)
			if j < recordCount-1 {
				amount = amount/2 + rand.Intn(amount/2+1)
			} else {
				amount = remaining
			}
			if amount <= 0 {
				amount = 1
			}
			if amount > remaining {
				amount = remaining
			}

			source := "gameplay"
			var gameName *string
			if rand.Float64() < 0.3 {
				source = "daily_checkin"
			} else {
				g := gameNames[rand.Intn(len(gameNames))]
				gameName = &g
			}

			earnedAt := thisMonday.Add(time.Duration(rand.Intn(int(now.Sub(thisMonday).Seconds()))) * time.Second)
			if j == 0 || earnedAt.Before(firstEarned) {
				firstEarned = earnedAt
			}

			_, err := tx.Exec(ctx,
				`INSERT INTO gem_history (user_id, source, amount, game_name, created_at)
				 VALUES ($1, $2, $3, $4, $5)`,
				userID, source, amount, gameName, earnedAt,
			)
			if err != nil {
				log.Fatalf("failed to insert gem_history: %v", err)
			}
			remaining -= amount
		}

		// Create leaderboard entry
		_, err := tx.Exec(ctx,
			`INSERT INTO leaderboard_entries (challenge_id, user_id, weekly_gems, first_gem_earned_at, display_name)
			 VALUES ($1, $2, $3, $4, $5)`,
			activeChallengeID, userID, targetGems, firstEarned, displayNames[i],
		)
		if err != nil {
			log.Fatalf("failed to insert leaderboard entry: %v", err)
		}
	}
	log.Printf("active week seeded (%d users with gems, %d with 0 gems)", totalUsers-4, 4)

	// -----------------------------------------------------------
	// 6. Seed last week: gem_history + weekly_challenge_results
	// -----------------------------------------------------------
	log.Println("seeding last week results...")

	lastWeekUsers := make([]lastWeekUser, len(userIDs))
	for i := range userIDs {
		var gems int
		switch {
		case i < 3:
			gems = 11000 + rand.Intn(4001)
		case i < 10:
			gems = 6000 + rand.Intn(5001)
		case i < 25:
			gems = 2500 + rand.Intn(2501)
		case i < 50:
			gems = 400 + rand.Intn(2101)
		case i < 75:
			gems = 50 + rand.Intn(351)
		case i < 99:
			gems = 5 + rand.Intn(46)
		default:
			gems = 0
		}
		lastWeekUsers[i] = lastWeekUser{index: i, gems: gems}

		if gems == 0 {
			continue
		}

		// Insert gem_history records for last week too
		source := "gameplay"
		var gameName *string
		g := gameNames[rand.Intn(len(gameNames))]
		gameName = &g
		earnedAt := lastMonday.Add(time.Duration(rand.Intn(int(lastSunday.Sub(lastMonday).Seconds()))) * time.Second)
		_, err := tx.Exec(ctx,
			`INSERT INTO gem_history (user_id, source, amount, game_name, created_at)
			 VALUES ($1, $2, $3, $4, $5)`,
			userIDs[i], source, gems, gameName, earnedAt,
		)
		if err != nil {
			log.Fatalf("failed to insert last week gem_history: %v", err)
		}
	}

	// Sort by gems descending for ranking (only include users with gems > 0)
	sortLastWeek(lastWeekUsers)

	for rank, lwu := range lastWeekUsers {
		if lwu.gems == 0 {
			break // sorted descending, so all remaining are 0
		}
		_, err := tx.Exec(ctx,
			`INSERT INTO weekly_challenge_results (challenge_id, user_id, final_rank, final_gems, display_name)
			 VALUES ($1, $2, $3, $4, $5)`,
			lastChallengeID, userIDs[lwu.index], rank+1, lwu.gems, displayNames[lwu.index],
		)
		if err != nil {
			log.Fatalf("failed to insert weekly_challenge_result: %v", err)
		}
	}
	log.Println("last week results seeded")

	// -----------------------------------------------------------
	// 7. Create reward campaign for last week
	// -----------------------------------------------------------
	log.Println("creating reward campaign...")

	var campaignID string
	err = tx.QueryRow(ctx,
		`INSERT INTO reward_campaigns (challenge_id, name, banner_image, rules, non_gem_claim_email_subject, non_gem_claim_email_body, status)
		 VALUES ($1, $2, $3, '[]', $4, $5, 'completed') RETURNING id`,
		lastChallengeID,
		"Week 12 Rewards",
		"https://picsum.photos/600/200",
		"Congratulations! You won a reward from CashDino!",
		"Hi {{username}},\n\nCongratulations on reaching rank #{{rank}} in this week's challenge!\n\nYou've won: {{reward_type}} ({{reward_value}})\n\nClaim your reward by replying to this email.\n\nBest,\nCashDino Team",
	).Scan(&campaignID)
	if err != nil {
		log.Fatalf("failed to create reward campaign: %v", err)
	}

	// Create reward types
	rewardTypeIDs := make(map[string]string) // name -> id
	rewardDefs := []struct {
		name  string
		typ   string
		value string
		image *string
		stock int
	}{
		{"10K Gems", "gems", "10000", nil, 1},
		{"2K Gems", "gems", "2000", nil, 2},
		{"$10 Amazon Gift Card", "gift_card", "10", strPtr("https://picsum.photos/100"), 3},
	}

	for _, rd := range rewardDefs {
		var id string
		err := tx.QueryRow(ctx,
			`INSERT INTO reward_types (campaign_id, name, type, value, image, stock)
			 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			campaignID, rd.name, rd.typ, rd.value, rd.image, rd.stock,
		).Scan(&id)
		if err != nil {
			log.Fatalf("failed to create reward type %s: %v", rd.name, err)
		}
		rewardTypeIDs[rd.name] = id
	}

	// Build rules JSONB and update campaign
	rules := []rewardRule{
		{RankFrom: 1, RankTo: 1, RewardTypeIDs: []string{rewardTypeIDs["10K Gems"], rewardTypeIDs["$10 Amazon Gift Card"]}},
		{RankFrom: 2, RankTo: 3, RewardTypeIDs: []string{rewardTypeIDs["2K Gems"], rewardTypeIDs["$10 Amazon Gift Card"]}},
	}
	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		log.Fatalf("failed to marshal rules: %v", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE reward_campaigns SET rules = $1 WHERE id = $2`,
		rulesJSON, campaignID,
	)
	if err != nil {
		log.Fatalf("failed to update campaign rules: %v", err)
	}
	log.Printf("created campaign %s with %d reward types", campaignID, len(rewardDefs))

	// -----------------------------------------------------------
	// 8. Create reward distributions for top 10 from last week
	// -----------------------------------------------------------
	log.Println("creating reward distributions...")

	deliveredAt := lastSunday.Add(1 * time.Minute)
	distributionCount := 0

	for rank, lwu := range lastWeekUsers {
		r := rank + 1
		var typeIDs []string

		switch {
		case r == 1:
			typeIDs = []string{rewardTypeIDs["10K Gems"], rewardTypeIDs["$10 Amazon Gift Card"]}
		case r >= 2 && r <= 3:
			typeIDs = []string{rewardTypeIDs["2K Gems"], rewardTypeIDs["$10 Amazon Gift Card"]}
		default:
			continue
		}

		for _, rtID := range typeIDs {
			// Make rank #3's gift card distribution fail (simulate failed email)
			isFailed := r == 3 && rtID == rewardTypeIDs["$10 Amazon Gift Card"]

			if isFailed {
				_, err := tx.Exec(ctx,
					`INSERT INTO reward_distributions (campaign_id, user_id, reward_type_id, status, retry_count)
					 VALUES ($1, $2, $3, 'failed', 3)`,
					campaignID, userIDs[lwu.index], rtID,
				)
				if err != nil {
					log.Fatalf("failed to insert failed reward distribution: %v", err)
				}
			} else {
				var emailSentAt *time.Time
				if rtID == rewardTypeIDs["$10 Amazon Gift Card"] {
					t := deliveredAt
					emailSentAt = &t
				}
				_, err := tx.Exec(ctx,
					`INSERT INTO reward_distributions (campaign_id, user_id, reward_type_id, status, delivered_at, email_sent_at)
					 VALUES ($1, $2, $3, 'delivered', $4, $5)`,
					campaignID, userIDs[lwu.index], rtID, deliveredAt, emailSentAt,
				)
				if err != nil {
					log.Fatalf("failed to insert reward distribution: %v", err)
				}
			}
			distributionCount++
		}

		// For gem reward types, also insert gem_history with source='reward'
		if r == 1 {
			_, err := tx.Exec(ctx,
				`INSERT INTO gem_history (user_id, source, amount, created_at)
				 VALUES ($1, 'reward', 10000, $2)`,
				userIDs[lwu.index], deliveredAt,
			)
			if err != nil {
				log.Fatalf("failed to insert gem reward history: %v", err)
			}
		} else if r >= 2 && r <= 3 {
			_, err := tx.Exec(ctx,
				`INSERT INTO gem_history (user_id, source, amount, created_at)
				 VALUES ($1, 'reward', 2000, $2)`,
				userIDs[lwu.index], deliveredAt,
			)
			if err != nil {
				log.Fatalf("failed to insert gem reward history: %v", err)
			}
		}
	}
	log.Printf("created %d reward distributions", distributionCount)

	// -----------------------------------------------------------
	// 9. Create reward campaign for this week (active challenge)
	// -----------------------------------------------------------
	log.Println("creating this week's reward campaign...")

	var thisWeekCampaignID string
	err = tx.QueryRow(ctx,
		`INSERT INTO reward_campaigns (challenge_id, name, banner_image, rules, non_gem_claim_email_subject, non_gem_claim_email_body, status)
		 VALUES ($1, $2, $3, '[]', $4, $5, 'active') RETURNING id`,
		activeChallengeID,
		"Week 13 Rewards",
		"https://picsum.photos/600/200",
		"Congratulations! You won a reward from CashDino!",
		"Hi {{username}},\n\nCongratulations on reaching rank #{{rank}} in this week's challenge!\n\nYou've won: {{reward_type}} ({{reward_value}})\n\nClaim your reward by replying to this email.\n\nBest,\nCashDino Team",
	).Scan(&thisWeekCampaignID)
	if err != nil {
		log.Fatalf("failed to create this week reward campaign: %v", err)
	}

	thisWeekRewardTypeIDs := make(map[string]string)
	thisWeekRewardDefs := []struct {
		name  string
		typ   string
		value string
		image *string
		stock int
	}{
		{"10K Gems", "gems", "10000", nil, 1},
		{"2K Gems", "gems", "2000", nil, 2},
		{"$10 Amazon Gift Card", "gift_card", "10", strPtr("https://picsum.photos/100"), 3},
	}

	for _, rd := range thisWeekRewardDefs {
		var id string
		err := tx.QueryRow(ctx,
			`INSERT INTO reward_types (campaign_id, name, type, value, image, stock)
			 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			thisWeekCampaignID, rd.name, rd.typ, rd.value, rd.image, rd.stock,
		).Scan(&id)
		if err != nil {
			log.Fatalf("failed to create this week reward type %s: %v", rd.name, err)
		}
		thisWeekRewardTypeIDs[rd.name] = id
	}

	thisWeekRules := []rewardRule{
		{RankFrom: 1, RankTo: 1, RewardTypeIDs: []string{thisWeekRewardTypeIDs["10K Gems"], thisWeekRewardTypeIDs["$10 Amazon Gift Card"]}},
		{RankFrom: 2, RankTo: 3, RewardTypeIDs: []string{thisWeekRewardTypeIDs["2K Gems"], thisWeekRewardTypeIDs["$10 Amazon Gift Card"]}},
	}
	thisWeekRulesJSON, err := json.Marshal(thisWeekRules)
	if err != nil {
		log.Fatalf("failed to marshal this week rules: %v", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE reward_campaigns SET rules = $1 WHERE id = $2`,
		thisWeekRulesJSON, thisWeekCampaignID,
	)
	if err != nil {
		log.Fatalf("failed to update this week campaign rules: %v", err)
	}
	log.Printf("created this week campaign %s with %d reward types", thisWeekCampaignID, len(thisWeekRewardDefs))

	// -----------------------------------------------------------
	// 10. Create some user_daily_checkins for active week
	// -----------------------------------------------------------
	log.Println("creating user daily checkin records...")
	checkinCount := 0
	// Use only checkins from the last 7 days (this week's daily checkins)
	recentCheckins := checkinIDs[7:] // last 7 of 14
	for i, userID := range userIDs {
		// Top users check in more consistently, zero-gem users don't check in
		if i >= 99 {
			continue // zero-gem users skip check-ins
		}
		maxCheckins := 2
		if i < 10 {
			maxCheckins = 7
		} else if i < 25 {
			maxCheckins = 5
		} else if i < 50 {
			maxCheckins = 3
		}
		streak := 0
		for j := 0; j < len(recentCheckins) && j < maxCheckins; j++ {
			streak++
			gemsEarned := 50 + rand.Intn(51) // rough approximation
			_, err := tx.Exec(ctx,
				`INSERT INTO user_daily_checkins (user_id, checkin_id, gems_earned, current_streak)
				 VALUES ($1, $2, $3, $4)`,
				userID, recentCheckins[j], gemsEarned, streak,
			)
			if err != nil {
				// Skip duplicate constraint violations silently
				continue
			}
			checkinCount++
		}
	}
	log.Printf("created %d user daily checkin records", checkinCount)

	// -----------------------------------------------------------
	// Commit
	// -----------------------------------------------------------
	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("failed to commit transaction: %v", err)
	}

	log.Println("seed completed successfully!")
}

// generateDisplayNames creates masked display names per the PRD:
// first 2 chars + "****" + last char. If <=4 chars: first char + "****".
// On collision, append a random digit.
func generateDisplayNames(names []string) []string {
	seen := make(map[string]bool)
	result := make([]string, len(names))

	for i, name := range names {
		masked := maskName(name)
		for seen[masked] {
			masked = masked + fmt.Sprintf("%d", rand.Intn(10))
		}
		seen[masked] = true
		result[i] = masked
	}
	return result
}

func maskName(name string) string {
	runes := []rune(name)
	if len(runes) <= 4 {
		return string(runes[0:1]) + "****"
	}
	return string(runes[0:2]) + "****" + string(runes[len(runes)-1:])
}

// sortLastWeek sorts by gems descending (simple insertion sort)
func sortLastWeek(users []lastWeekUser) {
	for i := 1; i < len(users); i++ {
		key := users[i]
		j := i - 1
		for j >= 0 && users[j].gems < key.gems {
			users[j+1] = users[j]
			j--
		}
		users[j+1] = key
	}
}

func strPtr(s string) *string {
	return &s
}
