package utils

import (
	"fmt"

	"github.com/go-playground/validator"
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
