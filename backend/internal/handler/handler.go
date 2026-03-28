package handler

import (
	"context"

	"github.com/kesyafebriana/cashdino/backend/internal/model"
)

type ServiceInterface interface {
	EarnGems(ctx context.Context, req model.EarnGemsRequest) (*model.EarnGemsResponse, error)
	Checkin(ctx context.Context, req model.CheckinRequest) (*model.CheckinResponse, error)
}

type Handler struct {
	svc ServiceInterface
}

func New(svc ServiceInterface) *Handler {
	return &Handler{svc: svc}
}
