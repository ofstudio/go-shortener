package app

import (
	"errors"
	"net/http"
)

var (
	// ErrNotFound - не найдено
	ErrNotFound = NewError(http.StatusNotFound, "not found")

	// ErrValidation - ошибка валидации
	ErrValidation = NewError(http.StatusBadRequest, "validation error")

	// ErrAuth - ошибка авторизации
	ErrAuth = NewError(http.StatusUnauthorized, "unauthorized")

	// ErrDuplicate - дубликат
	ErrDuplicate = NewError(http.StatusConflict, "duplicate")

	// ErrDeleted - удалено
	ErrDeleted = NewError(http.StatusGone, "deleted")

	// ErrInternal - внутренняя ошибка
	ErrInternal = NewError(http.StatusInternalServerError, "internal error")
)

// Error - ошибка приложения
type Error struct {
	error
	HTTPStatus int
}

// NewError - конструктор ошибки
func NewError(httpStatus int, message string) *Error {
	return &Error{
		error:      errors.New(message),
		HTTPStatus: httpStatus,
	}
}
