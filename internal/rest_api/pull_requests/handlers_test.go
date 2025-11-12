package pullrequests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/qwerty268/pull_request_service/internal/rest_api/pull_requests/mocks"
	ucDto "github.com/qwerty268/pull_request_service/internal/usecases/pullrequests"
	"github.com/qwerty268/pull_request_service/internal/utils"
)

func Test_CreatePR(t *testing.T) {
	t.Run("error_validate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prCreatorMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prCreatorMock,
		}

		// Тест с невалидным запросом (отсутствует pull_request_id)
		invalidReq := map[string]interface{}{
			"pull_request_name": "Add search",
			"author_id":         "u1",
		}

		reqBody, _ := json.Marshal(invalidReq)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		resp := "validate: Key: 'CreatePRRequest.PullRequestID' Error:Field validation for 'PullRequestID' failed on the 'required' tag"

		err := h.CreatePR(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
		assert.Equal(t, resp, err.(*echo.HTTPError).Message)
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prCreatorMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prCreatorMock,
		}

		reqData := CreatePRRequest{
			PullRequestID:   "pr-1001",
			PullRequestName: "Add search",
			AuthorID:        "u1",
		}

		createdAt := time.Now()
		ucPr := &ucDto.PullRequest{
			PullRequestID:     "pr-1001",
			PullRequestName:   "Add search",
			AuthorID:          "u1",
			Status:            "OPEN",
			AssignedReviewers: []string{"u2", "u3"},
			CreatedAt:         createdAt,
		}

		prCreatorMock.EXPECT().
			CreatePR(gomock.Any(), gomock.Any()).
			Return(ucPr, nil).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedPR := PullRequest{
			PullRequestID:     ucPr.PullRequestID,
			PullRequestName:   ucPr.PullRequestName,
			AuthorID:          ucPr.AuthorID,
			Status:            ucPr.Status,
			AssignedReviewers: ucPr.AssignedReviewers,
			CreatedAt:         &ucPr.CreatedAt,
			MergedAt:          nil,
		}

		err := h.CreatePR(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var actualPR PullRequest
		err = json.Unmarshal(rec.Body.Bytes(), &actualPR)
		assert.NoError(t, err)

		assert.Equal(t, expectedPR.CreatedAt.Format(time.RFC3339), actualPR.CreatedAt.Format(time.RFC3339))
		expectedPR.CreatedAt = actualPR.CreatedAt
		assert.Equal(t, expectedPR, actualPR)
	})

	t.Run("pr_already_exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prCreatorMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prCreatorMock,
		}

		reqData := CreatePRRequest{
			PullRequestID:   "pr-1001",
			PullRequestName: "Add search",
			AuthorID:        "u1",
		}

		prCreatorMock.EXPECT().
			CreatePR(gomock.Any(), gomock.Any()).
			Return(nil, ucDto.ErrAlreadyExists).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := utils.ErrorResponse{
			Error: utils.ErrorDetail{
				Code:    utils.PrExists,
				Message: "PR id already exists",
			},
		}

		err := h.CreatePR(c)
		assert.NoError(t, err)

		var response utils.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("author_not_found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prCreatorMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prCreatorMock,
		}

		reqData := CreatePRRequest{
			PullRequestID:   "pr-1001",
			PullRequestName: "Add search",
			AuthorID:        "unknown",
		}

		prCreatorMock.EXPECT().
			CreatePR(gomock.Any(), gomock.Any()).
			Return(nil, ucDto.ErrNotFound).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := utils.ErrorResponse{
			Error: utils.ErrorDetail{
				Code:    utils.NotFound,
				Message: "author or team not found",
			},
		}

		err := h.CreatePR(c)
		assert.NoError(t, err)

		var response utils.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("internal_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prCreatorMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prCreatorMock,
		}

		reqData := CreatePRRequest{
			PullRequestID:   "pr-1001",
			PullRequestName: "Add search",
			AuthorID:        "u1",
		}

		prCreatorMock.EXPECT().
			CreatePR(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("internal error")).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.CreatePR(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
	})
}

