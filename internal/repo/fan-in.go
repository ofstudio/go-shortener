package repo

import (
	"context"
	"sync"
)

// fanIn - мультиплексор каналов.
// Собирает значения из всех каналов chans в один канал.
// Если контекст завершится раньше, чем все каналы закончатся,
// то закроет выходной канал и завершит выполнение.
func fanIn(ctx context.Context, chans ...chan string) <-chan string {
	outCh := make(chan string)

	go func() {
		wg := &sync.WaitGroup{}

		// Запускаем по одной "канальной" горутине для каждого из входного канала.
		for _, inputCh := range chans {
			wg.Add(1)
			go func(inputCh <-chan string) {
				defer wg.Done()
				for {
					select {
					// Если контекст завершился, завершаем горутину
					case <-ctx.Done():
						return
					case item, ok := <-inputCh:
						// Если канал закрыт, завершаем горутину
						if !ok {
							return
						}
						// Если канал не закрыт, отправляем значение в выходной канал
						outCh <- item
					}
				}
			}(inputCh)
		}

		// Ждем завершения всех "канальных" горутин и закрываем выходной канал
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
