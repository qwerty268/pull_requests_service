package users

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/qwerty268/pull_request_service/internal/rest_api/users/mocks"
	ucDto "github.com/qwerty268/pull_request_service/internal/usecases/users"
	"github.com/qwerty268/pull_request_service/internal/utils"
)

func Test_SetUserActive(t *testing.T) {
	isActive := false
	t.Run("error_validate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userGetterMock := mocks.NewMockUserGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &UserHandlers{
			userGetter: userGetterMock,
		}

		// Тест с невалидным запросом (отсутствует user_id)
		invalidReq := map[string]interface{}{
			"is_active": true,
		}

		reqBody, _ := json.Marshal(invalidReq)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		resp := "validate: Key: 'SetUserActiveRequest.UserID' Error:Field validation for 'UserID' failed on the 'required' tag"

		err := h.SetUserActive(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
		assert.Equal(t, resp, err.(*echo.HTTPError).Message)
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userGetterMock := mocks.NewMockUserGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &UserHandlers{
			userGetter: userGetterMock,
		}

		reqData := SetUserActiveRequest{
			UserID:   "u1",
			IsActive: &isActive,
		}

		expectedUser := &ucDto.User{
			UserID:   "u1",
			Username: "Alice",
			TeamName: "backend",
			IsActive: false,
		}

		userGetterMock.EXPECT().
			SetUserActive(gomock.Any(), "u1", isActive).
			Return(expectedUser, nil).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.SetUserActive(c)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, rec.Code)

		var response SetUserActiveResponse

		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, SetUserActiveResponse(*expectedUser), response)
	})

	t.Run("user_not_found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userGetterMock := mocks.NewMockUserGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &UserHandlers{
			userGetter: userGetterMock,
		}

		localIsActive := true
		reqData := SetUserActiveRequest{
			UserID:   "unknown",
			IsActive: &localIsActive,
		}

		userGetterMock.EXPECT().
			SetUserActive(gomock.Any(), "unknown", localIsActive).
			Return(nil, ucDto.ErrNotFound).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := ErrorResponse{
			Error: ErrorDetail{
				Code:    userNotFound,
				Message: "user not found",
			},
		}

		err := h.SetUserActive(c)
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

		userGetterMock := mocks.NewMockUserGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &UserHandlers{
			userGetter: userGetterMock,
		}

		reqData := SetUserActiveRequest{
			UserID:   "u1",
			IsActive: &isActive,
		}

		userGetterMock.EXPECT().
			SetUserActive(gomock.Any(), "u1", isActive).
			Return(nil, errors.New("internal error")).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.SetUserActive(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
	})
}

func Test_GetUserReviewRequests(t *testing.T) {
	t.Run("error_validate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userGetterMock := mocks.NewMockUserGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &UserHandlers{
			userGetter: userGetterMock,
		}

		// Пустой запрос (отсутствует user_id)
		req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		resp := "validate: Key: 'GetUserReviewRequestsRequest.UserID' Error:Field validation for 'UserID' failed on the 'required' tag"

		err := h.GetUserReviewRequests(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
		assert.Equal(t, resp, err.(*echo.HTTPError).Message)
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userGetterMock := mocks.NewMockUserGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &UserHandlers{
			userGetter: userGetterMock,
		}

		expectedPrs := []ucDto.PullRequestShort{
			{
				PullRequestID:   "pr-1001",
				PullRequestName: "Add search",
				AuthorID:        "u1",
				Status:          "OPEN",
			},
		}

		userGetterMock.EXPECT().
			GetUserReviewRequests(gomock.Any(), "u2").
			Return(expectedPrs, nil).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.GetUserReviewRequests(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response GetUserReviewRequestsResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "u2", response.UserID)
		assert.Len(t, response.PullRequests, 1)
		assert.Equal(t, expectedPrs[0].PullRequestID, response.PullRequests[0].PullRequestID)
	})

	t.Run("user_not_found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		userGetterMock := mocks.NewMockUserGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &UserHandlers{
			userGetter: userGetterMock,
		}

		userGetterMock.EXPECT().
			GetUserReviewRequests(gomock.Any(), "unknown").
			Return(nil, ucDto.ErrNotFound).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=unknown", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := ErrorResponse{
			Error: ErrorDetail{
				Code:    userNotFound,
				Message: "user not found",
			},
		}

		err := h.GetUserReviewRequests(c)
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

		userGetterMock := mocks.NewMockUserGetter(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &UserHandlers{
			userGetter: userGetterMock,
		}

		userGetterMock.EXPECT().
			GetUserReviewRequests(gomock.Any(), "u1").
			Return(nil, errors.New("internal error")).
			Times(1)

		req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.GetUserReviewRequests(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
	})
}
