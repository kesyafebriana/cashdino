package server

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/kesyafebriana/cashdino/backend/internal/handler"
	"github.com/kesyafebriana/cashdino/backend/internal/repository"
	"github.com/kesyafebriana/cashdino/backend/internal/service"
)

func New(pool *pgxpool.Pool) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Wire layers
	repo := repository.New(pool)
	svc := service.New(repo)
	h := handler.New(svc)

	// Routes
	e.GET("/api/health", h.Health)
	e.POST("/api/gems/earn", h.EarnGems)
	e.POST("/api/checkin", h.Checkin)

	return e
}
