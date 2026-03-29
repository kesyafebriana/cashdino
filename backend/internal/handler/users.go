package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *Handler) ListUsers(c echo.Context) error {
	// If usernames param is provided, fetch those specific users
	if names := c.QueryParam("usernames"); names != "" {
		usernames := strings.Split(names, ",")
		users, err := h.svc.ListUsers(c.Request().Context(), 0, usernames)
		if err != nil {
			c.Logger().Errorf("ListUsers error: %v", err)
			return mapServiceError(c, err)
		}
		return c.JSON(http.StatusOK, users)
	}

	limit := 3
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	users, err := h.svc.ListUsers(c.Request().Context(), limit, nil)
	if err != nil {
		c.Logger().Errorf("ListUsers error: %v", err)
		return mapServiceError(c, err)
	}

	return c.JSON(http.StatusOK, users)
}
