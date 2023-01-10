package handlers

import (
	"net/http"
	"net/http/cookiejar"
	"strings"
)

func Example() {
	host := "http://localhost:8080"
	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: cookieJar,
	}

	// POST /api/shorten
	json := `{"url":"https://very.long.url/a/b/c/d/"}`
	res, _ := client.Post(host+"/api/shorten", "application/json", strings.NewReader(json))
	_ = res.Body.Close()
	// Response:
	// HTTP/1.1 200 OK
	// Content-Type: application/json
	//
	// {"result":"http://localhost:8080/abcd1234"}

	// POST /shorten/batch
	json = `
	[
		{
			"correlation_id": "1",
			"original_url": "https://a.very.long.url/a/b/c/d/"
		},
		{
			"correlation_id": "2",
			"original_url": "https://another.very.long.url/a/b/c/d/"
		}
	]`
	res, _ = client.Post(host+"/shorten/batch", "application/json", strings.NewReader(json))
	_ = res.Body.Close()
	// Response:
	// HTTP/1.1 201 Created
	// Content-Type: application/json
	//
	// [
	// 	{
	// 		"correlation_id": "1",
	// 		"short_url": "http://localhost:8080/abcd1235"
	// 	},
	// 	{
	// 		"correlation_id": "2",
	// 		"short_url": "http://localhost:8080/abcd1236"
	// 	}
	// ]

	// DELETE /user/urls
	json = `["abcd1234", "abcd1235"]`
	req, _ := http.NewRequest(http.MethodDelete, host+"/user/urls", strings.NewReader(json))
	req.Header.Set("Content-Type", "application/json")
	res, _ = client.Do(req)
	_ = res.Body.Close()
	// Response:
	// HTTP/1.1 202 Accepted

	// GET /user/urls
	res, _ = client.Get(host + "/user/urls")
	_ = res.Body.Close()
	// Response:
	// HTTP/1.1 200 OK
	// Content-Type: application/json
	//
	// [
	// 	{
	// 		"short_url": "http://localhost:8080/abcd1236",
	// 		"original_url": "https://another.very.long.url/a/b/c/d/"
	// 	}
	// ]
}
