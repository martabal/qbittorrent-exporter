FROM golang:1.25.0-alpine3.22@sha256:f18a072054848d87a8077455f0ac8a25886f2397f88bfdd222d6fafbb5bba440 AS builder

ARG BUILD_VERSION

WORKDIR /app

COPY src .

RUN if [ -n "${BUILD_VERSION}" ]; then \
        go build -o /go/bin/qbittorrent-exporter -ldflags="-X 'qbit-exp/app.version=${BUILD_VERSION}'" . ; \
    else \
        go build -o /go/bin/qbittorrent-exporter . ; \
    fi

FROM alpine:3.22@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1

WORKDIR /go/bin

COPY --from=builder /go/bin/qbittorrent-exporter .

CMD ["/go/bin/qbittorrent-exporter"]
