package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
)

// mustParseRequestURI - парсит URL.
// В случае ошибки приложение завершается с ошибкой.
func mustParseRequestURI(rawURL string) url.URL {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		log.Fatal(err)
	}
	return *u
}

// randSecret - генерирует случайный ключ.
func randSecret(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// mustRandSecret - генерирует случайный ключ.
// В случае ошибки приложение завершается с ошибкой.
func mustRandSecret(n int) string {
	b, err := randSecret(n)
	if err != nil {
		log.Fatal(err)
	}
	return b
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
