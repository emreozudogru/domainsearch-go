BINARY := bin/domainsearch

.PHONY: all build test vet fmt tidy lint run clean

all: build

## build: compile the binary into ./bin
build:
	go build -o $(BINARY) ./cmd/domainsearch

## test: run unit tests
test:
	go test ./...

## vet: run go vet
vet:
	go vet ./...

## fmt: format and tidy
fmt:
	gofmt -l .
	go mod tidy

## lint: run golangci-lint
lint:
	golangci-lint run

## tidy: tidy go modules
tidy:
	go mod tidy

## run: run the tool from source
run:
	go run ./cmd/domainsearch

## clean: remove build artifacts
clean:
	rm -rf bin/
