.PHONY: all test

all:
	docker build -t driver-location -f Dockerfile ../.

test:
	go test -v -race ./...
