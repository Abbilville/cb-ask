.PHONY: all build test clean install

BINARY_NAME=oss-ask

all: test build

build:
	go build -o bin/$(BINARY_NAME) ./cmd/oss-ask

install:
	go install ./cmd/oss-ask

test:
	go test -v ./...

clean:
	rm -rf bin/ dist/
