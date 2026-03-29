package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/kesyafebriana/cashdino/backend/internal/model"
)

type ServiceInterface interface {
	ListUsers(ctx context.Context, limit int, usernames []string) ([]model.User, error)
	EarnGems(ctx context.Context, req model.EarnGemsRequest) (*model.EarnGemsResponse, error)
	Checkin(ctx context.Context, req model.CheckinRequest) (*model.CheckinResponse, error)
	GetBanner(ctx context.Context, userID string) (*model.BannerResponse, error)
	GetCurrentLeaderboard(ctx context.Context, userID string) (*model.CurrentLeaderboardResponse, error)
	GetLastWeekLeaderboard(ctx context.Context, userID string) (*model.LastWeekResponse, error)
	ListChallenges(ctx context.Context) ([]model.WeeklyChallenge, error)
	ListCampaigns(ctx context.Context) ([]model.AdminCampaignListItem, error)
	DeleteCampaign(ctx context.Context, id string) error
	MarkDistributionDelivered(ctx context.Context, id string) error
	RetrySingleDistribution(ctx context.Context, id string) error
	GetCampaign(ctx context.Context, id string) (*model.AdminCampaignDetail, error)
	CreateCampaign(ctx context.Context, req model.CreateCampaignRequest) (*model.AdminCampaignDetail, error)
	UpdateCampaign(ctx context.Context, id string, req model.CreateCampaignRequest) (*model.AdminCampaignDetail, error)
	GetDistributions(ctx context.Context, campaignID string) ([]model.AdminDistributionRow, error)
	WeeklyReset(ctx context.Context) (*model.WeeklyResetResponse, error)
}

type Handler struct {
	svc ServiceInterface
}

func New(svc ServiceInterface) *Handler {
	return &Handler{svc: svc}
}

func mapServiceError(c echo.Context, err error) error {
	if errors.Is(err, model.ErrNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	if errors.Is(err, model.ErrValidation) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
