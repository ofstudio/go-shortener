package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/ofstudio/go-shortener/internal/app/config"
	"github.com/ofstudio/go-shortener/internal/app/services"
	"github.com/ofstudio/go-shortener/internal/repo"
)

var _ = Describe("AuthCookie Middleware", func() {
	var cookie *http.Cookie
	server := &ghttp.Server{}
	cfg, _ := config.Default(nil)

	BeforeEach(func() {
		repository := repo.NewMemoryRepo()
		srv := services.NewContainer(cfg, repository)
		server = ghttp.NewServer()
		r := chi.NewRouter()
		r.Use(NewAuthCookie(srv).WithSecret([]byte(cfg.AuthSecret)).Handler)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			Expect(w.Write([]byte("Hello, World!"))).Error().ShouldNot(HaveOccurred())
		})
		server.AppendHandlers(r.ServeHTTP)
	})

	AfterEach(func() {
		server.Close()
	})

	When("user is not authorized", func() {

		It("should set cookie if cookie is not set", func() {
			res, err := http.Get(server.URL() + "/")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Cookies()).Should(HaveLen(1))
			cookie = res.Cookies()[0]
			Expect(cookie.Name).Should(Equal("auth_token"))
		})

		It("should not set cookie if valid cookie is set", func() {
			req, err := http.NewRequest("GET", server.URL()+"/", nil)
			Expect(err).ShouldNot(HaveOccurred())
			req.AddCookie(cookie)
			res, err := http.DefaultClient.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Cookies()).Should(HaveLen(0))
		})

		It("should set new cookie if invalid cookie is set", func() {
			req, err := http.NewRequest("GET", server.URL()+"/", nil)
			Expect(err).ShouldNot(HaveOccurred())
			cookie.Value = "invalid"
			req.AddCookie(cookie)
			res, err := http.DefaultClient.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(res.Cookies()).Should(HaveLen(1))
			newCookie := res.Cookies()[0]
			Expect(newCookie.Name).Should(Equal("auth_token"))
			Expect(newCookie.Value).ShouldNot(Equal(cookie.Value))
		})
	})

})
