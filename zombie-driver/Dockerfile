FROM golang:1.14.11

ADD . /go/src/app

WORKDIR /go/src/app/zombie-driver

RUN go install ./...

ENTRYPOINT ["/go/bin/zombie-driver-server"]
