package storage

type PullRequestShort struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          string
}

type User struct {
	UserID   string
	Username string
	TeamName string
	IsActive bool
}
