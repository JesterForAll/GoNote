.PHONY: all tidy lint test build

VERSION := $(shell echo "0.1.1")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

all: tidy lint test build

tidy:
	go mod tidy

lint: tidy
	golangci-lint run

test: lint
	go test ./... -race -cover

build: test
	go build -ldflags "-X github.com/JesterForAll/gonote/internal/version.Version=$(VERSION) -X github.com/JesterForAll/gonote/internal/version.BuildDate=$(BUILD_DATE)" -o bin/gonote/app ./cmd/gonote/

clean:
	rm -rf bin/
