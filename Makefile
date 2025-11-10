.PHONY: build install test clean

build:
	go build -o bin/envault ./cmd/envault

install:
	go install ./cmd/envault

test:
	go test -v ./...

clean:
	rm -rf bin/

lint:
	golangci-lint run

fmt:
	go fmt ./...
