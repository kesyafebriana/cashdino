package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/kesyafebriana/cashdino/backend/internal/model"
)

// --- Mock Service ---

type mockService struct {
	earnGems func(ctx context.Context, req model.EarnGemsRequest) (*model.EarnGemsResponse, error)
	checkin  func(ctx context.Context, req model.CheckinRequest) (*model.CheckinResponse, error)
}

func (m *mockService) EarnGems(ctx context.Context, req model.EarnGemsRequest) (*model.EarnGemsResponse, error) {
	return m.earnGems(ctx, req)
}

func (m *mockService) Checkin(ctx context.Context, req model.CheckinRequest) (*model.CheckinResponse, error) {
	return m.checkin(ctx, req)
}

func newTestHandler(svc *mockService) (*Handler, *echo.Echo) {
	e := echo.New()
	h := New(svc)
	return h, e
}

// =====================================================================
// EarnGems handler tests
// =====================================================================

func TestEarnGemsHandler_ValidRequest_Returns200(t *testing.T) {
	svc := &mockService{
		earnGems: func(_ context.Context, req model.EarnGemsRequest) (*model.EarnGemsResponse, error) {
			return &model.EarnGemsResponse{UserID: req.UserID, WeeklyGems: 1500}, nil
		},
	}
	h, e := newTestHandler(svc)

	body := `{"user_id":"user-1","source":"gameplay","amount":500,"game_name":"Candy Crush"}`
	req := httptest.NewRequest(http.MethodPost, "/api/gems/earn", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.EarnGems(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp model.EarnGemsResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "user-1", resp.UserID)
	assert.Equal(t, 1500, resp.WeeklyGems)
}

func TestEarnGemsHandler_InvalidJSON_Returns400(t *testing.T) {
	svc := &mockService{}
	h, e := newTestHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/gems/earn", strings.NewReader("{invalid"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.EarnGems(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid request body")
}

func TestEarnGemsHandler_ValidationError_Returns400(t *testing.T) {
	svc := &mockService{
		earnGems: func(_ context.Context, _ model.EarnGemsRequest) (*model.EarnGemsResponse, error) {
			return nil, model.ValidationErr("invalid source: must be one of gameplay, survey, referral, boost")
		},
	}
	h, e := newTestHandler(svc)

	body := `{"user_id":"user-1","source":"payout","amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/gems/earn", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.EarnGems(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid source")
}

func TestEarnGemsHandler_NotFoundError_Returns404(t *testing.T) {
	svc := &mockService{
		earnGems: func(_ context.Context, _ model.EarnGemsRequest) (*model.EarnGemsResponse, error) {
			return nil, fmt.Errorf("validating user: %w", model.ErrNotFound)
		},
	}
	h, e := newTestHandler(svc)

	body := `{"user_id":"nonexistent","source":"gameplay","amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/gems/earn", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.EarnGems(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "validating user")
}

func TestEarnGemsHandler_InternalError_Returns500(t *testing.T) {
	svc := &mockService{
		earnGems: func(_ context.Context, _ model.EarnGemsRequest) (*model.EarnGemsResponse, error) {
			return nil, fmt.Errorf("recording gem history: db connection lost")
		},
	}
	h, e := newTestHandler(svc)

	body := `{"user_id":"user-1","source":"gameplay","amount":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/gems/earn", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.EarnGems(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "internal server error")
	assert.NotContains(t, rec.Body.String(), "db connection lost")
}

// =====================================================================
// Checkin handler tests
// =====================================================================

func TestCheckinHandler_ValidRequest_Returns200(t *testing.T) {
	svc := &mockService{
		checkin: func(_ context.Context, _ model.CheckinRequest) (*model.CheckinResponse, error) {
			return &model.CheckinResponse{GemsEarned: 150, CurrentStreak: 3, WeeklyGems: 2000}, nil
		},
	}
	h, e := newTestHandler(svc)

	body := `{"user_id":"user-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/checkin", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Checkin(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp model.CheckinResponse
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 150, resp.GemsEarned)
	assert.Equal(t, 3, resp.CurrentStreak)
	assert.Equal(t, 2000, resp.WeeklyGems)
}

func TestCheckinHandler_InvalidJSON_Returns400(t *testing.T) {
	svc := &mockService{}
	h, e := newTestHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/checkin", strings.NewReader("{bad"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Checkin(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid request body")
}

func TestCheckinHandler_AlreadyCheckedIn_Returns400(t *testing.T) {
	svc := &mockService{
		checkin: func(_ context.Context, _ model.CheckinRequest) (*model.CheckinResponse, error) {
			return nil, model.ValidationErr("already checked in today")
		},
	}
	h, e := newTestHandler(svc)

	body := `{"user_id":"user-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/checkin", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Checkin(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "already checked in today")
}

func TestCheckinHandler_NoCheckinAvailable_Returns400(t *testing.T) {
	svc := &mockService{
		checkin: func(_ context.Context, _ model.CheckinRequest) (*model.CheckinResponse, error) {
			return nil, model.ValidationErr("no check-in available today")
		},
	}
	h, e := newTestHandler(svc)

	body := `{"user_id":"user-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/checkin", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Checkin(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "no check-in available today")
}

func TestCheckinHandler_NotFoundError_Returns404(t *testing.T) {
	svc := &mockService{
		checkin: func(_ context.Context, _ model.CheckinRequest) (*model.CheckinResponse, error) {
			return nil, fmt.Errorf("validating user: %w", model.ErrNotFound)
		},
	}
	h, e := newTestHandler(svc)

	body := `{"user_id":"nonexistent"}`
	req := httptest.NewRequest(http.MethodPost, "/api/checkin", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Checkin(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "validating user")
}

func TestCheckinHandler_InternalError_Returns500(t *testing.T) {
	svc := &mockService{
		checkin: func(_ context.Context, _ model.CheckinRequest) (*model.CheckinResponse, error) {
			return nil, fmt.Errorf("recording gem history: db connection lost")
		},
	}
	h, e := newTestHandler(svc)

	body := `{"user_id":"user-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/checkin", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Checkin(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "internal server error")
	assert.NotContains(t, rec.Body.String(), "db connection lost")
}
