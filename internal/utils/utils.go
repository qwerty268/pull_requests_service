package utils

import (
	"fmt"

	"github.com/go-playground/validator"
)

const (
	TeamExists  = "TEAM_EXISTS"
	PeExists    = "PR_EXISTS"
	PrMerged    = "PR_MERGED"
	NotAssigned = "NOT_ASSIGNED"
	NoCandidate = "NO_CANDIDATE"
	NotFound    = "NOT_FOUND"
)


type HTTPRequestValidator struct {
	validator *validator.Validate
}

func NewHTTPRequestValidator() *HTTPRequestValidator {
	return &HTTPRequestValidator{
		validator: validator.New(),
	}
}

func (cv *HTTPRequestValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		return fmt.Errorf("validate: %v", err)
	}
	return nil
}
