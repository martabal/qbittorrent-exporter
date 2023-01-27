FROM golang:alpine AS builder

WORKDIR $GOPATH/src/mypackage/myapp/

COPY . .

RUN go get -d -v && \
    go build -o /go/bin/qbittorrent-promtheus

FROM alpine

COPY --from=builder /go/bin/qbittorrent-promtheus /go/bin/qbittorrent-promtheus

CMD ["/go/bin/qbittorrent-promtheus"]
