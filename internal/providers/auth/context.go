package auth

import "context"

type ctxKey struct {
	name string
}

// userIDKey - ключ для userID в контексте запроса
var userIDKey = &ctxKey{"user_id"}

// FromContext - возвращает userID из контекста
func FromContext(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(userIDKey).(uint)
	return userID, ok
}

// ToContext - добавляет userID в контекст
func ToContext(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
