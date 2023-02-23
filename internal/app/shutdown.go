package app

import (
	"context"
	"os"
	"os/signal"

	"github.com/rs/zerolog/log"
)

// ContextWithShutdown - возвращает контекст, который завершается при получении любого из сигналов sig.
func ContextWithShutdown(ctx context.Context, sig ...os.Signal) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, sig...)
		s := <-stop
		log.Warn().Msgf("Received signal: %s", s)
		cancel()
	}()
	return ctx, cancel
}
