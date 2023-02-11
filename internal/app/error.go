package app

import (
	"errors"
	"net/http"
)

// ErrNotFound - не найдено
var ErrNotFound = NewError(http.StatusNotFound, "not found")

// ErrValidation - ошибка валидации
var ErrValidation = NewError(http.StatusBadRequest, "validation error")

// ErrAuth - ошибка авторизации
var ErrAuth = NewError(http.StatusUnauthorized, "unauthorized")

// ErrDuplicate - дубликат
var ErrDuplicate = NewError(http.StatusConflict, "duplicate")

// ErrDeleted - удалено
var ErrDeleted = NewError(http.StatusGone, "deleted")

// ErrInternal - внутренняя ошибка
var ErrInternal = NewError(http.StatusInternalServerError, "internal error")

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
