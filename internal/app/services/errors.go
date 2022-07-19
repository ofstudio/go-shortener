package services

import "errors"

var (
	ErrNotFound   = errors.New("not found")
	ErrDuplicate  = errors.New("duplicate")
	ErrDeleted    = errors.New("deleted")
	ErrInternal   = errors.New("internal error")
	ErrValidation = errors.New("validation error")
)
