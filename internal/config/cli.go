package config

import (
	"flag"
)

// FromCLI - конфигурационная функция, которая считывает конфигурацию приложения из переменных окружения.
//
// Флаги командной строки:
//
//	 -c <path>      - путь к файлу конфигурации
//		-a <host:port> - адрес для запуска HTTP-сервера
//		-g <host:port> - адрес для запуска gRPC-сервера
//		-b <url>       - базовый адрес сокращённого URL
//		-s             - использовать HTTPS с самоподписанным сертификатом
//		-f <path>      - файл для хранения данных
//		-t <cidr>      - подсеть, из которой разрешено обращение к внутреннему API
//		-d <dsn>       - строка с адресом подключения к БД
//
// Если какие-либо значения не заданы в командной строке, то используются значения переданные в cfg.
func FromCLI(args ...string) CfgFunc {
	return func(cfg *Config) (*Config, error) {
		f := flagSet(cfg)
		if err := f.Parse(args); err != nil {
			return nil, err
		}
		if err := cfg.validate(); err != nil {
			return nil, err
		}
		return cfg, nil
	}
}

// flagSet - возвращает FlagSet для чтения конфигурации из командной строки.
func flagSet(cfg *Config) *flag.FlagSet {
	f := flag.NewFlagSet("shortener", flag.ExitOnError)

	f.StringVar(&cfg.configFName, "c", cfg.configFName, "config file path")
	f.StringVar(&cfg.HTTPServerAddress, "a", cfg.HTTPServerAddress, "HTTP server address")
	f.StringVar(&cfg.GRPCServerAddress, "g", cfg.GRPCServerAddress, "gRPC server address")
	f.BoolVar(&cfg.EnableHTTPS, "s", cfg.EnableHTTPS, "Use HTTPS with self-signed certificate")
	f.Func("b", "Base URL", urlParseFunc(&cfg.BaseURL))
	f.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "File storage path")
	f.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	f.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "Trusted IP subnet for internal API access")
	return f
}
