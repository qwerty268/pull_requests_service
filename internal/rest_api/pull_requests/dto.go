package pullrequests

import "time"

// CreatePRRequest - запрос на создание PR
type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
}

// PullRequest - полная информация о PR
type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id" validate:"required"`
	PullRequestName   string     `json:"pull_request_name" validate:"required"`
	AuthorID          string     `json:"author_id" validate:"required"`
	Status            string     `json:"status" validate:"required,oneof=OPEN MERGED"`
	AssignedReviewers []string   `json:"assigned_reviewers" validate:"max=2"`
	CreatedAt         *time.Time `json:"createdAt,omitempty" format:"date-time" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	MergedAt          *time.Time `json:"mergedAt,omitempty" format:"date-time" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
}

// ReassignReviewerRequest - запрос на переназначение ревьювера
type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
	OldUserID     string `json:"old_user_id" validate:"required"`
}

// ReassignReviewerResponse - ответ на переназначение ревьювера
type ReassignReviewerResponse struct {
	PR         PullRequest `json:"pr"`
	ReplacedBy string      `json:"replaced_by"`
}