func Test_MergePR(t *testing.T) {
	t.Run("error_validate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prUsecaseMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prUsecaseMock,
		}

		// Тест с пустым запросом (отсутствует pull_request_id)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader([]byte("{}")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		resp := "validate: Key: 'MergePRRequest.PullRequestID' Error:Field validation for 'PullRequestID' failed on the 'required' tag"

		err := h.MergePR(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
		assert.Equal(t, resp, err.(*echo.HTTPError).Message)
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prUsecaseMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prUsecaseMock,
		}

		reqData := MergePRRequest{
			PullRequestID: "pr-1001",
		}

		mergedAt := time.Now()
		usecasePr := &ucDto.PullRequest{
			PullRequestID:     "pr-1001",
			PullRequestName:   "Add search",
			AuthorID:          "u1",
			Status:            "MERGED",
			AssignedReviewers: []string{"u2", "u3"},
			MergedAt:          mergedAt,
		}

		prUsecaseMock.EXPECT().
			MergePR(gomock.Any(), "pr-1001").
			Return(usecasePr, nil).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedPr := PullRequest{
			PullRequestID:     usecasePr.PullRequestID,
			PullRequestName:   usecasePr.PullRequestName,
			AuthorID:          usecasePr.AuthorID,
			Status:            usecasePr.Status,
			AssignedReviewers: usecasePr.AssignedReviewers,
			MergedAt:          &usecasePr.MergedAt,
		}

		err := h.MergePR(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var actualPr PullRequest
		err = json.Unmarshal(rec.Body.Bytes(), &actualPr)
		assert.NoError(t, err)

		assert.Equal(t, expectedPr.MergedAt.Format(time.RFC3339), actualPr.MergedAt.Format(time.RFC3339))
		expectedPr.MergedAt = actualPr.MergedAt
		assert.Equal(t, expectedPr, actualPr)
	})

	t.Run("pr_not_found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prUsecaseMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prUsecaseMock,
		}

		reqData := MergePRRequest{
			PullRequestID: "unknown",
		}

		prUsecaseMock.EXPECT().
			MergePR(gomock.Any(), "unknown").
			Return(nil, ucDto.ErrNotFound).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := utils.ErrorResponse{
			Error: utils.ErrorDetail{
				Code:    utils.NotFound,
				Message: "pull request not found",
			},
		}

		err := h.MergePR(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		var response utils.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})

	t.Run("internal_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prUsecaseMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prUsecaseMock,
		}

		reqData := MergePRRequest{
			PullRequestID: "pr-1001",
		}

		prUsecaseMock.EXPECT().
			MergePR(gomock.Any(), "pr-1001").
			Return(nil, errors.New("internal error")).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.MergePR(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
	})
}

