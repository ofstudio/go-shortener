package handlers

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/ofstudio/go-shortener/internal/config"
	"github.com/ofstudio/go-shortener/internal/providers/auth"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/repo"
	"github.com/ofstudio/go-shortener/internal/usecases"
)

var _ = Describe("shortURL handlers", func() {
	server := &ghttp.Server{}
	cfg, _ := config.Default(nil)
	repository := repo.NewMemoryRepo()
	u := usecases.NewContainer(context.Background(), cfg, repository)
	shortURLPath := ""

	BeforeEach(func() {
		server = ghttp.NewServer()
		cfg.BaseURL = testParseURL(server.URL() + "/")
		r := chi.NewRouter()
		r.Use(auth.NewSHA256Provider(cfg, u.User).Handler)
		r.Mount("/", NewHTTPHandlers(u).Routes())
		server.AppendHandlers(r.ServeHTTP, r.ServeHTTP)
	})

	AfterEach(func() {
		server.Close()
	})

	When("successfully create and retrieve short url", func() {
		It("successfully create short url", func() {
			res := testHTTPRequest("POST", server.URL()+"/", "", "https://www.google.com")
			Expect(res.StatusCode).Should(Equal(http.StatusCreated))
			resBody, err := io.ReadAll(res.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Body.Close()).Error().ShouldNot(HaveOccurred())
			Expect(string(resBody)).ShouldNot(BeEmpty())
			shortURLPath = testParseURL(string(resBody)).Path
		})
		It("successfully retrieve short url", func() {
			res := testHTTPRequest("GET", server.URL()+shortURLPath, "", "")
			Expect(res.StatusCode).Should(Equal(http.StatusTemporaryRedirect))
			Expect(res.Header.Get("Location")).Should(Equal("https://www.google.com"))

		})
	})

	When("invalid post endpoint", func() {
		It("returns 405", func() {
			res := testHTTPRequest("POST", server.URL()+"/invalid", "", "https://www.google.com")
			Expect(res.StatusCode).Should(Equal(http.StatusMethodNotAllowed))
		})
	})

	When("original url is too long", func() {
		It("returns 401 error", func() {
			res := testHTTPRequest("POST", server.URL()+"/", "", "https://ya.ru/"+strings.Repeat("a", models.URLMaxLen))
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})

	When("original url is invalid", func() {
		It("returns 401 error", func() {
			res := testHTTPRequest("POST", server.URL()+"/", "", "invalid url")
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})

	When("original url is not http", func() {
		It("returns 401 error", func() {
			res := testHTTPRequest("POST", server.URL()+"/", "", "ftp://www.google.com")
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})

	When("original url is not specified", func() {
		It("returns 401 error", func() {
			res := testHTTPRequest("POST", server.URL()+"/", "", "")
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})

	When("short url id not exists", func() {
		It("returns 404 error", func() {
			res := testHTTPRequest("GET", server.URL()+"/123", "", "")
			Expect(res.StatusCode).Should(Equal(http.StatusNotFound))
		})
	})

	When("duplicate url sent", func() {
		It("returns 409 error", func() {
			res1 := testHTTPRequest("POST", server.URL()+"/", "", "https://www.duplicate.com")
			Expect(res1.StatusCode).Should(Equal(http.StatusCreated))
			resBody1, err := io.ReadAll(res1.Body)
			Expect(err).ShouldNot(HaveOccurred())
			_ = res1.Body.Close()
			res2 := testHTTPRequest("POST", server.URL()+"/", "", "https://www.duplicate.com")
			Expect(res2.StatusCode).Should(Equal(http.StatusConflict))
			resBody2, err := io.ReadAll(res2.Body)
			Expect(err).ShouldNot(HaveOccurred())
			_ = res2.Body.Close()
			Expect(string(resBody1)).Should(Equal(string(resBody2)))
		})
	})

})

var _ = Describe("ping handler", func() {
	When("no sql database configured", func() {
		It("returns 200", func() {
		})
	})
	When("sql database configured and connected", func() {
		It("returns 200", func() {
		})
	})
	When("sql database configured but not connected", func() {
		It("returns 500", func() {
		})
	})
})

func testHTTPRequest(method, u, contentType, body string, cookies ...*http.Cookie) *http.Response {
	// HTTP клиент, который не переходит по редиректам
	// https://stackoverflow.com/a/38150816
	c := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(method, u, strings.NewReader(body))
	Expect(err).ShouldNot(HaveOccurred())
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for _, cookie := range cookies {
		if cookie != nil {
			req.AddCookie(cookie)
		}
	}
	res, err := c.Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	return res
}

func testParseURL(rawURL string) url.URL {
	u, err := url.ParseRequestURI(rawURL)
	Expect(err).ShouldNot(HaveOccurred())
	return *u
}
