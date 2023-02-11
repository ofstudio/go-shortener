package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Decompressor Middleware", func() {
	server := &ghttp.Server{}

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	When("no whitelist is set", func() {
		r := chi.NewRouter()
		r.Use(NewWhitelist().Handler)
		r.Get("/", testOkHandler)

		It("should return 403", func() {
			server.AppendHandlers(r.ServeHTTP)
			resp, err := http.Get(server.URL())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
		})
	})

	When("whitelist is set with invalid values", func() {
		r := chi.NewRouter()
		r.Use(NewWhitelist("asdft", "400.300.200.100", "1.2.3.4/999").Handler)
		r.Get("/", testOkHandler)

		It("should return 403", func() {
			server.AppendHandlers(r.ServeHTTP)
			resp, err := http.Get(server.URL())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
		})
	})

	When("whitelist is set with ips", func() {
		r := chi.NewRouter()
		r.Use(NewWhitelist("192.168.0.1", "10.10.12.128").Handler)
		r.Get("/", testOkHandler)

		It("should return 200 with allowed X-Real-IP", func() {
			for _, ip := range []string{"192.168.0.1", "10.10.12.128"} {
				server.AppendHandlers(r.ServeHTTP)
				req, err := http.NewRequest(http.MethodGet, server.URL(), nil)
				Expect(err).ShouldNot(HaveOccurred())
				req.Header.Set("X-Real-IP", ip)
				resp, err := http.DefaultClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			}
		})

		It("should return 403 with X-Real-IP: 127.0.0.1", func() {
			server.AppendHandlers(r.ServeHTTP)
			req, err := http.NewRequest(http.MethodGet, server.URL(), nil)
			Expect(err).ShouldNot(HaveOccurred())
			req.Header.Set("X-Real-IP", "127.0.0.1")
			resp, err := http.DefaultClient.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
		})

		It("should return 403 without X-Real-IP", func() {
			server.AppendHandlers(r.ServeHTTP)
			resp, err := http.Get(server.URL())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
		})
	})

	When("whitelist is set with cidrs", func() {
		r := chi.NewRouter()
		// ranges:
		//	- 10.0.0.0 .. 10.0.0.255
		// 	- 192.168.0.0 .. 192.168.255.255
		r.Use(NewWhitelist("10.0.0.0/24", "192.168.0.0/16").Handler)
		r.Get("/", testOkHandler)

		It("should return 200 with allowed X-Real-IP", func() {
			for _, ip := range []string{"10.0.0.10", "192.168.0.1"} {
				server.AppendHandlers(r.ServeHTTP)
				req, err := http.NewRequest(http.MethodGet, server.URL(), nil)
				Expect(err).ShouldNot(HaveOccurred())
				req.Header.Set("X-Real-IP", ip)
				resp, err := http.DefaultClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).Should(Equal(http.StatusOK))
			}
		})

		It("should return 403 with forbidden X-Real-IP", func() {
			for _, ip := range []string{"10.10.10.10", "192.169.0.1"} {
				server.AppendHandlers(r.ServeHTTP)
				req, err := http.NewRequest(http.MethodGet, server.URL(), nil)
				Expect(err).ShouldNot(HaveOccurred())
				req.Header.Set("X-Real-IP", ip)
				resp, err := http.DefaultClient.Do(req)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
			}
		})

		It("should return 403 without X-Real-IP", func() {
			server.AppendHandlers(r.ServeHTTP)
			resp, err := http.Get(server.URL())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
		})
	})

	When("custom headers is set", func() {
		r := chi.NewRouter()
		r.Use(NewWhitelist("127.0.0.1").UseHeaders("X-Forwarded-For", "X-My-IP").Handler)
		r.Get("/", testOkHandler)

		It("should return 200 with allowed X-Forwarded-For", func() {
			server.AppendHandlers(r.ServeHTTP)
			req, err := http.NewRequest(http.MethodGet, server.URL(), nil)
			Expect(err).ShouldNot(HaveOccurred())
			req.Header.Set("X-Forwarded-For", "127.0.0.1")
			resp, err := http.DefaultClient.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		})

		It("should return 200 with allowed X-My-IP", func() {
			server.AppendHandlers(r.ServeHTTP)
			req, err := http.NewRequest(http.MethodGet, server.URL(), nil)
			Expect(err).ShouldNot(HaveOccurred())
			req.Header.Set("X-My-IP", "127.0.0.1")
			resp, err := http.DefaultClient.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusOK))
		})

		It("should return 403 with forbidden X-Forwarded-For", func() {
			server.AppendHandlers(r.ServeHTTP)
			req, err := http.NewRequest(http.MethodGet, server.URL(), nil)
			Expect(err).ShouldNot(HaveOccurred())
			req.Header.Set("X-Forwarded-For", "127.0.0.2")
			resp, err := http.DefaultClient.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
		})

		It("should return 403 with forbidden X-My-IP", func() {
			server.AppendHandlers(r.ServeHTTP)
			req, err := http.NewRequest(http.MethodGet, server.URL(), nil)
			Expect(err).ShouldNot(HaveOccurred())
			req.Header.Set("X-My-IP", "127.0.0.2")
			resp, err := http.DefaultClient.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
		})

		It("should return 403 with X-Real-IP", func() {
			server.AppendHandlers(r.ServeHTTP)
			req, err := http.NewRequest(http.MethodGet, server.URL(), nil)
			Expect(err).ShouldNot(HaveOccurred())
			req.Header.Set("X-Real-IP", "127.0.0.1")
			resp, err := http.DefaultClient.Do(req)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
		})

		It("should return 403 without header", func() {
			server.AppendHandlers(r.ServeHTTP)
			resp, err := http.Get(server.URL())
			Expect(err).ShouldNot(HaveOccurred())
			Expect(resp.StatusCode).Should(Equal(http.StatusForbidden))
		})
	})
})

func testOkHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
