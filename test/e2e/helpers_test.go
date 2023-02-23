package e2e

import (
	"net/http"
	"net/http/cookiejar"
)

func clientJar() *http.Client {
	cookieJar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar: cookieJar,
		// HTTP клиент, который не переходит по редиректам
		// https://stackoverflow.com/a/38150816
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