func Test_ReassignReviewer(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prUsecaseMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prUsecaseMock,
		}

		reqData := ReassignReviewerRequest{
			PullRequestID: "pr-1001",
			OldUserID:     "u2",
		}

		usecasePR := &ucDto.PullRequest{
			PullRequestID:     "pr-1001",
			PullRequestName:   "Add search",
			AuthorID:          "u1",
			Status:            "OPEN",
			AssignedReviewers: []string{"u3", "u5"},
			CreatedAt:         time.Now(),
		}
		newReviewer := "u5"

		prUsecaseMock.EXPECT().
			ReassignReviewer(gomock.Any(), "pr-1001", "u2").
			Return(usecasePR, newReviewer, nil).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := ReassignReviewerResponse{
			PR: PullRequest{
				PullRequestID:     usecasePR.PullRequestID,
				PullRequestName:   usecasePR.PullRequestName,
				AuthorID:          usecasePR.AuthorID,
				Status:            usecasePR.Status,
				AssignedReviewers: usecasePR.AssignedReviewers,
				CreatedAt:         &usecasePR.CreatedAt,
			},
			ReplacedBy: newReviewer,
		}

		err := h.ReassignReviewer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var actualPR ReassignReviewerResponse
		err = json.Unmarshal(rec.Body.Bytes(), &actualPR)
		assert.NoError(t, err)

		assert.Equal(t, expectedResponse.PR.CreatedAt.Format(time.RFC3339), actualPR.PR.CreatedAt.Format(time.RFC3339))
		expectedResponse.PR.CreatedAt = actualPR.PR.CreatedAt
		assert.Equal(t, expectedResponse, actualPR)
	})

	t.Run("error_validate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prUsecaseMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prUsecaseMock,
		}

		// Тест с невалидным запросом (отсутствует pull_request_id)
		invalidReq := map[string]interface{}{
			"old_user_id": "u2",
		}

		reqBody, _ := json.Marshal(invalidReq)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedError := "validate: Key: 'ReassignReviewerRequest.PullRequestID' Error:Field validation for 'PullRequestID' failed on the 'required' tag"

		err := h.ReassignReviewer(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, err.(*echo.HTTPError).Code)
		assert.Equal(t, expectedError, err.(*echo.HTTPError).Message)
	})

	t.Run("pr_not_found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prUsecaseMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prUsecaseMock,
		}

		reqData := ReassignReviewerRequest{
			PullRequestID: "unknown",
			OldUserID:     "u2",
		}

		prUsecaseMock.EXPECT().
			ReassignReviewer(gomock.Any(), "unknown", "u2").
			Return(nil, "", ucDto.ErrNotFound).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := utils.ErrorResponse{
			Error: utils.ErrorDetail{
				Code:    utils.NotFound,
				Message: "PR or user not found",
			},
		}

		err := h.ReassignReviewer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		var response utils.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})

	t.Run("pr_merged", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prUsecaseMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prUsecaseMock,
		}

		reqData := ReassignReviewerRequest{
			PullRequestID: "pr-1001",
			OldUserID:     "u2",
		}

		prUsecaseMock.EXPECT().
			ReassignReviewer(gomock.Any(), "pr-1001", "u2").
			Return(nil, "", ucDto.ErrPRMerged).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := utils.ErrorResponse{
			Error: utils.ErrorDetail{
				Code:    utils.PrMerged,
				Message: "cannot reassign on merged PR",
			},
		}

		err := h.ReassignReviewer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, rec.Code)

		var response utils.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})

	t.Run("reviewer_not_assigned", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prUsecaseMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prUsecaseMock,
		}

		reqData := ReassignReviewerRequest{
			PullRequestID: "pr-1001",
			OldUserID:     "u99",
		}

		prUsecaseMock.EXPECT().
			ReassignReviewer(gomock.Any(), "pr-1001", "u99").
			Return(nil, "", ucDto.ErrNotAssigned).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := utils.ErrorResponse{
			Error: utils.ErrorDetail{
				Code:    utils.NotAssigned,
				Message: "reviewer is not assigned to this PR",
			},
		}

		err := h.ReassignReviewer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, rec.Code)

		var response utils.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})

	t.Run("no_candidate", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prUsecaseMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prUsecaseMock,
		}

		reqData := ReassignReviewerRequest{
			PullRequestID: "pr-1001",
			OldUserID:     "u2",
		}

		prUsecaseMock.EXPECT().
			ReassignReviewer(gomock.Any(), "pr-1001", "u2").
			Return(nil, "", ucDto.ErrNoCandidate).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expectedResponse := utils.ErrorResponse{
			Error: utils.ErrorDetail{
				Code:    utils.NoCandidate,
				Message: "no active replacement candidate in team",
			},
		}

		err := h.ReassignReviewer(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, rec.Code)

		var response utils.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)
	})

	t.Run("internal_error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		prUsecaseMock := mocks.NewMockPRCreator(ctrl)

		e := echo.New()
		e.Validator = utils.NewHTTPRequestValidator()

		h := &PRHandlers{
			prUsecase: prUsecaseMock,
		}

		reqData := ReassignReviewerRequest{
			PullRequestID: "pr-1001",
			OldUserID:     "u2",
		}

		prUsecaseMock.EXPECT().
			ReassignReviewer(gomock.Any(), "pr-1001", "u2").
			Return(nil, "", errors.New("internal error")).
			Times(1)

		reqBody, _ := json.Marshal(reqData)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := h.ReassignReviewer(c)
		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
	})
}
