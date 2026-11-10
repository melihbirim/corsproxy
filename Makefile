.PHONY: build test lint

build:
	go build -o bin/cors-proxy .

test:
	go test ./...

lint:
	go fmt ./...
	golangci-lint run
