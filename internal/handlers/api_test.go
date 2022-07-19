package handlers_test

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/handlers"
	"github.com/ofstudio/go-shortener/internal/middleware"
	"github.com/ofstudio/go-shortener/internal/models"
	"github.com/ofstudio/go-shortener/internal/repo"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var _ = Describe("POST /shorten ", func() {
	var server *ghttp.Server
	cfg := &config.DefaultConfig
	repository := repo.NewMemoryRepo()
	srv := services.NewContainer(cfg, repository)

	BeforeEach(func() {
		server = ghttp.NewServer()
		cfg.BaseURL = testParseURL(server.URL() + "/")
		r := chi.NewRouter()
		r.Use(middleware.NewAuthCookie(srv).Handler)
		r.Mount("/", handlers.NewAPIHandlers(srv).Routes())
		server.AppendHandlers(r.ServeHTTP, r.ServeHTTP)
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
			u := "https://www.google.com/" + strings.Repeat("a", models.URLMaxLen)
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
	When("duplicate url sent", func() {
		It("should return 409", func() {
			res1 := testHTTPRequest("POST", server.URL()+"/shorten", "application/json", `{"url":"https://www.duplicate.com"}`)
			Expect(res1.StatusCode).Should(Equal(http.StatusCreated))
			resBody1, err := ioutil.ReadAll(res1.Body)
			Expect(err).ShouldNot(HaveOccurred())
			_ = res1.Body.Close()

			res2 := testHTTPRequest("POST", server.URL()+"/shorten", "application/json", `{"url":"https://www.duplicate.com"}`)
			Expect(res2.StatusCode).Should(Equal(http.StatusConflict))
			resBody2, err := ioutil.ReadAll(res2.Body)
			Expect(err).ShouldNot(HaveOccurred())
			_ = res2.Body.Close()
			Expect(resBody1).Should(MatchJSON(resBody2))
		})
	})
})

var _ = Describe("POST /shorten/batch", func() {
	var server *ghttp.Server
	cfg := &config.DefaultConfig
	repository := repo.NewMemoryRepo()
	srv := services.NewContainer(cfg, repository)

	BeforeEach(func() {
		server = ghttp.NewServer()
		cfg.BaseURL = testParseURL(server.URL() + "/")
		r := chi.NewRouter()
		r.Use(middleware.NewAuthCookie(srv).Handler)
		r.Mount("/", handlers.NewAPIHandlers(srv).Routes())
		server.AppendHandlers(r.ServeHTTP)
	})
	AfterEach(func() {
		server.Close()
	})

	When("valid json sent", func() {
		It("should successfully create short urls", func() {
			body := `[
				{"correlation_id":"1","original_url":"https://www.google.com"},
				{"correlation_id":"2","original_url":"https://www.amazon.com"},
				{"correlation_id":"3","original_url":"https://www.facebook.com"}
			]`
			res := testHTTPRequest("POST", server.URL()+"/shorten/batch", "application/json", body)
			Expect(res.StatusCode).Should(Equal(http.StatusCreated))
			Expect(res.Header.Get("Content-Type")).Should(Equal("application/json"))
			resBody, err := ioutil.ReadAll(res.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Body.Close()).Error().ShouldNot(HaveOccurred())
			resJSON := make([]struct {
				CorrelationID string `json:"correlation_id"`
				ShortURL      string `json:"short_url"`
			}, 0)
			Expect(json.Unmarshal(resBody, &resJSON)).Should(Succeed())
			Expect(resJSON).Should(HaveLen(3))
			Expect(resJSON[0].CorrelationID).Should(Equal("1"))
			Expect(resJSON[0].ShortURL).ShouldNot(BeEmpty())
			Expect(resJSON[1].CorrelationID).Should(Equal("2"))
			Expect(resJSON[1].ShortURL).ShouldNot(BeEmpty())
			Expect(resJSON[2].CorrelationID).Should(Equal("3"))
			Expect(resJSON[2].ShortURL).ShouldNot(BeEmpty())
		})
	})
})

var _ = Describe("GET /user/urls", func() {
	var server *ghttp.Server
	cfg := &config.DefaultConfig
	repository := repo.NewMemoryRepo()
	srv := services.NewContainer(cfg, repository)
	var cookie *http.Cookie

	BeforeEach(func() {
		server = ghttp.NewServer()
		cfg.BaseURL = testParseURL(server.URL() + "/")
		r := chi.NewRouter()
		r.Use(middleware.NewAuthCookie(srv).Handler)
		r.Mount("/", handlers.NewAPIHandlers(srv).Routes())
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

var _ = Describe("DELETE /user/urls", func() {
	var server *ghttp.Server
	var cookie *http.Cookie
	cfg := &config.DefaultConfig
	repository := repo.NewMemoryRepo()
	srv := services.NewContainer(cfg, repository)

	BeforeEach(func() {
		server = ghttp.NewServer()
		cfg.BaseURL = testParseURL(server.URL() + "/")
		r := chi.NewRouter()
		r.Use(middleware.NewAuthCookie(srv).Handler)
		r.Mount("/", handlers.NewHTTPHandlers(srv).Routes())
		r.Mount("/api", handlers.NewAPIHandlers(srv).Routes())
		server.AppendHandlers(r.ServeHTTP)
	})

	AfterEach(func() {
		server.Close()
	})

	When("successful batch delete", func() {
		urls := []string{
			"https://www.google.com",
			"https://www.apple.com",
			"https://www.microsoft.com",
		}
		ids := make([]string, len(urls))
		for i, u := range urls {
			func(i int, u string) {
				It("should create 3 urls", func() {
					res := testHTTPRequest("POST", server.URL()+"/api/shorten", "application/json", fmt.Sprintf(`{"url":"%s"}`, u), cookie)
					Expect(res.StatusCode).Should(Equal(http.StatusCreated))
					if cookie == nil {
						cookie = res.Cookies()[0]
					}
					resBody, err := ioutil.ReadAll(res.Body)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(res.Body.Close()).Error().ShouldNot(HaveOccurred())
					resJSON := &struct {
						Result string `json:"result"`
					}{}
					Expect(json.Unmarshal(resBody, resJSON)).Should(Succeed())
					Expect(resJSON.Result).ShouldNot(BeEmpty())
					su, err := url.Parse(resJSON.Result)
					Expect(err).ShouldNot(HaveOccurred())
					ids[i] = su.Path[1:]
				})
			}(i, u)
		}
		It("should delete 2 urls", func() {
			body := fmt.Sprintf(`["%s", "%s"]`, ids[0], ids[1])
			res := testHTTPRequest("DELETE", server.URL()+"/api/user/urls", "application/json", body, cookie)
			Expect(res.StatusCode).Should(Equal(http.StatusAccepted))

		})
		It("should return Gone for deleted url", func() {
			res := testHTTPRequest("GET", server.URL()+"/"+ids[0], "", "", cookie)
			Expect(res.StatusCode).Should(Equal(http.StatusGone))
		})
		It("should return Gone for deleted url", func() {
			res := testHTTPRequest("GET", server.URL()+"/"+ids[1], "", "", cookie)
			Expect(res.StatusCode).Should(Equal(http.StatusGone))
		})
		It("should return list of remaining urls", func() {
			res := testHTTPRequest("GET", server.URL()+"/api/user/urls", "", "", cookie)
			Expect(res.StatusCode).Should(Equal(http.StatusOK))
			Expect(res.Header.Get("Content-Type")).Should(Equal("application/json"))
			resBody, err := ioutil.ReadAll(res.Body)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Body.Close()).Error().ShouldNot(HaveOccurred())
			Expect(resBody).Should(ContainSubstring(ids[2]))
		})
	})

})
