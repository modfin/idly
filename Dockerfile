FROM golang:1.23-alpine3.20 AS builder

WORKDIR /src

COPY . /src

RUN GOOS=linux GOARCH=amd64 go build -o /idly /src/cmd/idlyd/main.go

# -----

FROM alpine:3.20

EXPOSE 8080

RUN apk add --no-cache tzdata ca-certificates
COPY --from=builder /idly /idly
VOLUME /var/lib/idly
ENTRYPOINT ["/idly"]
