package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) GetBanner(c echo.Context) error {
	userID := c.QueryParam("user_id")

	resp, err := h.svc.GetBanner(c.Request().Context(), userID)
	if err != nil {
		return mapServiceError(c, err)
	}

	if resp == nil {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "no_active_challenge",
			"message": "Next challenge starts Monday",
		})
	}

	return c.JSON(http.StatusOK, resp)
}
