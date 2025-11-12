package restapi

// GetTeamRequest - параметр запроса для имени команды
// c.QueryParam("team_name") в Echo
type GetTeamRequest struct {
	TeamName string `query:"team_name" validate:"required"`
}

type ErrorDetail struct {
	Code    string `json:"code" validate:"required,oneof=TEAM_EXISTS PR_EXISTS PR_MERGED NOT_ASSIGNED NO_CANDIDATE NOT_FOUND"`
	Message string `json:"message" validate:"required"`
}

// ErrorResponse - структура для ошибок API
type ErrorResponse struct {
	Error ErrorDetail `json:"error" validate:"required"`
}

// TeamMemberRequest - участник команды
type TeamMemberRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	IsActive *bool  `json:"is_active" validate:"required"`
}

// AddTeamRequest - команда с участниками
type AddTeamRequest struct {
	TeamName string              `json:"team_name" validate:"required"`
	Members  []TeamMemberRequest `json:"members" validate:"required,dive"`
}

type TeamMemberResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResponse struct {
	TeamName string               `json:"team_name"`
	Members  []TeamMemberResponse `json:"members"`
}
