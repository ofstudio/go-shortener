package app

// Config - конфигурация приложения
type Config struct {
	// Максимальная длина URL в _байтах_
	// Формально, размер URL ничем не ограничен.
	// Разные версии разных браузеров имеют свои конкретные ограничения: от 2048 байт до мегабайт.
	// В случае нашего сервиса необходимо некое разумное ограничение
	urlMaxLen int

	// Публичный URL, по которому доступно приложение
	publicURL string
}

func NewConfig(urlMaxLen int, publicURL string) *Config {
	// Нормализуем publicURL - добавляем `/` в конце
	if publicURL[len(publicURL)-1:] != "/" {
		publicURL += "/"
	}
	return &Config{urlMaxLen: urlMaxLen, publicURL: publicURL}
}
