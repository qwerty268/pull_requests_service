package restapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/qwerty268/pull_request_service/internal/rest_api/teams/mocks"
	ucDto "github.com/qwerty268/pull_request_service/internal/usecases/teams"
	"github.com/qwerty268/pull_request_service/internal/utils"
)

func Test_AddTeam(t *testing.T) {
	isActive := false
	t.Run("error_validate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		getterMock := mocks.NewMockGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &Handlers{
			getter: getterMock,
		}

		// Тест с невалидным запросом (отсутствует team_name)
		invalidReq := map[string]interface{}{
			"members": []map[string]interface{}{
				{
					"user_id":   "u1",
					"username":  "Alice",
					"is_active": &isActive,
				},
			},
		}

		reqBody, _ := json.Marshal(invalidReq)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		resp := "validate: Key: 'AddTeamRequest.TeamName' Error:Field validation for 'TeamName' failed on the 'required' tag"

		err := h.AddTeam(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
		assert.Equal(t, resp, err.(*echo.HTTPError).Message)
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		getterMock := mocks.NewMockGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &Handlers{
			getter: getterMock,
		}

		validReq := AddTeamRequest{
			TeamName: "backend",
			Members: []TeamMemberRequest{
				{
					UserID:   "u1",
					Username: "Alice",
					IsActive: &isActive,
				},
			},
		}

		getterMock.EXPECT().
			AddTeam(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(1)

		reqBody, _ := json.Marshal(validReq)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.AddTeam(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var response AddTeamRequest
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, validReq.TeamName, response.TeamName)
		assert.Len(t, response.Members, 1)
	})

	t.Run("team_already_exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		getterMock := mocks.NewMockGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &Handlers{
			getter: getterMock,
		}

		validReq := AddTeamRequest{
			TeamName: "backend",
			Members: []TeamMemberRequest{
				{
					UserID:   "u1",
					Username: "Alice",
					IsActive: &isActive,
				},
			},
		}

		getterMock.EXPECT().
			AddTeam(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("ailed to add new teram: %w", ucDto.ErrAlreadyExists)).
			Times(1)

		reqBody, _ := json.Marshal(validReq)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := ErrorResponse{
			Error: ErrorDetail{
				Code:    utils.TeamExists,
				Message: "team_name already exists",
			},
		}

		err := h.AddTeam(c)
		assert.NoError(t, err)

		var response ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})

	t.Run("internal_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		getterMock := mocks.NewMockGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &Handlers{
			getter: getterMock,
		}

		validReq := AddTeamRequest{
			TeamName: "backend",
			Members: []TeamMemberRequest{
				{
					UserID:   "u1",
					Username: "Alice",
					IsActive: &isActive,
				},
			},
		}

		getterMock.EXPECT().
			AddTeam(gomock.Any(), gomock.Any()).
			Return(errors.New("internal error")).
			Times(1)

		reqBody, _ := json.Marshal(validReq)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.AddTeam(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
	})
}

func Test_GetTeam(t *testing.T) {
	t.Run("error_validate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		getterMock := mocks.NewMockGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &Handlers{
			getter: getterMock,
		}

		// Пустой запрос (отсутствует team_name)
		req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		resp := "validate: Key: 'GetTeamRequest.TeamName' Error:Field validation for 'TeamName' failed on the 'required' tag"

		err := h.GetTeam(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
		assert.Equal(t, resp, err.(*echo.HTTPError).Message)
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		getterMock := mocks.NewMockGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &Handlers{
			getter: getterMock,
		}

		expectedTeam := &ucDto.Team{
			TeamName: "backend",
			Members: []ucDto.TeamMember{
				{
					UserID:   "u1",
					Username: "Alice",
					IsActive: true,
				},
			},
		}

		getterMock.EXPECT().
			GetTeam(gomock.Any(), "backend").
			Return(expectedTeam, nil).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.GetTeam(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response TeamResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedTeam.TeamName, response.TeamName)
		assert.Len(t, response.Members, 1)
		assert.Equal(t, expectedTeam.Members[0].UserID, response.Members[0].UserID)
	})

	t.Run("team_not_found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		getterMock := mocks.NewMockGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &Handlers{
			getter: getterMock,
		}

		getterMock.EXPECT().
			GetTeam(gomock.Any(), "unknown").
			Return(nil, ucDto.ErrNotFound).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=unknown", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := ErrorResponse{
			Error: ErrorDetail{
				Code:    utils.NotFound,
				Message: "team not found",
			},
		}

		err := h.GetTeam(c)
		assert.NoError(t, err)

		var response ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		getterMock := mocks.NewMockGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &Handlers{
			getter: getterMock,
		}

		getterMock.EXPECT().
			GetTeam(gomock.Any(), "backend").
			Return(nil, errors.New("internal error")).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.GetTeam(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
	})
}
