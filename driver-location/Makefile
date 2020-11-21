.PHONY: all test

all:
	docker build -t driver-location .

test:
	go test -v -race ./...
