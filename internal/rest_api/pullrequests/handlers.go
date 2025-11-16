//go:generate mockgen --source=handlers.go --destination=mocks/handlers.go -package=mocks

package pullrequests

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	ucDto "github.com/qwerty268/pull_request_service/internal/usecases/pullrequests"
	"github.com/qwerty268/pull_request_service/internal/utils"
)

type PRCreator interface {
	CreatePR(ctx context.Context, pr ucDto.CreatePROpst) (*ucDto.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*ucDto.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*ucDto.ReassignedRewiew, error)
}

type PRHandlers struct {
	prUsecase PRCreator
}

func NewHandlers(prCreator PRCreator) *PRHandlers {
	return &PRHandlers{
		prUsecase: prCreator,
	}
}

func (h *PRHandlers) RegisterHandlers(e *echo.Echo) {
	e.POST("/pullRequest/create", h.CreatePR)
	e.POST("/pullRequest/merge", h.MergePR)
}

// CreatePR создает PR и назначает ревьюверов
func (h *PRHandlers) CreatePR(c echo.Context) error {
	ctx := context.Background()

	req := new(CreatePRRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ucReq := ucDto.CreatePROpst{
		PullRequestID:   req.PullRequestID,
		PullRequestName: req.PullRequestName,
		AuthorID:        req.AuthorID,
	}

	pr, err := h.prUsecase.CreatePR(ctx, ucReq)
	if err != nil {
		if errors.Is(err, ucDto.ErrAlreadyExists) {
			return utils.ReturnConflict(
				c,
				utils.ErrorDetail{
					Code:    utils.PrExists,
					Message: "PR уже существует",
				},
			)
		}
		if errors.Is(err, ucDto.ErrNotFound) {
			return utils.ReturnNotFound(
				c,
				utils.ErrorDetail{
					Code:    utils.NotFound,
					Message: "Автор/команда не найдены",
				},
			)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, responseFromPr(pr))
}

func (h *PRHandlers) MergePR(c echo.Context) error {
	ctx := context.Background()

	req := new(MergePRRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	pr, err := h.prUsecase.MergePR(ctx, req.PullRequestID)
	if err != nil {
		if errors.Is(err, ucDto.ErrNotFound) {
			return utils.ReturnNotFound(
				c,
				utils.ErrorDetail{
					Code:    utils.NotFound,
					Message: "pull request not found",
				},
			)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, responseFromPr(pr))
}

// ReassignReviewer переназначает ревьювера
func (h *PRHandlers) ReassignReviewer(c echo.Context) error {
	ctx := context.Background()

	req := new(ReassignReviewerRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	reassignedPr, err := h.prUsecase.ReassignReviewer(ctx, req.PullRequestID, req.OldUserID)
	if err != nil {
		if errors.Is(err, ucDto.ErrNotFound) {
			return utils.ReturnNotFound(
				c,
				utils.ErrorDetail{
					Code:    utils.NotFound,
					Message: "PR or user not found",
				},
			)
		}
		if errors.Is(err, ucDto.ErrPRMerged) {
			return utils.ReturnConflict(
				c,
				utils.ErrorDetail{
					Code:    utils.PrMerged,
					Message: "cannot reassign on merged PR",
				},
			)
		}
		if errors.Is(err, ucDto.ErrNotAssigned) {
			return utils.ReturnConflict(
				c,
				utils.ErrorDetail{
					Code:    utils.NotAssigned,
					Message: "reviewer is not assigned to this PR",
				},
			)
		}
		if errors.Is(err, ucDto.ErrNoCandidate) {
			return utils.ReturnConflict(
				c,
				utils.ErrorDetail{
					Code:    utils.NoCandidate,
					Message: "no active replacement candidate in team",
				},
			)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ReassignReviewerResponse{
		PR:         responseFromPr(&reassignedPr.Pr),
		ReplacedBy: reassignedPr.NewReviewer,
	})
}

var emptyTime = time.Time{}

func responseFromPr(ucPr *ucDto.PullRequest) PullRequest {
	response := PullRequest{
		PullRequestID:     ucPr.PullRequestID,
		PullRequestName:   ucPr.PullRequestName,
		AuthorID:          ucPr.AuthorID,
		Status:            ucPr.Status,
		AssignedReviewers: ucPr.AssignedReviewers,
	}

	if ucPr.CreatedAt != emptyTime {
		response.CreatedAt = &ucPr.CreatedAt
	}

	if ucPr.MergedAt != emptyTime {
		response.MergedAt = &ucPr.MergedAt
	}

	return response
}
