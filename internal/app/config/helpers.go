package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
)

// randSecret - генерирует случайный ключ.
func randSecret(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// urlParseFunc - функция для парсинга URL из флага
func urlParseFunc(value *url.URL) func(string) error {
	return func(rawURL string) error {
		if value == nil {
			return fmt.Errorf("url value is nil")
		}
		u, err := url.Parse(rawURL)
		if err != nil {
			return err
		}
		*value = *u
		return nil
	}
}
