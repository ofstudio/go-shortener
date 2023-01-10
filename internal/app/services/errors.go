package services

import "errors"

var (
	ErrNotFound   = errors.New("not found")        // ErrNotFound - не найдено
	ErrDuplicate  = errors.New("duplicate")        // ErrDuplicate - дубликат
	ErrDeleted    = errors.New("deleted")          // ErrDeleted - удалено
	ErrInternal   = errors.New("internal error")   // ErrInternal - внутренняя ошибка
	ErrValidation = errors.New("validation error") // ErrValidation - ошибка валидации
)
