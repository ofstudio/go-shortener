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
	return &Config{
		urlMaxLen: urlMaxLen,
		publicURL: normalizeURL(publicURL),
	}
}

// normalizeURL - нормализует URL: добавляем `/` в конце
func normalizeURL(url string) string {
	if url[len(url)-1:] != "/" {
		url += "/"
	}
	return url
}
