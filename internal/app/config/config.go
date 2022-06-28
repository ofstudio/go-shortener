package config

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"log"
	"net/url"
	"os"
	"time"
)

// Config - конфигурация приложения
type Config struct {
	// URLMaxLen - максимальная длина URL в байтах.
	// Формально, размер URL ничем не ограничен.
	// Разные версии разных браузеров имеют свои конкретные ограничения: от 2048 байт до нескольких мегабайт.
	// В случае нашего сервиса необходимо некое разумное ограничение.
	URLMaxLen int

	// BaseURL - базовый адрес сокращённого URL.
	BaseURL url.URL `env:"BASE_URL"`

	// ServerAddress - адрес для запуска HTTP-сервера.
	ServerAddress string `env:"SERVER_ADDRESS"`

	// FileStoragePath - файл для хранения данных.
	// Если не задан, данные будут храниться в памяти.
	FileStoragePath string `env:"FILE_STORAGE_PATH"`

	// AuthTTL - время жизни авторизационного токена
	AuthTTL time.Duration `env:"AUTH_TTL"`

	// AuthSecret - секретный ключ для подписи авторизационного токена
	AuthSecret string `env:"AUTH_SECRET,unset"`

	// DatabaseDSN - строка с адресом подключения к БД
	DatabaseDSN string `env:"DATABASE_DSN"`
}

// DefaultConfig - конфигурация по умолчанию
var (
	DefaultConfig = Config{
		URLMaxLen:     4096,
		BaseURL:       mustParseRequestURI("http://localhost:8080/"),
		ServerAddress: "0.0.0.0:8080",
		AuthTTL:       time.Minute * 60 * 24 * 30,
		AuthSecret:    mustRandSecret(64),
		DatabaseDSN:   "",
	}
)

// MustNewFromEnvAndCLI - инициализирует конфигурацию приложения из переменных окружения и командной строки.
// В случае ошибки приложение завершается с ошибкой.
func MustNewFromEnvAndCLI() *Config {
	cfg, err := NewFromEnvAndCLI()
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

// NewFromEnvAndCLI - инициализирует конфигурацию приложения из переменных окружения и командной строки.
// Значения из командной строки перекрывают значения из окружения.
//
// Флаги командной строки:
//    -a <host:port> - адрес для запуска HTTP-сервера
//    -b <url>       - базовый адрес сокращённого URL
//    -f <path>      - файл для хранения данных
//    -t <duration>  - время жизни авторизационного токена
//	  -d <dsn>       - строка с адресом подключения к БД
//
// Если какие-либо значения не заданы ни в переменных окружения, ни в командной строке,
// то используются значения по умолчанию из DefaultConfig.
func NewFromEnvAndCLI() (*Config, error) {
	return newFromEnvAndCLI(os.Args[1:])
}

// newFromEnvAndCLI - логика для NewFromEnvAndCLI.
func newFromEnvAndCLI(arguments []string) (*Config, error) {

	cfg, err := NewFromEnv()
	if err != nil {
		return nil, err
	}

	// Парсим командную строку
	cli := flag.NewFlagSet("config", flag.ExitOnError)
	cli.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
	cli.Func("b", "Base URL", urlParseFunc(&cfg.BaseURL))
	cli.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path (default: in-memory)")
	cli.DurationVar(&cfg.AuthTTL, "t", cfg.AuthTTL, "Auth token TTL")
	cli.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	if err = cli.Parse(arguments); err != nil {
		return nil, err
	}

	return validate(cfg)
}

// NewFromEnv - инициализирует конфигурацию приложения из переменных окружения.
//
// Переменные окружения:
//    SERVER_ADDRESS    - адрес для запуска HTTP-сервера
//    BASE_URL          - базовый адрес сокращённого URL
//    FILE_STORAGE_PATH - файл для хранения данных
//    AUTH_TTL          - время жизни авторизационного токена
//	  AUTH_SECRET       - секретный ключ для подписи авторизационного токена
//
// Если какие-либо переменные окружения не заданы,
// используются значения по умолчанию из DefaultConfig.
func NewFromEnv() (*Config, error) {
	// Параметры по умолчанию
	cfg := DefaultConfig
	// Получаем параметры из окружения
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}
	return validate(&cfg)
}

// mustParseRequestURI - парсит URL.
// В случае ошибки приложение завершается с ошибкой.
func mustParseRequestURI(rawURL string) url.URL {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		log.Fatal(err)
	}
	return *u
}

// mustRandSecret - генерирует случайный ключ.
// В случае ошибки приложение завершается с ошибкой.
func mustRandSecret(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// urlParseFunc - функция для парсинга URL из флага
func urlParseFunc(value *url.URL) func(string) error {
	return func(rawURL string) error {
		if value == nil {
			return fmt.Errorf("url value is nil")
		}
		u, err := url.ParseRequestURI(rawURL)
		if err != nil {
			return err
		}
		*value = *u
		return nil
	}
}
