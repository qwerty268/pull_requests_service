package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
)

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
)

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

func (s *Storage) AddTeam(team Team) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("start tx: %v", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 1. Создать команду.
	_, err = tx.Exec(`INSERT INTO team (team_name) VALUES ($1)`, team.TeamName)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("insert team: %w", ErrAlreadyExists)
		}
		return fmt.Errorf("insert team: %v", err)
	}

	// 2. Обработать всех пользователей.
	for _, m := range team.Members {
		// Мб пользователь был уже в бд. Тогда перепишем ему комаду.
		_, err = tx.Exec(`
			INSERT INTO "user" (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) 
			DO UPDATE SET 
				username=excluded.username,
				team_name=excluded.team_name, 
				is_active=excluded.is_active
		`, m.UserID, m.Username, team.TeamName, m.IsActive)
		if err != nil {
			return fmt.Errorf("insert/update user: %v", err)
		}

		_, err = tx.Exec(`
			INSERT INTO team_user_map (team_name, user_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, team.TeamName, m.UserID)
		if err != nil {
			return fmt.Errorf("insert team_user_map: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (s *Storage) GetTeam(teamName string) (*Team, error) {
	query := `
	SELECT
		u.user_id,
		u.username,
		u.team_name,
		u.is_active
	FROM user AS u
	JOIN team_user_map AS tum ON tum.user_id = u.user_id
	WHERE tum.team_name = $1
	`

	rows, err := s.db.Query(query, teamName)
	if err != nil {
		return nil, fmt.Errorf("query: %v", err)
	}
	defer rows.Close()

	var members []TeamMember
	for rows.Next() {
		var m TeamMember
		// user_id, username, team_name, is_active
		if err := rows.Scan(&m.UserID, &m.Username, new(string), &m.IsActive); err != nil {
			return nil, fmt.Errorf("scan: %v", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %v", err)
	}

	if len(members) == 0 {
		return nil, ErrNotFound
	}

	team := &Team{
		TeamName: teamName,
		Members:  members,
	}
	return team, nil
}

func (s *Storage) GetUserActiveTeammates(userID string) ([]string, error) {
	query := `
	SELECT u.user_id
	FROM user AS u
	JOIN team_user_map AS tum ON tum.user_id = u.user_id
	WHERE tum.team_name = (SELECT team_name FROM team_user_map WHERE user_id = $1 LIMIT 1)
	AND u.user_id != $1 AND u.is_active
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %v", err)
	}

	members := make([]string, 0)

	for rows.Next() {
		var userID string
		err = rows.Scan(&userID)
		if err != nil {
			return nil, fmt.Errorf("scan: %v", err)
		}
		members = append(members, userID)
	}

	return members, nil
}

func (s *Storage) CheckUserInCommand(userID string) (bool, error) {
	query := `SELECT user_id FROM team_user_map WHERE user_id = $1 LIMIT 1`

	row := s.db.QueryRow(query, userID)

	var dummy string
	err := row.Scan(&dummy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query: %v", err)
	}
	return true, nil
}
