.PHONY: all build test clean install

BINARY_NAME=cb-ask

all: test build

build:
	go build -o bin/$(BINARY_NAME) ./cmd/cb-ask

install:
	go install ./cmd/cb-ask

test:
	go test -v ./...

clean:
	rm -rf bin/ dist/
