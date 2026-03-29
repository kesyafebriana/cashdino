package server

import (
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/kesyafebriana/cashdino/backend/internal/handler"
	"github.com/kesyafebriana/cashdino/backend/internal/repository"
	"github.com/kesyafebriana/cashdino/backend/internal/service"
)

// NewService creates a wired service with email support for use by both server and cron.
func NewService(pool *pgxpool.Pool) *service.Service {
	repo := repository.New(pool)
	email := newEmailService()
	return service.New(repo, email)
}

func newEmailService() *service.EmailService {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	return service.NewEmailService(host, port, user, pass)
}

func New(svc *service.Service) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Wire layers
	h := handler.New(svc)

	// Routes
	e.GET("/api/health", h.Health)
	e.GET("/api/users", h.ListUsers)
	e.POST("/api/gems/earn", h.EarnGems)
	e.POST("/api/checkin", h.Checkin)
	e.GET("/api/challenge/banner", h.GetBanner)
	e.GET("/api/leaderboard/current", h.GetCurrentLeaderboard)
	e.GET("/api/leaderboard/last-week", h.GetLastWeekLeaderboard)

	// Admin routes
	e.GET("/api/admin/campaigns", h.ListCampaigns)
	e.GET("/api/admin/campaigns/:id", h.GetCampaign)
	e.POST("/api/admin/campaigns", h.CreateCampaign)
	e.PUT("/api/admin/campaigns/:id", h.UpdateCampaign)
	e.GET("/api/admin/campaigns/:id/distributions", h.GetDistributions)
	e.POST("/api/admin/reset-week", h.ResetWeek)

	return e
}
