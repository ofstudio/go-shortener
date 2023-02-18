package services

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ofstudio/go-shortener/internal/pkgerrors"
)

// Error - возвращает клиенту grpc-ошибку, соответствующую ошибке приложения
func Error(err error) error {
	if appError, ok := err.(*pkgerrors.Error); ok {
		return status.Error(appError.GRPCCode, appError.Error())
	}
	return status.Error(codes.Unknown, "unknown error")
}
