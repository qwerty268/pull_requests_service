package users

type SetUserActiveRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive *bool  `json:"is_active" validate:"required"`
}

type SetUserActiveResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type GetUserReviewRequestsRequest struct {
	UserID string `query:"user_id" validate:"required"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
	Status          string `json:"status" validate:"required,oneof=OPEN MERGED"`
}

type GetUserReviewRequestsResponse struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}

type ErrorDetail struct {
	Code    string `json:"code" validate:"required,oneof=TEAM_EXISTS PR_EXISTS PR_MERGED NOT_ASSIGNED NO_CANDIDATE NOT_FOUND"`
	Message string `json:"message" validate:"required"`
}

// ErrorResponse - структура для ошибок API
type ErrorResponse struct {
	Error ErrorDetail `json:"error" validate:"required"`
}
