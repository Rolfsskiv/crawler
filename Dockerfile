FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o benchmark -ldflags="-w -s" .

FROM alpine:latest

WORKDIR $GOPATH/src/github.com/rolf/concurrency-crawler-benchmark
RUN apk add --update ca-certificates && \
        rm -rf /var/cache/apk/* /tmp/*
COPY --from=builder /app/benchmark /usr/bin/

ENTRYPOINT ["benchmark"]
