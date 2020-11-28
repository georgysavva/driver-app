# Build stage
FROM golang:1.15.5-alpine3.12 as build

WORKDIR /go/src/app

# Cache Go dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /go/bin/ ./...


# Final stage
FROM alpine:3.12 as final
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=build /go/bin/gateway-server /go/src/app/config.yaml ./

ENTRYPOINT ["./gateway-server"]
