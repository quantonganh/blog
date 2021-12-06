lint:
	golangci-lint run -v ./...

test:
	go test -v ./...

build:
	CGO_ENABLED=0 go build -v -ldflags="-s -w" -o blog cmd/blog/main.go