FROM golang:1.14.11

ADD . /go/src/app

WORKDIR /go/src/app/driver-location

RUN go install ./...

ENTRYPOINT ["/go/bin/driver-location-server"]
