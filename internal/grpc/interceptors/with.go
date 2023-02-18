package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
)

// With - унарный серверный интерцептор, который применяет interceptor только к вызовам сервиса service.
func With(service string, interceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !strings.HasPrefix(info.FullMethod, "/"+service+"/") {
			return handler(ctx, req)
		}
		return interceptor(ctx, req, info, handler)
	}
}
