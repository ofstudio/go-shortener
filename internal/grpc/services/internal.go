package services

import (
	"context"

	"github.com/ofstudio/go-shortener/api/proto"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

// InternalService - реализация gRPC сервиса для внутреннего API.
type InternalService struct {
	proto.UnimplementedInternalServer
	u *usecases.Container
}

// NewInternalService - конструктор InternalService.
func NewInternalService(u *usecases.Container) *InternalService {
	return &InternalService{u: u}
}

// Stats - возвращает статистику сервиса.
func (s *InternalService) Stats(ctx context.Context, _ *proto.StatsRequest) (*proto.StatsResponse, error) {
	usersCount, err := s.u.User.Count(ctx)
	if err != nil {
		return nil, Error(err)
	}
	shortURLCount, err := s.u.ShortURL.Count(ctx)
	if err != nil {
		return nil, Error(err)
	}
	return &proto.StatsResponse{
		Users: uint32(usersCount),
		Urls:  uint32(shortURLCount),
	}, nil
}
