package services

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrInternal   = errors.New("internal error")
	ErrValidation = errors.New("validation error")
	ErrDuplicate  = errors.New("duplicate")
)
