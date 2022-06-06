package services

import "errors"

var (
	ErrShortURLNotFound = errors.New("short URL not found")
	ErrInternal         = errors.New("internal error")
	ErrValidation       = errors.New("validation error")
)
