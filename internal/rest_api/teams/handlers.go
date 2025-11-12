//go:generate mockgen --source=handlers.go --destination=mocks/handlers.go -package=mocks

package restapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	ucDto "github.com/qwerty268/pull_request_service/internal/usecases/teams"
	"github.com/qwerty268/pull_request_service/internal/utils"
)

type Getter interface {
	AddTeam(ctx context.Context, team ucDto.Team) error
	GetTeam(ctx context.Context, teamName string) (*ucDto.Team, error)
}

type Handlers struct {
	getter Getter
}

func NewHandlers(getter Getter) *Handlers {
	return &Handlers{
		getter: getter,
	}
}

func (h *Handlers) RegisterHandlers(e *echo.Echo) {
	e.POST("/team/add", h.AddTeam)
	e.GET("/team/get", h.GetTeam)
}

func (h *Handlers) AddTeam(c echo.Context) error {
	ctx := context.Background()

	req := new(AddTeamRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.getter.AddTeam(ctx, addTeamrequestToUcDto(req))
	if err != nil {
		if errors.Is(err, ucDto.ErrAlreadyExists) {
			return returnBadRequest(
				c,
				ErrorDetail{
					Code:    utils.TeamExists,
					Message: "team_name already exists",
				},
			)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, req)
}

// GetTeam обработчик для получения информации о команде
func (h *Handlers) GetTeam(c echo.Context) error {
	ctx := context.Background()

	req := new(GetTeamRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	team, err := h.getter.GetTeam(ctx, req.TeamName)
	if err != nil {
		if errors.Is(err, ucDto.ErrNotFound) {
			return returnNotFound(
				c,
				ErrorDetail{
					Code:    utils.NotFound,
					Message: "team not found",
				},
			)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ucDtoToTeamResponse(team))
}

func ucDtoToTeamResponse(team *ucDto.Team) *TeamResponse {
	if team == nil {
		return nil
	}

	resp := &TeamResponse{
		TeamName: team.TeamName,
		Members:  make([]TeamMember, len(team.Members)),
	}

	for i, v := range team.Members {
		resp.Members[i] = TeamMember{
			UserID:   v.UserID,
			Username: v.Username,
			IsActive: &v.IsActive,
		}
	}

	return resp
}

func addTeamrequestToUcDto(req *AddTeamRequest) ucDto.Team {
	team := ucDto.Team{
		TeamName: req.TeamName,
		Members:  make([]ucDto.TeamMember, len(req.Members)),
	}

	for i, v := range req.Members {
		team.Members[i] = ucDto.TeamMember{
			UserID:   v.UserID,
			Username: v.Username,
			IsActive: *v.IsActive,
		}
	}

	return team
}

func returnBadRequest(c echo.Context, err ErrorDetail) error {
	return c.JSON(
		http.StatusBadRequest,
		ErrorResponse{err},
	)
}

func returnNotFound(c echo.Context, err ErrorDetail) error {
	return c.JSON(
		http.StatusNotFound,
		ErrorResponse{err},
	)
}
