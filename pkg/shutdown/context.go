package shutdown

import (
	"context"
	"os"
	"os/signal"
)

// ContextWithShutdown - возвращает контекст, который завершается при получении любого из сигналов sig.
func ContextWithShutdown(ctx context.Context, sig ...os.Signal) (context.Context, func()) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, sig...)
		<-stop
		cancel()
	}()
	return ctx, cancel
}
