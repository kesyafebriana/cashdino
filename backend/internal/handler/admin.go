package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/kesyafebriana/cashdino/backend/internal/model"
)

func (h *Handler) ListCampaigns(c echo.Context) error {
	items, err := h.svc.ListCampaigns(c.Request().Context())
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *Handler) GetCampaign(c echo.Context) error {
	id := c.Param("id")
	campaign, err := h.svc.GetCampaign(c.Request().Context(), id)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, campaign)
}

func (h *Handler) CreateCampaign(c echo.Context) error {
	var req model.CreateCampaignRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	campaign, err := h.svc.CreateCampaign(c.Request().Context(), req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusCreated, campaign)
}

func (h *Handler) UpdateCampaign(c echo.Context) error {
	id := c.Param("id")
	var req model.CreateCampaignRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	campaign, err := h.svc.UpdateCampaign(c.Request().Context(), id, req)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, campaign)
}

func (h *Handler) GetDistributions(c echo.Context) error {
	id := c.Param("id")
	rows, err := h.svc.GetDistributions(c.Request().Context(), id)
	if err != nil {
		return mapServiceError(c, err)
	}
	return c.JSON(http.StatusOK, rows)
}
