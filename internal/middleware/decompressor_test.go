package middleware_test

import (
	"bytes"
	"compress/gzip"
	"github.com/go-chi/chi/v5"
	"github.com/ofstudio/go-shortener/internal/middleware"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"io/ioutil"
	"net/http"
)

// ginkgo — без комментариев! ©
var _ = Describe("Decompressor Middleware", func() {
	server := &ghttp.Server{}
	plainRequest := []byte("Hello, World!")
	compressedRequest := testGzipCompress(plainRequest)

	BeforeEach(func() {
		// создаем echo-сервер
		server = ghttp.NewServer()
		r := chi.NewRouter()
		r.Use(middleware.Decompressor)
		// echo handler
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			body, err := ioutil.ReadAll(r.Body)
			Expect(err).ShouldNot(HaveOccurred())
			defer Expect(r.Body.Close()).Should(Succeed())
			Expect(w.Write(body)).Error().ShouldNot(HaveOccurred())
		})
		server.AppendHandlers(r.ServeHTTP)
	})

	AfterEach(func() {
		server.Close()
	})

	When("Content-Encoding set to gzip and body is compressed", func() {
		It("should return decompressed body", func() {
			body := testDecompressorRequest(server.URL(), compressedRequest, true)
			Expect(body).Should(Equal(plainRequest))
		})
	})

	When("Content-Encoding not set isn't compressed", func() {
		It("should return body as is", func() {
			body := testDecompressorRequest(server.URL(), plainRequest, false)
			Expect(body).Should(Equal(plainRequest))
		})
	})
})

// testGzipCompress - сжимает данные
func testGzipCompress(data []byte) []byte {
	buffer := bytes.NewBuffer(nil)
	gzipWriter := gzip.NewWriter(buffer)
	Expect(gzipWriter.Write(data)).Error().ShouldNot(HaveOccurred())
	Expect(gzipWriter.Close()).Should(Succeed())
	return buffer.Bytes()
}

// testDecompressorRequest - тестовый POST-запрос.
// Возвращает body ответа.
func testDecompressorRequest(u string, body []byte, compress bool) []byte {
	req, err := http.NewRequest("POST", u, bytes.NewBuffer(body))
	Expect(err).ShouldNot(HaveOccurred())
	if compress {
		req.Header.Set("Content-Encoding", "gzip")
	}
	res, err := http.DefaultClient.Do(req)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(res.StatusCode).Should(Equal(http.StatusOK))
	body, err = ioutil.ReadAll(res.Body)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(res.Body.Close()).Should(Succeed())
	return body
}
