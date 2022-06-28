package handlers_test

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/handlers"
	"github.com/ofstudio/go-shortener/internal/middleware"
	"github.com/ofstudio/go-shortener/internal/repo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"io/ioutil"
	"net/http"
	"strings"
)

var _ = Describe("API POST /shorten ", func() {
	var server *ghttp.Server
	cfg := &config.DefaultConfig
	repository := repo.NewMemoryRepo()
	shortURLService := services.NewShortURLService(cfg, repository)
	userService := services.NewUserService(cfg, repository)

	BeforeEach(func() {
		server = ghttp.NewServer()
		cfg.BaseURL = testParseURL(server.URL() + "/")
		r := chi.NewRouter()
		r.Use(middleware.NewAuthCookie(userService).Handler)
		r.Mount("/", handlers.NewAPIHandlers(shortURLService).Routes())
		server.AppendHandlers(r.ServeHTTP)
	})

	AfterEach(func() {
		server.Close()
	})

	When("valid json sent", func() {
		It("should successfully create short url", func() {
			res := testHTTPRequest("POST", server.URL()+"/shorten", "application/json", `{"url":"https://www.google.com"}`)
			Expect(res.StatusCode).Should(Equal(http.StatusCreated))
			Expect(res.Header.Get("Content-Type")).Should(Equal("application/json"))
			resBody, err := ioutil.ReadAll(res.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Body.Close()).Error().ShouldNot(HaveOccurred())
			resJSON := &struct {
				Result string `json:"result"`
			}{}
			Expect(json.Unmarshal(resBody, resJSON)).Should(Succeed())
			Expect(resJSON.Result).ShouldNot(BeEmpty())
		})
	})
	When("too long url sent", func() {
		It("should return 400", func() {
			u := "https://www.google.com/" + strings.Repeat("a", cfg.URLMaxLen)
			res := testHTTPRequest("POST", server.URL()+"/shorten", "application/json", `{"url":"`+u+`"}`)
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})
	When("empty url sent", func() {
		It("should return 400", func() {
			res := testHTTPRequest("POST", server.URL()+"/shorten", "application/json", `{"url":""}`)
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})
	When("invalid url scheme sent", func() {
		It("should return 400", func() {
			res := testHTTPRequest("POST", server.URL()+"/shorten", "application/json", `{"url":"ftp://www.google.com"}`)
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})
	When("wrong json sent", func() {
		It("should return 400", func() {
			res := testHTTPRequest("POST", server.URL()+"/shorten", "application/json", `{"wrong":true}`)
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})
	When("malformed json sent", func() {
		It("should return 400", func() {
			res := testHTTPRequest("POST", server.URL()+"/shorten", "application/json", `}malformed{`)
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})
	When("empty body sent", func() {
		It("should return 400", func() {
			res := testHTTPRequest("POST", server.URL()+"/shorten", "application/json", ``)
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})
	When("wrong content type sent", func() {
		It("should return 400", func() {
			res := testHTTPRequest("POST", server.URL()+"/shorten", "application/xml", `{"url":"https://www.google.com"}`)
			Expect(res.StatusCode).Should(Equal(http.StatusBadRequest))
		})
	})
})

var _ = Describe("API GET /user/urls", func() {
	var server *ghttp.Server
	cfg := &config.DefaultConfig
	repository := repo.NewMemoryRepo()
	shortURLService := services.NewShortURLService(cfg, repository)
	userService := services.NewUserService(cfg, repository)
	var cookie *http.Cookie

	BeforeEach(func() {
		server = ghttp.NewServer()
		cfg.BaseURL = testParseURL(server.URL() + "/")
		r := chi.NewRouter()
		r.Use(middleware.NewAuthCookie(userService).Handler)
		r.Mount("/", handlers.NewAPIHandlers(shortURLService).Routes())
		server.AppendHandlers(r.ServeHTTP)
	})
	AfterEach(func() {
		server.Close()
	})

	When("cookie is provided", func() {
		It("should return cookie on first request", func() {
			res := testHTTPRequest("POST", server.URL()+"/shorten", "application/json", `{"url":"https://www.google.com"}`)
			Expect(res.StatusCode).Should(Equal(http.StatusCreated))
			Expect(res.Cookies()).ShouldNot(BeEmpty())
			cookie = res.Cookies()[0]
		})
		It("should accept cookie on second request", func() {
			res := testHTTPRequest("POST", server.URL()+"/shorten", "application/json", `{"url":"https://www.apple.com"}`, cookie)
			Expect(res.StatusCode).Should(Equal(http.StatusCreated))
		})
		It("should return list of urls", func() {
			res := testHTTPRequest("GET", server.URL()+"/user/urls", "", "", cookie)
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			Expect(res.Header.Get("Content-Type")).Should(Equal("application/json"))
			resBody, err := ioutil.ReadAll(res.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Body.Close()).Error().ShouldNot(HaveOccurred())
			var resJSON []struct {
				ShortURL    string `json:"short_url"`
				OriginalURL string `json:"original_url"`
			}
			Expect(json.Unmarshal(resBody, &resJSON)).Should(Succeed())
			Expect(resJSON).Should(HaveLen(2))
			Expect(resJSON[0].ShortURL).ShouldNot(BeEmpty())
			Expect(resJSON[0].OriginalURL).Should(Equal("https://www.google.com"))
			Expect(resJSON[1].ShortURL).ShouldNot(BeEmpty())
			Expect(resJSON[1].OriginalURL).Should(Equal("https://www.apple.com"))
		})
	})
	When("cookie is invalid", func() {
		It("should return 204", func() {
			res := testHTTPRequest("GET", server.URL()+"/user/urls", "", "", &http.Cookie{Name: "auth_token", Value: "invalid!invalid!invalid!invalid!"})
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
		})
	})
	When("cookie is not provided", func() {
		It("should return 204", func() {
			res := testHTTPRequest("GET", server.URL()+"/user/urls", "", "")
			Expect(res.StatusCode).Should(Equal(http.StatusNoContent))
		})
	})
})
