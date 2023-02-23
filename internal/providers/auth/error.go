package auth

import "errors"

// ErrSigningError - ошибка при подписании токена
var ErrSigningError = errors.New("signing error")

// ErrInvalidToken - невалидный токен
var ErrInvalidToken = errors.New("invalid token")

// ErrExpiredToken - просроченный токен
var ErrExpiredToken = errors.New("expired token")
