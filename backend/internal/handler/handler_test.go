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
	listUsers            func(ctx context.Context, limit int, usernames []string) ([]model.User, error)
	earnGems             func(ctx context.Context, req model.EarnGemsRequest) (*model.EarnGemsResponse, error)
	checkin              func(ctx context.Context, req model.CheckinRequest) (*model.CheckinResponse, error)
	getBanner            func(ctx context.Context, userID string) (*model.BannerResponse, error)
	getCurrentLeaderboard func(ctx context.Context, userID string) (*model.CurrentLeaderboardResponse, error)
	getLastWeekLeaderboard func(ctx context.Context, userID string) (*model.LastWeekResponse, error)
	listCampaigns        func(ctx context.Context) ([]model.AdminCampaignListItem, error)
	getCampaign          func(ctx context.Context, id string) (*model.AdminCampaignDetail, error)
	createCampaign       func(ctx context.Context, req model.CreateCampaignRequest) (*model.AdminCampaignDetail, error)
	updateCampaign       func(ctx context.Context, id string, req model.CreateCampaignRequest) (*model.AdminCampaignDetail, error)
	getDistributions     func(ctx context.Context, campaignID string) ([]model.AdminDistributionRow, error)
	weeklyReset          func(ctx context.Context) (*model.WeeklyResetResponse, error)
}

