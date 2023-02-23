package app

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"

	"github.com/ofstudio/go-shortener/internal/config"
	"github.com/ofstudio/go-shortener/internal/providers"
	"github.com/ofstudio/go-shortener/internal/providers/auth"
	"github.com/ofstudio/go-shortener/internal/providers/ipcheck"
	"github.com/ofstudio/go-shortener/internal/providers/tlsconf"
	"github.com/ofstudio/go-shortener/internal/repo"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

// App - приложение.
type App struct {
	cfg *config.Config
}

// NewApp - конструктор App.
func NewApp(cfg *config.Config) *App {
	return &App{cfg: cfg}
}

// Start - запускает приложение.
func (a *App) Start(ctx context.Context) error {
	// Создаём репозиторий
	repository, err := repo.Fabric(a.cfg)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}
	//goland:noinspection GoUnhandledErrorResult
	defer repository.Close()

	// Создаем юзкейсы
	u := usecases.NewContainer(ctx, a.cfg, repository)

	// Создаём провайдеры
	p := &providers.Container{
		Auth:    auth.NewSHA256Provider(a.cfg, u.User),
		IPCheck: ipcheck.NewWhitelist(a.cfg.TrustedSubnet),
		TLSConf: tlsconf.NewSelfSignedProvider(a.cfg.Cert),
	}

	// Создаём серверы
	httpServer := NewHTTPServer(a.cfg, u, p)
	grpcServer := NewGRPCServer(a.cfg, u, p)

	// Запускаем серверы.
	// Агрегатор errgroup доложен запускаться обязательно с контекстом.
	// Иначе, при ошибке запуска одного из серверов, другие не остановятся.
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return httpServer.Start(ctx) })
	g.Go(func() error { return grpcServer.Start(ctx) })

	return g.Wait()
}
