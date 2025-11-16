package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("not found")

type Storage struct {
	db    *sqlx.DB
	close func() error
}

func NewStorage(db *sqlx.DB) *Storage {
	return &Storage{
		db: db,
		close: func() error {
			return fmt.Errorf("close: %v", db.Close())
		},
	}
}
func (s *Storage) SetUserActive(userID string, isActive bool) (*User, error) {
	query := `
		UPDATE "user"
		SET is_active = $2
		WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active;
	`

	var user User
	err := s.db.QueryRow(query, userID, isActive).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("SetUserActive query: %w", err)
	}
	return &user, nil
}

func (s *Storage) CheckUserExists(userID string) (bool, error) {
	query := `SELECT user_id FROM "user" WHERE user_id = $1`
	var id string
	err := s.db.QueryRow(query, userID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("CheckUserExists: %w", err)
	}
	return true, nil
}
