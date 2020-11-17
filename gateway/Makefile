.PHONY: all test

all:
	docker build -t gateway -f Dockerfile ../.

test:
	go test -v -race ./...
