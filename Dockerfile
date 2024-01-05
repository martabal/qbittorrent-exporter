FROM golang:1.21-alpine3.19 AS builder

ARG PROJECT_VERSION

WORKDIR /app

COPY go.* .
COPY src src

RUN if [ -n "${PROJECT_VERSION}" ]; then \
        go build -o /go/bin/qbittorrent-exporter -ldflags="-X 'main.Version=${PROJECT_VERSION}'" ./src; \
    else \
        go build -o /go/bin/qbittorrent-exporter ./src; \
    fi

FROM alpine:3.19

COPY --from=builder /go/bin/qbittorrent-exporter /go/bin/qbittorrent-exporter

WORKDIR /go/bin

CMD ["/go/bin/qbittorrent-exporter"]
