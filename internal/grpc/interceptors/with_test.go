package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ofstudio/go-shortener/test/pbsuite"
)

type withSuite struct {
	pbsuite.Suite
}

func TestWithSuite(t *testing.T) {
	suite.Run(t, new(withSuite))
}

func (suite *withSuite) SetupSuite() {
	suite.StartServer(grpc.NewServer(
		grpc.UnaryInterceptor(With("pbsuite.AnswerService", testInterceptor)),
	))
}

func (suite *withSuite) TestWith() {

	suite.Run("should return hello world", func() {
		res, err := suite.HelloClient.Hello(context.Background(), &pbsuite.Empty{})
		suite.Require().NoError(err)
		suite.Require().Equal("Hello world!", res.Message)
	})

	suite.Run("should return error'", func() {
		_, err := suite.AnswerClient.Answer(context.Background(), &pbsuite.Empty{})
		suite.Require().Error(err)
		suite.Equal(codes.Code(100500), status.Code(err))
		suite.Equal("Insane!", status.Convert(err).Message())
	})
}

// testInterceptor - всегда возвращает ошибку с кодом 100500 и сообщением "Insane!"
func testInterceptor(_ context.Context, _ interface{}, _ *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (interface{}, error) {
	return nil, status.Error(100500, "Insane!")
}
