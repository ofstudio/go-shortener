package config

// CfgFunc - конфигурационная функция для Compose.
// Принимает на вход текущую конфигурацию и возвращает новую конфигурацию.
type CfgFunc func(*Config) (*Config, error)

// Compose - объединяет набор конфигурационных функций в итоговую конфигурацию
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
