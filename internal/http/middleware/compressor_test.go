package middleware

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Compressor Middleware", func() {
	server := &ghttp.Server{}
	plainResponse := []byte("Hello, World!")
	minSize := int64(len(plainResponse) / 2)
	compressedResponse := testGzipCompress(plainResponse)

	BeforeEach(func() {
		server = ghttp.NewServer()
		r := chi.NewRouter()
		r.Use(NewCompressor(minSize, gzip.DefaultCompression).AddType("text/html").Handler)

		// Ответ короче минимального размера
		r.Get("/short-response", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			Expect(w.Write(plainResponse[:minSize-1])).Error().ShouldNot(HaveOccurred())
		})

		// Ответ больше минимального размера
		r.Get("/long-response", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			Expect(w.Write(plainResponse)).Error().ShouldNot(HaveOccurred())
		})

		// Ответ больше минимального размера и отправляется тремя вызовами w.Write
		r.Get("/multi-write-response", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/html")
			Expect(w.Write(plainResponse[:minSize-1])).Error().ShouldNot(HaveOccurred())
			Expect(w.Write(plainResponse[minSize-1 : minSize+1])).Error().ShouldNot(HaveOccurred())
			Expect(w.Write(plainResponse[minSize+1:])).Error().ShouldNot(HaveOccurred())
		})

		// Ответ больше минимального размера, но не подходит Content-Type
		r.Get("/content-type-mismatch", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/plain")
			Expect(w.Write(plainResponse)).Error().ShouldNot(HaveOccurred())
		})
		server.AppendHandlers(r.ServeHTTP)
	})

	AfterEach(func() {
		server.Close()
	})

	When("Accept-Encoding: gzip", func() {
		Context("and response is short", func() {
			It("should not compress", func() {
				body, encoding := testCompressorRequest(server, "/short-response", true)
				Expect(body).Should(Equal(plainResponse[:minSize-1]))
				Expect(encoding).Should(Equal(""))
			})
		})

		Context("and response is long", func() {
			It("should compress", func() {
				body, encoding := testCompressorRequest(server, "/long-response", true)
				Expect(body).Should(Equal(compressedResponse))
				Expect(encoding).Should(Equal("gzip"))
			})
		})

		Context("and response is long and multi-write", func() {
			It("should compress", func() {
				body, encoding := testCompressorRequest(server, "/multi-write-response", true)
				Expect(body).Should(Equal(compressedResponse))
				Expect(encoding).Should(Equal("gzip"))
			})
		})

		Context("and response is long and content-type mismatch", func() {
			It("should not compress", func() {
				body, encoding := testCompressorRequest(server, "/content-type-mismatch", true)
				Expect(body).Should(Equal(plainResponse))
				Expect(encoding).Should(Equal(""))
			})
		})
	})

	When("Accept-Encoding is not gzip", func() {
		Context("and response is long", func() {
			It("should not compress", func() {
				body, encoding := testCompressorRequest(server, "/long-response", false)
				Expect(body).Should(Equal(plainResponse))
				Expect(encoding).Should(Equal(""))
			})
		})
		Context("and response is long and multi-write", func() {
			It("should not compress", func() {
				body, encoding := testCompressorRequest(server, "/multi-write-response", false)
				Expect(body).Should(Equal(plainResponse))
				Expect(encoding).Should(Equal(""))
			})
		})
	})

})

// testCompressorRequest - тестовый GET-запрос.
// Возвращает body ответа и заголовок Content-Encoding.
func testCompressorRequest(server *ghttp.Server, path string, acceptGzip bool) ([]byte, string) {
	req, err := http.NewRequest("GET", server.URL()+path, nil)
	Expect(err).ShouldNot(HaveOccurred())
	if acceptGzip {
		req.Header.Set("Accept-Encoding", "gzip")
	}
	// DefaultHTTPClient форсит Accept-Encoding: gzip
	c := http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	res, err := c.Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(res.StatusCode).Should(Equal(http.StatusOK))
	body, err := io.ReadAll(res.Body)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(res.Body.Close()).Should(Succeed())
	return body, res.Header.Get("Content-Encoding")
}
