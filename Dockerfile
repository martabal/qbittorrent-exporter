FROM golang:1.19-alpine3.17 AS builder

WORKDIR $GOPATH/src/mypackage/myapp/

COPY . .

RUN go get -d -v && \
    go build -o /go/bin/qbittorrent-exporter

FROM alpine:3.17

COPY --from=builder /go/bin/qbittorrent-exporter /go/bin/qbittorrent-exporter

CMD ["/go/bin/qbittorrent-exporter"]
