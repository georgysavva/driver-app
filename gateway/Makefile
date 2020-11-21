.PHONY: all test

all:
	docker build -t gateway .

test:
	go test -v -race ./...
