package pullrequests

import "time"

type CreatePROpst struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
}

// PullRequest - полная информация о PR
type PullRequest struct {
	PullRequestID     string
	PullRequestName   string
	AuthorID          string
	Status            string
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          time.Time
}

type ReassignedRewiew struct {
	Pr          PullRequest
	NewReviewer string
}
