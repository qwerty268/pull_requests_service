package users

// PullRequestShort - сокращенная информация о PR
type PullRequestShort struct {
	PullRequestID   string
	PullRequestName string
	AuthorID        string
	Status          string
}

// User - пользователь системы
type User struct {
	UserID   string
	Username string
	TeamName string
	IsActive bool
}
