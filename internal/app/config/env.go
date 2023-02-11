package config

import "github.com/caarlos0/env/v6"

// FromEnv - конфигурационная функция, которая читывает конфигурацию приложения из переменных окружения.
//
// Переменные окружения:
//
//	SERVER_ADDRESS    - адрес для запуска HTTP-сервера
//	BASE_URL          - базовый адрес сокращённого URL
//	USE_TLS           - использовать TLS с самоподписанным сертификатом
//	FILE_STORAGE_PATH - файл для хранения данных
//	AUTH_TTL          - время жизни авторизационного токена
//	AUTH_SECRET       - секретный ключ для подписи авторизационного токена
//	TRUSTED_SUBNET   - подсеть, из которой разрешено обращение к внутреннему API
//
// Если какие-либо переменные окружения не заданы, то используются значения переданные в cfg.
func FromEnv(cfg *Config) (*Config, error) {
	// Получаем параметры из окружения
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	if err = cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}
