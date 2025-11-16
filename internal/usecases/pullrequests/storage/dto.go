package storage

import "time"

type PullRequest struct {
	PullRequestID     string
	PullRequestName   string
	AuthorID          string
	IsMerged          bool
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          time.Time
}

type ResetReviewerFilter struct {
	PrID      string
	OldUserID string
	NewUserID string
}

type PullRequestShort struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	IsMerged        bool
}
