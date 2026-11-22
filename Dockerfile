FROM golang:1.27.4-alpine3.24@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS builder

ARG BUILD_VERSION

WORKDIR /app

COPY . .

RUN if [ -n "${BUILD_VERSION}" ]; then \
        go build -o /go/bin/qbittorrent-exporter -ldflags="-X 'qbit-exp/app.version=${BUILD_VERSION}'" . ; \
    else \
        go build -o /go/bin/qbittorrent-exporter . ; \
    fi

FROM alpine:3.24@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

WORKDIR /go/bin

COPY --from=builder /go/bin/qbittorrent-exporter .

CMD ["/go/bin/qbittorrent-exporter"]
