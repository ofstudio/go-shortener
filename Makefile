.PHONY: run, build, test

run:
	go run ./cmd/shortener/main.go

build:
	go build -o ./cmd/shortener/shortener ./cmd/shortener/main.go

test:
	go test -v ./...
