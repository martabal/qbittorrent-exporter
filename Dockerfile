FROM golang:1.23.4-alpine3.20@sha256:9a31ef0803e6afdf564edc8ba4b4e17caed22a0b1ecd2c55e3c8fdd8d8f68f98 AS builder

ARG BUILD_VERSION

WORKDIR /app

COPY src .

RUN if [ -n "${BUILD_VERSION}" ]; then \
        go build -o /go/bin/qbittorrent-exporter -ldflags="-X 'main.Version=${BUILD_VERSION}'" . ; \
    else \
        go build -o /go/bin/qbittorrent-exporter . ; \
    fi

FROM alpine:3.21@sha256:21dc6063fd678b478f57c0e13f47560d0ea4eeba26dfc947b2a4f81f686b9f45

WORKDIR /go/bin

COPY --from=builder /go/bin/qbittorrent-exporter .

CMD ["/go/bin/qbittorrent-exporter"]
