//go:generate mockgen --source=handlers.go --destination=mocks/handlers.go -package=mocks

package users

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	ucDto "github.com/qwerty268/pull_request_service/internal/usecases/users"
)

const (
	userNotFound = "USER_NOT_FOUND"
)

type UserGetter interface {
	SetUserActive(ctx context.Context, userID string, isActive bool) (*ucDto.User, error)
	GetUserReviewRequests(ctx context.Context, userID string) ([]ucDto.PullRequestShort, error)
}

type UserHandlers struct {
	userGetter UserGetter
}

func NewUserHandlers(userGetter UserGetter) *UserHandlers {
	return &UserHandlers{
		userGetter: userGetter,
	}
}

func (h *UserHandlers) RegisterHandlers(e *echo.Echo) {
	e.POST("/users/setIsActive", h.SetUserActive)
	e.GET("/users/getReview", h.GetUserReviewRequests)
}

// SetUserActive устанавливает флаг активности пользователя
func (h *UserHandlers) SetUserActive(c echo.Context) error {
	ctx := context.Background()

	req := new(SetUserActiveRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := h.userGetter.SetUserActive(ctx, req.UserID, *req.IsActive)
	if err != nil {
		if errors.Is(err, ucDto.ErrNotFound) {
			return returnNotFound(
				c,
				ErrorDetail{
					Code:    userNotFound,
					Message: "user not found",
				},
			)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(
		http.StatusOK,
		SetUserActiveResponse{
			UserID:   user.UserID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	)
}

// GetUserReviewRequests получает PR, где пользователь назначен ревьювером
func (h *UserHandlers) GetUserReviewRequests(c echo.Context) error {
	ctx := context.Background()

	req := new(GetUserReviewRequestsRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	prs, err := h.userGetter.GetUserReviewRequests(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, ucDto.ErrNotFound) {
			return returnNotFound(
				c,
				ErrorDetail{
					Code:    userNotFound,
					Message: "user not found",
				},
			)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, GetUserReviewRequestsResponse{
		UserID:       req.UserID,
		PullRequests: prsToPrsResponse(prs),
	})
}

func prsToPrsResponse(ucPrs []ucDto.PullRequestShort) []PullRequestShort {
	prs := make([]PullRequestShort, len(ucPrs))

	for i, v := range ucPrs {
		prs[i] = PullRequestShort(v)
	}
	return prs
}

func returnNotFound(c echo.Context, err ErrorDetail) error {
	return c.JSON(
		http.StatusNotFound,
		ErrorResponse{err},
	)
}
