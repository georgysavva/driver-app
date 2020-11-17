.PHONY: all test

all:
	docker build -t zombie-driver -f Dockerfile ../.

test:
	go test -v -race ./...
