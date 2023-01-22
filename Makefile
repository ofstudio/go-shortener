.PHONY: run, build, test, swag

run:
	go run ./cmd/shortener/main.go

build:
	go build -o ./cmd/shortener/shortener ./cmd/shortener/main.go

test:
	go test -v ./...

swag:
	swag init -d ./internal/handlers/ --g api.go -o ./api/ && rm ./api/docs.go api/swagger.json
