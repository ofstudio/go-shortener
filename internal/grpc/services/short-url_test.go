package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ofstudio/go-shortener/api/proto"
	"github.com/ofstudio/go-shortener/internal/config"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/pkgerrors"
	"github.com/ofstudio/go-shortener/internal/providers/auth"
	"github.com/ofstudio/go-shortener/internal/repo"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

type ShortURLServiceSuite struct {
	suite.Suite
	u *usecases.Container
	s *ShortURLService
}

func (suite *ShortURLServiceSuite) SetupTest() {
	cfg, _ := config.Default(nil)
	suite.u = usecases.NewContainer(context.Background(), cfg, repo.NewMemoryRepo())
	// Создаем двух тестовых пользователей
	suite.Require().NoError(suite.u.User.Create(context.Background(), &models.User{ID: 1}))
	suite.Require().NoError(suite.u.User.Create(context.Background(), &models.User{ID: 2}))
	suite.s = NewShortURLService(suite.u)
}

func (suite *ShortURLServiceSuite) TestCreate() {
	suite.Run("unauthenticated", func() {
		_, err := suite.s.Create(context.Background(), &proto.ShortURLCreateRequest{Url: "https://google.com"})
		suite.Require().Error(err)
		st, ok := status.FromError(err)
		suite.Require().True(ok)
		suite.Equal(codes.Unauthenticated, st.Code())
	})

	suite.Run("should create short url", func() {
		ctx := auth.ToContext(context.Background(), 1)
		res, err := suite.s.Create(ctx, &proto.ShortURLCreateRequest{Url: "https://google.com"})
		suite.Require().NoError(err)
		suite.NotEmpty(res.Result)
	})

	suite.Run("should return error if url is invalid", func() {
		ctx := auth.ToContext(context.Background(), 1)
		_, err := suite.s.Create(ctx, &proto.ShortURLCreateRequest{Url: "invalid-url"})
		suite.Require().Error(err)
		st, ok := status.FromError(err)
		suite.Require().True(ok)
		suite.Equal(codes.InvalidArgument, st.Code())
	})

	suite.Run("should same short return url if duplicate", func() {
		ctx := auth.ToContext(context.Background(), 1)
		res, err := suite.s.Create(ctx, &proto.ShortURLCreateRequest{Url: "https://google.com"})
		suite.Require().NoError(err)
		suite.NotEmpty(res.Result)
		u1 := res.Result
		res, err = suite.s.Create(ctx, &proto.ShortURLCreateRequest{Url: "https://google.com"})
		suite.Require().NoError(err)
		suite.NotEmpty(res.Result)
		suite.Require().Equal(u1, res.Result)
	})
}

func (suite *ShortURLServiceSuite) TestCreateBatch() {
	suite.Run("unauthenticated", func() {
		_, err := suite.s.CreateBatch(context.Background(), &proto.ShortURLCreateBatchRequest{})
		suite.Require().Error(err)
		st, ok := status.FromError(err)
		suite.Require().True(ok)
		suite.Equal(codes.Unauthenticated, st.Code())
	})

	suite.Run("should batch create short url", func() {
		ctx := auth.ToContext(context.Background(), 1)
		items := []*proto.ShortURLCreateBatchRequest_Item{
			{
				CorrelationId: "100",
				OriginalUrl:   "https://google.com",
			},
			{
				CorrelationId: "200",
				OriginalUrl:   "https://facebook.com",
			},
		}
		res, err := suite.s.CreateBatch(ctx, &proto.ShortURLCreateBatchRequest{Items: items})
		suite.Require().NoError(err)
		suite.Len(res.Items, 2)
		suite.NotEmpty(res.Items[0].ShortUrl)
		suite.NotEmpty(res.Items[1].ShortUrl)
		suite.Equal("100", res.Items[0].CorrelationId)
		suite.Equal("200", res.Items[1].CorrelationId)
	})

}

