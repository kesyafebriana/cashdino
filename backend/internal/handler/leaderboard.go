package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetCurrentLeaderboard(c echo.Context) error {
	userID := c.QueryParam("user_id")

	resp, err := h.svc.GetCurrentLeaderboard(c.Request().Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetLastWeekLeaderboard(c echo.Context) error {
	userID := c.QueryParam("user_id")

	resp, err := h.svc.GetLastWeekLeaderboard(c.Request().Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, resp)
}
