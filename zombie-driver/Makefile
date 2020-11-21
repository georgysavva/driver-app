.PHONY: all test

all:
	docker build -t zombie-driver .

test:
	go test -v -race ./...
