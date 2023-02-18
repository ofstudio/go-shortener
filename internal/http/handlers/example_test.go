package handlers

import (
	"net/http"
	"strings"
)

func Example() {
	host := "http://localhost:8080"

	// POST /api/shorten
	body := `{"url":"https://very.long.url/a/b/c/d/"}`
	res, err := http.DefaultClient.Post(host+"/api/shorten", "application/json", strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Response:
	// HTTP/1.1 200 OK
	// Content-Type: application/body
	//
	// {"result":"http://localhost:8080/abcd1234"}

}
