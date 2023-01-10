package services

import "errors"

var (
	// ErrNotFound - не найдено
	ErrNotFound = errors.New("not found")
	// ErrDuplicate - дубликат
	ErrDuplicate = errors.New("duplicate")
	// ErrDeleted - удалено
	ErrDeleted = errors.New("deleted")
	// ErrInternal - внутренняя ошибка
	ErrInternal = errors.New("internal error")
	// ErrValidation - ошибка валидации
	ErrValidation = errors.New("validation error")
)
