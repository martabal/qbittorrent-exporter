FROM golang:1.20-alpine3.18 AS builder

WORKDIR /app

COPY . .

RUN go get -d -v ./src/ && \
    go build -o /go/bin/qbittorrent-exporter ./src 

FROM alpine:3.18

COPY --from=builder /go/bin/qbittorrent-exporter /go/bin/qbittorrent-exporter
COPY package.json /go/bin/

WORKDIR /go/bin
CMD ["/go/bin/qbittorrent-exporter"]
