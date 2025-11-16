package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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

func (s *Storage) GetUserReviewRequests(userID string) ([]PullRequestShort, error) {
	query := `
	SELECT
		pr.pull_request_id,
		pr.pull_request_name,
		pr.AuthorID,
		pr.isMerged
	FROM pull_request AS pr
	JOIN pr_reviewers_map AS prm ON prm.pull_request_id = pr.pull_request_id
	WHERE prm.user_id = $1;
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %v", err)
	}
	defer rows.Close()

	var prs []PullRequestShort
	for rows.Next() {
		var pr PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.IsMerged); err != nil {
			return nil, fmt.Errorf("scan: %v", err)
		}
		prs = append(prs, pr)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %v", err)
	}

	if len(prs) == 0 {
		return nil, ErrNotFound
	}

	return prs, nil
}

func (s *Storage) AddPr(pr PullRequest) error {
	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Добавляем pull_request.
	_, err = tx.Exec(`
		INSERT INTO pull_request 
			(pull_request_id, pull_request_name, author_id, is_merged, assigned_reviewers, created_at, merged_at)
		VALUES 
			($1, $2, $3, $4, $5, $6, $7)
	`, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.IsMerged, pq.Array(pr.AssignedReviewers), pr.CreatedAt, pr.MergedAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return fmt.Errorf("insert team: %w", ErrAlreadyExists)
		}
		return fmt.Errorf("insert pull_request: %v", err)
	}

	// 2. Добавляем всех ревьюеров в pr_reviewers_map.
	for _, reviewerID := range pr.AssignedReviewers {
		_, err := tx.Exec(`
			INSERT INTO pr_reviewers_map (pull_request_id, user_id)
			VALUES ($1, $2)
		`, pr.PullRequestID, reviewerID)
		if err != nil {
			return fmt.Errorf("insert pr_reviewers_map: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (s *Storage) SetPrMerged(prID string) (*PullRequest, error) {
	now := time.Now()
	query := `
		UPDATE pull_request
		SET is_merged = TRUE, merged_at = $2
		WHERE pull_request_id = $1
		RETURNING pull_request_id, pull_request_name, author_id, is_merged, assigned_reviewers, created_at, merged_at
	`
	var pr PullRequest
	err := s.db.QueryRow(query, prID, now).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.IsMerged,
		pq.Array(&pr.AssignedReviewers),
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("SetPrMerged: %w", err)
	}
	return &pr, nil
}

func (s *Storage) CheckUserInPr(prID, userID string) (bool, error) {
	query := `
		SELECT 1
		FROM pr_reviewers_map
		WHERE pull_request_id = $1 AND user_id = $2
		LIMIT 1
	`
	var dummy int
	err := s.db.QueryRow(query, prID, userID).Scan(&dummy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("CheckUserInPr: %w", err)
	}
	return true, nil
}

func (s *Storage) GetPrByID(prID string) (*PullRequest, error) {
	query := `
        SELECT 
            pull_request_id, 
            pull_request_name, 
            author_id, 
            is_merged, 
            assigned_reviewers,
            created_at,
            merged_at
        FROM pull_request 
        WHERE pull_request_id = $1
    `
	var pr PullRequest
	err := s.db.QueryRow(query, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.IsMerged,
		pq.Array(&pr.AssignedReviewers),
		&pr.CreatedAt,
		&pr.MergedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("GetPrByID: %w", err)
	}
	return &pr, nil
}

func (s *Storage) ResetPrMember(filter ResetReviewerFilter) error {
	query := `
		UPDATE pr_reviewers_map
		SET user_id = $3
		WHERE pull_request_id = $1 AND user_id = $2
	`
	_, err := s.db.Exec(query, filter.PrID, filter.OldUserID, filter.NewUserID)
	if err != nil {
		return fmt.Errorf("ResetPrMember: %w", err)
	}
	return nil
}
