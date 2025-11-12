package pullrequests

import "errors"

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrNotFound      = errors.New("not found")
	ErrPRMerged      = errors.New("already merged")
	ErrNotAssigned   = errors.New("not assigned")
	ErrNoCandidate   = errors.New("no condidate")
)
