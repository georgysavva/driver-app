FROM golang:1.14.11

ADD . /go/src/app

WORKDIR /go/src/app/gateway

RUN go install ./...

ENTRYPOINT ["/go/bin/gateway-server"]
