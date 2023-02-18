package services

import (
	"context"
	"errors"

	"google.golang.org/grpc/status"

	"github.com/ofstudio/go-shortener/api/proto"
	"github.com/ofstudio/go-shortener/internal/pkgerrors"
	"github.com/ofstudio/go-shortener/internal/providers/auth"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

// ShortURLService - реализация gRPC сервиса для работы с короткими ссылками.
type ShortURLService struct {
	proto.UnimplementedShortURLServer
	u *usecases.Container
}

// NewShortURLService - конструктор ShortURLService.
func NewShortURLService(u *usecases.Container) *ShortURLService {
	return &ShortURLService{u: u}
}

func (s ShortURLService) Create(ctx context.Context, request *proto.ShortURLCreateRequest) (*proto.ShortURLCreateResponse, error) {
	// Проверяем аутентифицирован ли пользователь
	userID, ok := auth.FromContext(ctx)
	if !ok {
		return nil, Error(pkgerrors.ErrAuth)
	}

	// Создаем короткую ссылку
	shortURL, err := s.u.ShortURL.Create(ctx, userID, request.Url)
	if err != nil && !errors.Is(err, pkgerrors.ErrDuplicate) {
		return nil, Error(err)
	}

	// Возвращаем результат
	return &proto.ShortURLCreateResponse{
		Result: s.u.ShortURL.Resolve(shortURL.ID),
	}, nil
}

func (s ShortURLService) CreateBatch(ctx context.Context, request *proto.ShortURLCreateBatchRequest) (*proto.ShortURLCreateBatchResponse, error) {
	// Проверяем аутентифицирован ли пользователь
	userID, ok := auth.FromContext(ctx)
	if !ok {
		return nil, Error(pkgerrors.ErrAuth)
	}

	// Создаем короткие ссылки
	res := &proto.ShortURLCreateBatchResponse{
		Items: make([]*proto.ShortURLCreateBatchResponse_Item, 0, len(request.Items)),
	}
	for _, item := range request.Items {
		shortUrl, err := s.u.ShortURL.Create(ctx, userID, item.OriginalUrl)
		if err != nil && !errors.Is(err, pkgerrors.ErrDuplicate) {
			return nil, Error(err)
		}
		res.Items = append(res.Items, &proto.ShortURLCreateBatchResponse_Item{
			CorrelationId: item.CorrelationId,
			ShortUrl:      s.u.ShortURL.Resolve(shortUrl.ID),
		})
	}
	return res, nil
}

func (s ShortURLService) DeleteBatch(ctx context.Context, request *proto.ShortURLDeleteBatchRequest) (*proto.ShortURLDeleteBatchResponse, error) {
	// Проверяем аутентифицирован ли пользователь
	userID, ok := auth.FromContext(ctx)
	if !ok {
		return nil, Error(pkgerrors.ErrAuth)
	}

	if len(request.Items) == 0 {
		return nil, Error(pkgerrors.ErrValidation)
	}
	// Удаляем короткие ссылки
	go func() {
		_ = s.u.ShortURL.DeleteBatch(ctx, userID, request.Items)
	}()
	// Не дожидаемся окончания удаления, возвращаем ответ
	return &proto.ShortURLDeleteBatchResponse{}, nil
}

func (s ShortURLService) GetByUserID(ctx context.Context, _ *proto.ShortURLGetByUserIDRequest) (*proto.ShortURLGetByUserIDResponse, error) {
	// Проверяем аутентифицирован ли пользователь
	userID, ok := auth.FromContext(ctx)
	if !ok {
		return nil, Error(pkgerrors.ErrAuth)
	}

	// Получаем короткие ссылки
	shortURLs, err := s.u.ShortURL.GetByUserID(ctx, userID)
	if err != nil {
		return nil, Error(err)
	}
	// Если коротких ссылок нет, возвращаем no content
	if len(shortURLs) == 0 {
		return nil, status.Error(pkgerrors.GRPCNoContent, "no content")
	}

	// Формируем ответ
	res := &proto.ShortURLGetByUserIDResponse{
		Items: make([]*proto.ShortURLGetByUserIDResponse_Item, 0, len(shortURLs)),
	}
	for _, shortURL := range shortURLs {
		res.Items = append(res.Items, &proto.ShortURLGetByUserIDResponse_Item{
			OriginalUrl: shortURL.OriginalURL,
			ShortUrl:    s.u.ShortURL.Resolve(shortURL.ID),
		})
	}
	return res, nil
}
