package restapi

// GetTeamRequest - параметр запроса для имени команды
// c.QueryParam("team_name") в Echo
type GetTeamRequest struct {
	TeamName string `query:"team_name" validate:"required"`
}

// UserIdQuery - параметр запроса для ID пользователя
// c.QueryParam("user_id") в Echo
type UserIdQuery struct {
	UserID string `query:"user_id" validate:"required"`
}

type ErrorDetail struct {
	Code    string `json:"code" validate:"required,oneof=TEAM_EXISTS PR_EXISTS PR_MERGED NOT_ASSIGNED NO_CANDIDATE NOT_FOUND"`
	Message string `json:"message" validate:"required"`
}

// ErrorResponse - структура для ошибок API
type ErrorResponse struct {
	Error ErrorDetail `json:"error" validate:"required"`
}

// TeamMember - участник команды
type TeamMember struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	IsActive bool   `json:"is_active" validate:"required"`
}

// AddTeamRequest - команда с участниками
type AddTeamRequest struct {
	TeamName string       `json:"team_name" validate:"required"`
	Members  []TeamMember `json:"members" validate:"required,dive"`
}

type TeamResponse struct {
	TeamName string       `json:"team_name" validate:"required"`
	Members  []TeamMember `json:"members" validate:"required,dive"`
}

// User - пользователь системы
type User struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	TeamName string `json:"team_name" validate:"required"`
	IsActive bool   `json:"is_active" validate:"required"`
}

// PullRequest - полная информация о PR
type PullRequest struct {
	PullRequestID     string   `json:"pull_request_id" validate:"required"`
	PullRequestName   string   `json:"pull_request_name" validate:"required"`
	AuthorID          string   `json:"author_id" validate:"required"`
	Status            string   `json:"status" validate:"required,oneof=OPEN MERGED"`
	AssignedReviewers []string `json:"assigned_reviewers" validate:"max=2"`
	CreatedAt         *string  `json:"createdAt,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	MergedAt          *string  `json:"mergedAt,omitempty" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

// PullRequestShort - сокращенная информация о PR
type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
	Status          string `json:"status" validate:"required,oneof=OPEN MERGED"`
}
