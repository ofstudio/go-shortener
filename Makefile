.PHONY: pb, run, build, test, swag, bench, pprof, pprof_diff, lint

pb:
	protoc --go_out=. --go-grpc_out=. ./api/*.proto

run: pb
	go run ./cmd/shortener/main.go

build: pb
	go build \
	-ldflags " -s -w \
		-X main.buildVersion=v1.1.0 \
		-X 'main.buildDate=$$(date +'%Y/%m/%d %H:%M:%S')' \
		-X main.buildCommit=$$(git log -1 --pretty=format:%h)" \
	-o ./cmd/shortener/shortener ./cmd/shortener/main.go

test: pb
	go test -v ./...

swag:
	swag init -d ./internal/http/handlers --g api.go -o ./api/ && rm ./api/docs.go api/swagger.json

bench:
	go test -bench=. ./internal/repo -benchmem -memprofile ./profiles/${p}.pprof

pprof:
	go tool pprof  -http=":9090" repo.test profiles/${p}.pprof

pprof_diff:
	go tool pprof -top -diff_base=profiles/${p1}.pprof profiles/${p2}.pprof

lint:
	go build -o ./cmd/staticlint/staticlint ./cmd/staticlint/main.go && \
	./cmd/staticlint/staticlint ./cmd/... ./internal/... ./pkg/...

