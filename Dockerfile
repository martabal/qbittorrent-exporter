FROM golang:1.24.3-alpine3.21@sha256:ef18ee7117463ac1055f5a370ed18b8750f01589f13ea0b48642f5792b234044 AS builder

ARG BUILD_VERSION

WORKDIR /app

COPY src .

RUN if [ -n "${BUILD_VERSION}" ]; then \
        go build -o /go/bin/qbittorrent-exporter -ldflags="-X 'qbit-exp/app.version=${BUILD_VERSION}'" . ; \
    else \
        go build -o /go/bin/qbittorrent-exporter . ; \
    fi

FROM alpine:3.22@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715

WORKDIR /go/bin

COPY --from=builder /go/bin/qbittorrent-exporter .

CMD ["/go/bin/qbittorrent-exporter"]
