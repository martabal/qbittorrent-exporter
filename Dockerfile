FROM golang:1.22.2-alpine3.19 AS builder

ARG BUILD_VERSION

WORKDIR /app

COPY src .

RUN if [ -n "${BUILD_VERSION}" ]; then \
        go build -o /go/bin/qbittorrent-exporter -ldflags="-X 'main.Version=${BUILD_VERSION}'" . ; \
    else \
        go build -o /go/bin/qbittorrent-exporter . ; \
    fi

FROM alpine:3.19

WORKDIR /go/bin

COPY --from=builder /go/bin/qbittorrent-exporter .

CMD ["/go/bin/qbittorrent-exporter"]
