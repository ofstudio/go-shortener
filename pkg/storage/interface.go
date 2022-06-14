package storage

// Interface - интерфейс KV-стораджа
type Interface interface {
	// Get - запрос значения из хранилища по ключу
	Get(key string) (string, error)
	// Set - сохранение данных в хранилище
	Set(key, val string) error
}
