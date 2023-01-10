package services

import "errors"

// ErrNotFound - не найдено
var ErrNotFound = errors.New("not found")

// ErrDuplicate - дубликат
var ErrDuplicate = errors.New("duplicate")

// ErrDeleted - удалено
var ErrDeleted = errors.New("deleted")

// ErrInternal - внутренняя ошибка
var ErrInternal = errors.New("internal error")

// ErrValidation - ошибка валидации
var ErrValidation = errors.New("validation error")
