.PHONY: build install clean test

build:
	mkdir -p ./bin
	go build -o ./bin/llmcommit .

install:
	go install .

clean:
	rm -rf ./bin

test:
	go test ./...

.DEFAULT_GOAL := build