func (suite *ShortURLServiceSuite) TestDeleteBatch() {
	suite.Run("unauthenticated", func() {
		_, err := suite.s.DeleteBatch(context.Background(), &proto.ShortURLDeleteBatchRequest{})
		suite.Require().Error(err)
		st, ok := status.FromError(err)
		suite.Require().True(ok)
		suite.Equal(codes.Unauthenticated, st.Code())
	})

	suite.Run("should batch delete short url", func() {
		ctx := auth.ToContext(context.Background(), 1)
		s1, err := suite.u.ShortURL.Create(ctx, 1, "https://google.com")
		suite.Require().NoError(err)
		s2, err := suite.u.ShortURL.Create(ctx, 1, "https://facebook.com")
		suite.Require().NoError(err)
		s3, err := suite.u.ShortURL.Create(ctx, 1, "https://twitter.com")
		suite.Require().NoError(err)

		items := &proto.ShortURLDeleteBatchRequest{
			Items: []string{s1.ID, s2.ID},
		}
		_, err = suite.s.DeleteBatch(ctx, items)
		suite.Require().NoError(err)

		time.Sleep(100 * time.Millisecond)
		res, err := suite.u.ShortURL.GetByUserID(ctx, 1)
		suite.Require().NoError(err)
		suite.Len(res, 1)
		suite.Equal(s3.OriginalURL, res[0].OriginalURL)
	})

	suite.Run("should not delete short url if not owner", func() {
		time.Sleep(100 * time.Millisecond)
		ctx := auth.ToContext(context.Background(), 1)
		s1, err := suite.u.ShortURL.Create(ctx, 1, "https://apple.com")
		suite.Require().NoError(err)
		s2, err := suite.u.ShortURL.Create(ctx, 2, "https://amazon.com")
		suite.Require().NoError(err)

		items := &proto.ShortURLDeleteBatchRequest{
			Items: []string{s1.ID, s2.ID},
		}
		_, err = suite.s.DeleteBatch(ctx, items)
		suite.Require().NoError(err)

		time.Sleep(100 * time.Millisecond)
		res, err := suite.u.ShortURL.GetByUserID(ctx, 1)
		suite.Require().NoError(err)
		suite.Len(res, 1)
		suite.Equal("https://twitter.com", res[0].OriginalURL)

		res, err = suite.u.ShortURL.GetByUserID(ctx, 2)
		suite.Require().NoError(err)
		suite.Len(res, 1) //
		suite.Equal("https://amazon.com", res[0].OriginalURL)

	})

	suite.Run("should return error if no urls provided", func() {
		ctx := auth.ToContext(context.Background(), 1)
		_, err := suite.s.DeleteBatch(ctx, &proto.ShortURLDeleteBatchRequest{})
		suite.Require().Error(err)
		st, ok := status.FromError(err)
		suite.Require().True(ok)
		suite.Equal(codes.InvalidArgument, st.Code())
	})

}

func (suite *ShortURLServiceSuite) TestGetByUserID() {

	suite.Run("unauthenticated", func() {
		_, err := suite.s.GetByUserID(context.Background(), &proto.ShortURLGetByUserIDRequest{})
		suite.Require().Error(err)
		st, ok := status.FromError(err)
		suite.Require().True(ok)
		suite.Equal(codes.Unauthenticated, st.Code())
	})

	suite.Run("should get short url by user id", func() {
		ctx := auth.ToContext(context.Background(), 1)
		_, err := suite.u.ShortURL.Create(ctx, 1, "https://google.com")
		suite.Require().NoError(err)
		_, err = suite.u.ShortURL.Create(ctx, 1, "https://facebook.com")
		suite.Require().NoError(err)

		res, err := suite.s.GetByUserID(ctx, &proto.ShortURLGetByUserIDRequest{})
		suite.Require().NoError(err)
		suite.Len(res.Items, 2)
	})

	suite.Run("should return no content if no short url found", func() {
		ctx := auth.ToContext(context.Background(), 2)
		_, err := suite.s.GetByUserID(ctx, &proto.ShortURLGetByUserIDRequest{})
		suite.Require().Error(err)
		st, ok := status.FromError(err)
		suite.Require().True(ok)
		suite.Equal(pkgerrors.GRPCNoContent, st.Code())
	})
}

func TestShortURLServiceSuite(t *testing.T) {
	suite.Run(t, new(ShortURLServiceSuite))
}
