package config

import (
	"encoding/json"
	"net/url"
	"os"
)

// jsonDTO - структура для считывания конфигурации из JSON-файла.
type jsonDTO struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// FromJSONFile - конфигурационная функция, которая считывает конфигурацию приложения из JSON-файла.
//
// Формат файла:
//
//	{
//		"server_address": "localhost:8080",
//		"base_url": "http://localhost",
//		"file_storage_path": "/path/to/file.db",
//		"database_dsn": "",
//		"enable_https": true
//	}
//
// Имя файла конфигурации можно задать (в порядке приоритета):
//  1. через переменную окружения CONFIG
//  2. через флаг -c командной строки
func FromJSONFile(args []string) CfgFunc {
	return func(cfg *Config) (*Config, error) {
		// Считываем предварительную конфигурацию из переменных окружения…
		preCfg := &Config{
			configFName: os.Getenv("CONFIG"),
		}

		// …и из флагов командной строки.
		f := flagSet(preCfg)
		if err := f.Parse(args); err != nil {
			return nil, err
		}

		// Если имя файла конфигурации задано, то считываем его.
		if preCfg.configFName != "" {
			jsonFile, err := os.Open(preCfg.configFName)
			if err != nil {
				return nil, err
			}
			//goland:noinspection GoUnhandledErrorResult
			defer jsonFile.Close()

			// Декодируем JSON-файл во временную dto-структуру.
			d := json.NewDecoder(jsonFile)
			d.DisallowUnknownFields() // запрещаем неизвестные поля, чтобы предотвратить опечатки
			dto := &jsonDTO{}
			if err := d.Decode(dto); err != nil {
				return nil, err
			}

			// Перезаписываем поля из dto-структуры в конфигурацию
			if dto.ServerAddress != "" {
				cfg.ServerAddress = dto.ServerAddress
			}
			if dto.BaseURL != "" {
				if u, err := url.Parse(dto.BaseURL); err != nil {
					return nil, err
				} else {
					cfg.BaseURL = *u
				}
			}
			if dto.FileStoragePath != "" {
				cfg.FileStoragePath = dto.FileStoragePath
			}
			if dto.DatabaseDSN != "" {
				cfg.DatabaseDSN = dto.DatabaseDSN
			}
			if dto.EnableHTTPS {
				cfg.EnableHTTPS = dto.EnableHTTPS
			}

			// Проверяем конфигурацию.
			if err := cfg.validate(); err != nil {
				return nil, err
			}
		}

		return cfg, nil
	}
}
