FROM golang:1.14.11

ADD . /go/src/app

WORKDIR /go/src/app

RUN go install ./...

ENTRYPOINT ["/go/bin/driver-location-server"]
