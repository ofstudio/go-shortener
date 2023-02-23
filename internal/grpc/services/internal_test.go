package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ofstudio/go-shortener/api/proto"
	"github.com/ofstudio/go-shortener/internal/config"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/repo"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

type InternalServiceSuite struct {
	suite.Suite
	u *usecases.Container
	s *InternalService
}

func (suite *InternalServiceSuite) SetupTest() {
	cfg, _ := config.Default(nil)
	suite.u = usecases.NewContainer(context.Background(), cfg, repo.NewMemoryRepo())
	suite.s = NewInternalService(suite.u)
}

func (suite *InternalServiceSuite) TestStats() {

	suite.Run("should return stats", func() {
		suite.NoError(suite.u.User.Create(context.Background(), &models.User{}))
		_, err := suite.u.ShortURL.Create(context.Background(), 1, "https://google.com")
		suite.NoError(err)

		res, err := suite.s.Stats(context.Background(), &proto.StatsRequest{})
		suite.NoError(err)
		suite.Equal(uint32(1), res.Users)
		suite.Equal(uint32(1), res.Urls)
	})

}

func TestInternalServerSuite(t *testing.T) {
	suite.Run(t, new(InternalServiceSuite))
}
