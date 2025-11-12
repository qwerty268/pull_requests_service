package utils

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

const (
	TeamExists  = "TEAM_EXISTS"
	PrExists    = "PR_EXISTS"
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

func ReturnNotFound(c echo.Context, err ErrorDetail) error {
	return c.JSON(
		http.StatusNotFound,
		ErrorResponse{err},
	)
}

func ReturnConflict(c echo.Context, err ErrorDetail) error {
	return c.JSON(
		http.StatusConflict,
		ErrorResponse{err},
	)
}
