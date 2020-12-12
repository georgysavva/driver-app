FROM golang:1.15.5 as build

WORKDIR /go/src/app

# Cache Go dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# This is needed to prove that the Go standard library gets recompiled.
# Since in order to recompile it requires gcc.
RUN apt-get remove -y gcc
RUN go test ./...