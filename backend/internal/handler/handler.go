package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
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

func mapServiceError(c echo.Context, err error) error {
	if errors.Is(err, model.ErrNotFound) {
		return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
	}
	if errors.Is(err, model.ErrValidation) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
