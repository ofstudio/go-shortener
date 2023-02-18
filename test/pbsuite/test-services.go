package pbsuite

import (
	"context"
)

// HelloService - сервис, который возвращает приветствие
type HelloService struct {
	UnimplementedHelloServiceServer
}

// Hello - возвращает "Hello world!"
func (s *HelloService) Hello(_ context.Context, _ *Empty) (*HelloResponse, error) {
	return &HelloResponse{Message: "Hello world!"}, nil
}

// AnswerService - сервис, который возвращает ответ
type AnswerService struct {
	UnimplementedAnswerServiceServer
}

// Answer - возвращает 42
func (s *AnswerService) Answer(_ context.Context, _ *Empty) (*AnswerResponse, error) {
	return &AnswerResponse{Value: 42}, nil
}
