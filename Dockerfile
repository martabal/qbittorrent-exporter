FROM golang:1.24.4-alpine3.21@sha256:56a23791af0f77c87b049230ead03bd8c3ad41683415ea4595e84ce7eada121a AS builder

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
