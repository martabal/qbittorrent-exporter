# qbittorrent-prometheus

[![manual push](https://github.com/martabal/qbittorrent-prometheus/actions/workflows/push_docker.yml/badge.svg)](https://github.com/martabal/qbittorrent-prometheus/actions/workflows/push_docker.yml)

<p align="center">
<img src="img/qbittorrent.png" width=100> <img src="img/prometheus.png" width=100><img src="img/golang.png" width=100>
</p>

This app is a Prometheus exporter for qBittorrent.  
You must have version 4.1.0 of qBittorrent or higher.  
This app is made for to be integrated with the [qbittorrent-grafana-dashboard](https://github.com/caseyscarborough/qbittorrent-grafana-dashboard)  

## About this App

### Credits

I was using an excellent [exporter](https://github.com/caseyscarborough/qbittorrent-exporter) written in Java and I wanted to learn Go, that's how I got the idea to rewrite the exporter in Go.

### Resources

This app uses between 5-12Mo of RAM and uses bearly no CPU power.  
Docker compressed size is ~8 MB.

## Run it

### Docker-cli ([click here for more info](https://docs.docker.com/engine/reference/commandline/cli/))

```sh
docker run --name=qbit \
    -e QBITTORRENT_URL=http://192.168.1.10:8080 \
    -e QBITTORRENT_PASSWORD='<your_password>' \
    -e QBITTORRENT_USERNAME=admin \
    -p 8090:8090 \
    martabal/qbittorrent-prometheus
```

### Docker-compose

```yaml
version: "2.1"
services:
  immich:
    image: martabal/qbittorrent-prometheus:latest
    container_name: qbittorrent-prometheus
    environment:
      - QBITTORRENT_URL=http://192.168.1.10:8080
      - QBITTORRENT_PASSWORD='<your_password>'
      - QBITTORRENT_USERNAME=admin
    ports:
      - 8090:8090
    restart: unless-stopped
```

### Without docker

```sh
go get -d -v
go build -o ./qbittorrent-promtheus
./qbittorrent-prometheus
```

If you want to use an .env file, edit `.env.example` to match your setup then run it with :

```sh
./qbittorrent-prometheus -e
```

## Parameters

### Environment variables

| Parameters | Function | Default Value |
| :-----: | ----- | ----- |
| `-p 8090` | Webservice port |  |
| `-e QBITTORRENT_USERNAME` | qBittorrent username | `admin` |
| `-e QBITTORRENT_PASSWORD` | qBittorrent password | `adminadmin` |
| `-e QBITTORRENT_BASE_URL` | qBittorrent base URL | `http://localhost:8090` |

### Arguments

| Arguments | Function |
| :-----: | ----- |
| -e | Use the .env file (must be placed in the same directory) |
