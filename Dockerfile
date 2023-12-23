FROM golang:1.21-alpine3.18 AS builder

WORKDIR /app

COPY go.* .
COPY src src

RUN go build -o /go/bin/qbittorrent-exporter ./src

FROM alpine:3.18

COPY --from=builder /go/bin/qbittorrent-exporter /go/bin/qbittorrent-exporter

WORKDIR /go/bin

CMD ["/go/bin/qbittorrent-exporter"]
