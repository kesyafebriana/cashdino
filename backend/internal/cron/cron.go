package cron

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/kesyafebriana/cashdino/backend/internal/service"
)

// Start configures and starts the cron scheduler with weekly reset and email retry jobs.
// Returns the cron instance so the caller can stop it on shutdown.
func Start(svc *service.Service) *cron.Cron {
	c := cron.New(cron.WithLocation(time.UTC))

	// Weekly reset: Sunday 23:59 UTC
	_, err := c.AddFunc("59 23 * * 0", func() {
		log.Println("[CRON] starting weekly reset")
		resp, err := svc.WeeklyReset(context.Background())
		if err != nil {
			log.Printf("[CRON] weekly reset failed: %v", err)
			return
		}
		log.Printf("[CRON] weekly reset completed: archived=%d, distributed=%d, new_challenge=%s",
			resp.ResultsArchived, resp.RewardsDistributed, resp.NewChallengeID)
	})
	if err != nil {
		log.Fatalf("failed to schedule weekly reset cron: %v", err)
	}

	// Email retry: every 8 hours
	_, err = c.AddFunc("0 */8 * * *", func() {
		log.Println("[CRON] starting email retry")
		if err := svc.RetryFailedEmails(context.Background()); err != nil {
			log.Printf("[CRON] email retry failed: %v", err)
			return
		}
		log.Println("[CRON] email retry completed")
	})
	if err != nil {
		log.Fatalf("failed to schedule email retry cron: %v", err)
	}

	c.Start()
	log.Println("cron scheduler started")
	return c
}
