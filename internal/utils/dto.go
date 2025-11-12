package utils

type ErrorDetail struct {
	Code    string `json:"code" validate:"required,oneof=TEAM_EXISTS PR_EXISTS PR_MERGED NOT_ASSIGNED NO_CANDIDATE NOT_FOUND"`
	Message string `json:"message" validate:"required"`
}

// ErrorResponse - структура для ошибок API
type ErrorResponse struct {
	Error ErrorDetail `json:"error" validate:"required"`
}
