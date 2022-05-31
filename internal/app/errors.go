package app

import "errors"

var (
	ErrURLNotFound = errors.New("URL not found")
	ErrInternal    = errors.New("internal error")
	ErrValidation  = errors.New("validation error")
)
