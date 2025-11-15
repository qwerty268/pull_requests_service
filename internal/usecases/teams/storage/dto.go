package storage

type TeamMember struct {
	UserID   string
	Username string
	IsActive bool
}

type Team struct {
	TeamName string
	Members  []TeamMember
}
