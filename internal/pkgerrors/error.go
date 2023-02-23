// Package pkgerrors - пакет для работы с ошибками приложения
package pkgerrors

import (
	"errors"
	"net/http"

	"google.golang.org/grpc/codes"
)

// ErrNotFound - не найдено
var ErrNotFound = NewError(http.StatusNotFound, codes.NotFound, "not found")

// ErrValidation - ошибка валидации
var ErrValidation = NewError(http.StatusBadRequest, codes.InvalidArgument, "validation error")

// ErrAuth - ошибка авторизации
var ErrAuth = NewError(http.StatusUnauthorized, codes.Unauthenticated, "unauthorized")

// ErrDuplicate - дубликат
var ErrDuplicate = NewError(http.StatusConflict, codes.AlreadyExists, "duplicate")

// ErrDeleted - удалено
var ErrDeleted = NewError(http.StatusGone, GRPCDeleted, "deleted")

// ErrInternal - внутренняя ошибка
var ErrInternal = NewError(http.StatusInternalServerError, codes.Internal, "internal error")

// Error - ошибка приложения
type Error struct {
	error
	HTTPStatus int
	GRPCCode   codes.Code
}

// NewError - конструктор ошибки
func NewError(httpStatus int, grpcCode codes.Code, message string) *Error {
	return &Error{
		error:      errors.New(message),
		HTTPStatus: httpStatus,
		GRPCCode:   grpcCode,
	}
}

// Кастомные коды ошибок GRPC
const (
	// GRPCNoContent - нет контента
	GRPCNoContent codes.Code = 10204
	// GRPCDeleted - удалено
	GRPCDeleted codes.Code = 10410
)
