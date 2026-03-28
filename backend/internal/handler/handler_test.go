package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kesyafebriana/cashdino/backend/internal/model"
)

// --- Mock Service ---

type mockService struct {
	earnGems             func(ctx context.Context, req model.EarnGemsRequest) (*model.EarnGemsResponse, error)
	checkin              func(ctx context.Context, req model.CheckinRequest) (*model.CheckinResponse, error)
	getBanner            func(ctx context.Context, userID string) (*model.BannerResponse, error)
	getCurrentLeaderboard func(ctx context.Context, userID string) (*model.CurrentLeaderboardResponse, error)
	getLastWeekLeaderboard func(ctx context.Context, userID string) (*model.LastWeekResponse, error)
}

func (m *mockService) EarnGems(ctx context.Context, req model.EarnGemsRequest) (*model.EarnGemsResponse, error) {
	return m.earnGems(ctx, req)
}
func (m *mockService) Checkin(ctx context.Context, req model.CheckinRequest) (*model.CheckinResponse, error) {
	return m.checkin(ctx, req)
}
func (m *mockService) GetBanner(ctx context.Context, userID string) (*model.BannerResponse, error) {
	return m.getBanner(ctx, userID)
}
func (m *mockService) GetCurrentLeaderboard(ctx context.Context, userID string) (*model.CurrentLeaderboardResponse, error) {
	return m.getCurrentLeaderboard(ctx, userID)
}
func (m *mockService) GetLastWeekLeaderboard(ctx context.Context, userID string) (*model.LastWeekResponse, error) {
	return m.getLastWeekLeaderboard(ctx, userID)
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

// =====================================================================
// Banner handler tests
// =====================================================================

func TestBannerHandler_ValidRequest_Returns200(t *testing.T) {
	gap := 42
	svc := &mockService{
		getBanner: func(_ context.Context, _ string) (*model.BannerResponse, error) {
			return &model.BannerResponse{
				ChallengeID: "c-1", EndTime: time.Date(2026, 3, 29, 23, 59, 59, 0, time.UTC),
				WeeklyGems: 4320, RankDisplay: "#12", GapToNext: &gap, DisplayName: "ja****s",
			}, nil
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/challenge/banner?user_id=user-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetBanner(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp model.BannerResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "c-1", resp.ChallengeID)
	assert.Equal(t, "#12", resp.RankDisplay)
	assert.Equal(t, 4320, resp.WeeklyGems)
	require.NotNil(t, resp.GapToNext)
	assert.Equal(t, 42, *resp.GapToNext)
}

func TestBannerHandler_NoActiveChallenge_ReturnsNoChallenge(t *testing.T) {
	svc := &mockService{
		getBanner: func(_ context.Context, _ string) (*model.BannerResponse, error) { return nil, nil },
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/challenge/banner?user_id=user-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetBanner(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "no_active_challenge")
}

func TestBannerHandler_MissingUserID_Returns400(t *testing.T) {
	svc := &mockService{
		getBanner: func(_ context.Context, _ string) (*model.BannerResponse, error) {
			return nil, model.ValidationErr("user_id is required")
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/challenge/banner", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetBanner(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// =====================================================================
// Current leaderboard handler tests
// =====================================================================

func TestCurrentLeaderboardHandler_ValidRequest_Returns200(t *testing.T) {
	rank := 1
	svc := &mockService{
		getCurrentLeaderboard: func(_ context.Context, _ string) (*model.CurrentLeaderboardResponse, error) {
			return &model.CurrentLeaderboardResponse{
				Challenge: model.ChallengeInfo{ID: "c-1", Status: "active"},
				Leaderboard: []model.CurrentLeaderboardRow{
					{Rank: 1, DisplayName: "ke****m", WeeklyGems: 12450},
				},
				CurrentUser: &model.CurrentUserInfo{Rank: &rank, RankDisplay: "1", WeeklyGems: 12450},
			}, nil
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard/current?user_id=user-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetCurrentLeaderboard(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp model.CurrentLeaderboardResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "c-1", resp.Challenge.ID)
	assert.Len(t, resp.Leaderboard, 1)
}

func TestCurrentLeaderboardHandler_NoChallenge_Returns404(t *testing.T) {
	svc := &mockService{
		getCurrentLeaderboard: func(_ context.Context, _ string) (*model.CurrentLeaderboardResponse, error) {
			return nil, fmt.Errorf("getting active challenge: %w", model.ErrNotFound)
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard/current?user_id=user-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetCurrentLeaderboard(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// =====================================================================
// Last week leaderboard handler tests
// =====================================================================

func TestLastWeekHandler_ValidRequest_Returns200(t *testing.T) {
	svc := &mockService{
		getLastWeekLeaderboard: func(_ context.Context, _ string) (*model.LastWeekResponse, error) {
			return &model.LastWeekResponse{
				Challenge: &model.LastWeekChallengeInfo{ID: "c-0"},
				Leaderboard: []model.LastWeekRow{
					{Rank: 1, DisplayName: "ke****m", FinalGems: 12450, Rewards: []model.RewardInfo{{Name: "10K Gems", Type: "gems", Value: 10000}}},
				},
			}, nil
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard/last-week?user_id=user-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetLastWeekLeaderboard(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestLastWeekHandler_NoCompletedChallenge_Returns200WithNullChallenge(t *testing.T) {
	svc := &mockService{
		getLastWeekLeaderboard: func(_ context.Context, _ string) (*model.LastWeekResponse, error) {
			return &model.LastWeekResponse{Challenge: nil}, nil
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/leaderboard/last-week?user_id=user-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetLastWeekLeaderboard(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"challenge":null`)
}
