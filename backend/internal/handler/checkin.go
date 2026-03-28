package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/kesyafebriana/cashdino/backend/internal/model"
)

func (h *Handler) Checkin(c echo.Context) error {
	var req model.CheckinRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	resp, err := h.svc.Checkin(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, resp)
}
