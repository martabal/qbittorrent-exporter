FROM golang:1.25.3-alpine3.22@sha256:a7a64bfb3b8a4724993e1c49d01ad331b5d8bc52461b3ecf89e9aedf3aadc635 AS builder

ARG BUILD_VERSION

WORKDIR /app

COPY src .

RUN if [ -n "${BUILD_VERSION}" ]; then \
        go build -o /go/bin/qbittorrent-exporter -ldflags="-X 'qbit-exp/app.version=${BUILD_VERSION}'" . ; \
    else \
        go build -o /go/bin/qbittorrent-exporter . ; \
    fi

FROM alpine:3.22@sha256:9eec16c5eada75150a82666ba0ad6df76b164a6f8582ba5cb964c0813fa56625

WORKDIR /go/bin

COPY --from=builder /go/bin/qbittorrent-exporter .

CMD ["/go/bin/qbittorrent-exporter"]