func (m *mockService) ListUsers(ctx context.Context, limit int, usernames []string) ([]model.User, error) {
	return m.listUsers(ctx, limit, usernames)
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
func (m *mockService) ListChallenges(ctx context.Context) ([]model.WeeklyChallenge, error) {
	return nil, nil
}
func (m *mockService) DeleteCampaign(ctx context.Context, id string) error {
	return nil
}
func (m *mockService) MarkDistributionDelivered(ctx context.Context, id string) error {
	return nil
}
func (m *mockService) RetrySingleDistribution(ctx context.Context, id string) error {
	return nil
}
func (m *mockService) ListCampaigns(ctx context.Context) ([]model.AdminCampaignListItem, error) {
	return m.listCampaigns(ctx)
}
func (m *mockService) GetCampaign(ctx context.Context, id string) (*model.AdminCampaignDetail, error) {
	return m.getCampaign(ctx, id)
}
func (m *mockService) CreateCampaign(ctx context.Context, req model.CreateCampaignRequest) (*model.AdminCampaignDetail, error) {
	return m.createCampaign(ctx, req)
}
func (m *mockService) UpdateCampaign(ctx context.Context, id string, req model.CreateCampaignRequest) (*model.AdminCampaignDetail, error) {
	return m.updateCampaign(ctx, id, req)
}
func (m *mockService) GetDistributions(ctx context.Context, campaignID string) ([]model.AdminDistributionRow, error) {
	return m.getDistributions(ctx, campaignID)
}
func (m *mockService) WeeklyReset(ctx context.Context) (*model.WeeklyResetResponse, error) {
	return m.weeklyReset(ctx)
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
				TotalGems: 3886, WeeklyGems: 4320, RankDisplay: "#12", GapToNext: &gap, DisplayName: "ja****s",
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

// =====================================================================
// ListCampaigns handler tests
// =====================================================================

func TestListCampaignsHandler_Returns200(t *testing.T) {
	svc := &mockService{
		listCampaigns: func(_ context.Context) ([]model.AdminCampaignListItem, error) {
			return []model.AdminCampaignListItem{
				{ID: "camp-1", Name: "Week 14", Status: "active", RewardTypesCount: 3, TotalStock: 11},
			}, nil
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/campaigns", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListCampaigns(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "camp-1")
}

func TestListCampaignsHandler_InternalError_Returns500(t *testing.T) {
	svc := &mockService{
		listCampaigns: func(_ context.Context) ([]model.AdminCampaignListItem, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/campaigns", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ListCampaigns(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// =====================================================================
// GetCampaign handler tests
// =====================================================================

func TestGetCampaignHandler_Returns200(t *testing.T) {
	svc := &mockService{
		getCampaign: func(_ context.Context, id string) (*model.AdminCampaignDetail, error) {
			return &model.AdminCampaignDetail{
				ID: id, Name: "Week 14", Status: "active",
				RewardTypes: []model.RewardType{{ID: "rt-1", Name: "10K Gems"}},
				Rules:       []model.AdminCampaignRuleDetail{{RankFrom: 1, RankTo: 1, RewardNames: []string{"10K Gems"}}},
			}, nil
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/campaigns/camp-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("camp-1")

	err := h.GetCampaign(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "camp-1")
}

func TestGetCampaignHandler_NotFound_Returns404(t *testing.T) {
	svc := &mockService{
		getCampaign: func(_ context.Context, _ string) (*model.AdminCampaignDetail, error) {
			return nil, model.NotFoundErr("campaign not found")
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/campaigns/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.GetCampaign(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// =====================================================================
// CreateCampaign handler tests
// =====================================================================

func TestCreateCampaignHandler_ValidRequest_Returns201(t *testing.T) {
	svc := &mockService{
		createCampaign: func(_ context.Context, _ model.CreateCampaignRequest) (*model.AdminCampaignDetail, error) {
			return &model.AdminCampaignDetail{ID: "camp-new", Name: "Week 14"}, nil
		},
	}
	h, e := newTestHandler(svc)
	body := `{"challenge_id":"ch-1","name":"Week 14","banner_image":"https://img.png","reward_types":[{"name":"10K Gems","type":"gems","value":10000,"stock":1}],"rules":[{"rank_from":1,"rank_to":1,"reward_type_indexes":[0]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/campaigns", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateCampaign(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "camp-new")
}

func TestCreateCampaignHandler_InvalidJSON_Returns400(t *testing.T) {
	svc := &mockService{}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/campaigns", strings.NewReader("{bad"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateCampaign(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateCampaignHandler_ValidationError_Returns400(t *testing.T) {
	svc := &mockService{
		createCampaign: func(_ context.Context, _ model.CreateCampaignRequest) (*model.AdminCampaignDetail, error) {
			return nil, model.ValidationErr("overlapping rank ranges")
		},
	}
	h, e := newTestHandler(svc)
	body := `{"challenge_id":"ch-1","name":"Week 14","banner_image":"https://img.png","reward_types":[{"name":"10K Gems","type":"gems","value":10000,"stock":1}],"rules":[{"rank_from":1,"rank_to":1,"reward_type_indexes":[0]}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/admin/campaigns", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateCampaign(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "overlapping rank ranges")
}

// =====================================================================
// UpdateCampaign handler tests
// =====================================================================

func TestUpdateCampaignHandler_ValidRequest_Returns200(t *testing.T) {
	svc := &mockService{
		updateCampaign: func(_ context.Context, id string, _ model.CreateCampaignRequest) (*model.AdminCampaignDetail, error) {
			return &model.AdminCampaignDetail{ID: id, Name: "Week 14 Updated"}, nil
		},
	}
	h, e := newTestHandler(svc)
	body := `{"challenge_id":"ch-1","name":"Week 14 Updated","banner_image":"https://img.png","reward_types":[{"name":"10K Gems","type":"gems","value":10000,"stock":1}],"rules":[{"rank_from":1,"rank_to":1,"reward_type_indexes":[0]}]}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/campaigns/camp-1", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("camp-1")

	err := h.UpdateCampaign(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Week 14 Updated")
}

func TestUpdateCampaignHandler_NotFound_Returns404(t *testing.T) {
	svc := &mockService{
		updateCampaign: func(_ context.Context, _ string, _ model.CreateCampaignRequest) (*model.AdminCampaignDetail, error) {
			return nil, model.NotFoundErr("campaign not found")
		},
	}
	h, e := newTestHandler(svc)
	body := `{"challenge_id":"ch-1","name":"Week 14","banner_image":"https://img.png","reward_types":[{"name":"10K Gems","type":"gems","value":10000,"stock":1}],"rules":[{"rank_from":1,"rank_to":1,"reward_type_indexes":[0]}]}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/campaigns/nonexistent", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.UpdateCampaign(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateCampaignHandler_InvalidJSON_Returns400(t *testing.T) {
	svc := &mockService{}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodPut, "/api/admin/campaigns/camp-1", strings.NewReader("{bad"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("camp-1")

	err := h.UpdateCampaign(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// =====================================================================
// GetDistributions handler tests
// =====================================================================

func TestGetDistributionsHandler_Returns200(t *testing.T) {
	delivered := time.Date(2026, 3, 28, 12, 0, 0, 0, time.UTC)
	svc := &mockService{
		getDistributions: func(_ context.Context, _ string) ([]model.AdminDistributionRow, error) {
			return []model.AdminDistributionRow{
				{ID: "dist-1", UserID: "user-1", DisplayName: "ja****s", RewardName: "10K Gems", Status: "delivered", DeliveredAt: &delivered, FinalRank: 1},
			}, nil
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/campaigns/camp-1/distributions", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("camp-1")

	err := h.GetDistributions(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "dist-1")
}

func TestGetDistributionsHandler_InternalError_Returns500(t *testing.T) {
	svc := &mockService{
		getDistributions: func(_ context.Context, _ string) ([]model.AdminDistributionRow, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/campaigns/camp-1/distributions", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("camp-1")

	err := h.GetDistributions(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// =====================================================================
// ResetWeek handler tests
// =====================================================================

func TestResetWeekHandler_Success_Returns200(t *testing.T) {
	svc := &mockService{
		weeklyReset: func(_ context.Context) (*model.WeeklyResetResponse, error) {
			return &model.WeeklyResetResponse{
				Status:             "completed",
				ResultsArchived:    50,
				RewardsDistributed: 10,
				NewChallengeID:     "new-ch-1",
			}, nil
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/reset-week", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ResetWeek(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp model.WeeklyResetResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, 50, resp.ResultsArchived)
	assert.Equal(t, 10, resp.RewardsDistributed)
	assert.Equal(t, "new-ch-1", resp.NewChallengeID)
}

func TestResetWeekHandler_NoActiveChallenge_Returns404(t *testing.T) {
	svc := &mockService{
		weeklyReset: func(_ context.Context) (*model.WeeklyResetResponse, error) {
			return nil, fmt.Errorf("locking challenge: %w", model.ErrNotFound)
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/reset-week", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ResetWeek(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestResetWeekHandler_InternalError_Returns500(t *testing.T) {
	svc := &mockService{
		weeklyReset: func(_ context.Context) (*model.WeeklyResetResponse, error) {
			return nil, fmt.Errorf("db connection lost")
		},
	}
	h, e := newTestHandler(svc)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/reset-week", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ResetWeek(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "internal server error")
}
