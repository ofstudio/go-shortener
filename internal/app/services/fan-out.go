package services

import (
	"context"
	"runtime"
)

var batchWorkerCount = runtime.NumCPU()

// fanOut - демультиплексор каналов.
// Распределяет значения из inputCh в несколько выходных каналов (round-robin).
// Количество выходных каналов задается batchWorkerCount.
// Если контекст завершится раньше, чем закончатся значения во входном канале,
// то закроет все выходные каналы и завершит выполнение.
func fanOut(ctx context.Context, inputCh chan string) []chan string {
	// Создаем слайс выходных каналов
	outChans := make([]chan string, batchWorkerCount)
	for i := 0; i < batchWorkerCount; i++ {
		outChans[i] = make(chan string)
	}

	// Горутина, которая считывает значения из входного канала и передает их в каналы поочередно (round-robin).
	go func() {
		// Закрываем все выходные каналы при выходе из горутины
		defer func(outChans []chan string) {
			for _, ch := range outChans {
				close(ch)
			}
		}(outChans)

		// Считываем значения из входного канала
		for i := 0; ; i++ {
			select {
			// Если контекст завершился, завершаем горутину
			case <-ctx.Done():
				return
			case item, ok := <-inputCh:
				// Если входной канал закрыт, завершаем горутину
				if !ok {
					return
				}
				// Если канал не закрыт, отправляем значение в очередной выходной канал
				n := i % batchWorkerCount
				outChans[n] <- item
			}
		}
	}()

	return outChans
}
