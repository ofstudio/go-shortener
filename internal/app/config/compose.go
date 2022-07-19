package config

import "log"

type CfgFunc func(*Config) (*Config, error)

// Compose - объединяет набор конфигурационных функций в одну
func Compose(funcs ...CfgFunc) (*Config, error) {
	var err error
	cfg := &Config{}
	for _, fn := range funcs {
		cfg, err = fn(cfg)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// MustCompose - объединяет набор конфигурационных функций в одну
// В случае ошибки приложение завершается с ошибкой.
func MustCompose(funcs ...CfgFunc) *Config {
	cfg, err := Compose(funcs...)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}
