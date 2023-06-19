# qbittorrent-exporter

[![Publish Release](https://github.com/martabal/qbittorrent-exporter/actions/workflows/docker.yml/badge.svg)](https://github.com/martabal/qbittorrent-exporter/actions/workflows/docker.yml)
[![Build](https://github.com/martabal/qbittorrent-exporter/actions/workflows/build.yml/badge.svg)](https://github.com/martabal/qbittorrent-exporter/actions/workflows/build.yml)
[![Test](https://github.com/martabal/qbittorrent-exporter/actions/workflows/test.yml/badge.svg)](https://github.com/martabal/qbittorrent-exporter/actions/workflows/test.yml)

<p align="center">
<img src="img/qbittorrent.png" width=100> <img src="img/prometheus.png" width=100><img src="img/golang.png" width=100>
</p>

This app is a Prometheus exporter for qBittorrent.  
You must have version 4.1.0 of qBittorrent or higher.  
This app is made to be integrated with the [qbittorrent-grafana-dashboard](https://github.com/caseyscarborough/qbittorrent-grafana-dashboard)  

## About this App

### Credits

I was using an excellent [exporter](https://github.com/caseyscarborough/qbittorrent-exporter) written in Java and I wanted to learn Go, that's how I got the idea to rewrite the exporter in Go.

However, this project adds support for tags and categories. It tracks the categories and tags of each torrent and the global categories and tags.

### Resources

This app uses ~20 times less RAM compared to the [original exporter](https://github.com/caseyscarborough/qbittorrent-exporter) for the same amount of torrents.
Docker compressed size is ~10 MB.

## Run it

### Docker-cli ([click here for more info](https://docs.docker.com/engine/reference/commandline/cli/))

```sh
docker run --name=qbit \
    -e QBITTORRENT_URL=http://192.168.1.10:8080 \
    -e QBITTORRENT_PASSWORD='<your_password>' \
    -e QBITTORRENT_USERNAME=admin \
    -p 8090:8090 \
    martabal/qbittorrent-exporter
```

### Docker-compose

```yaml
version: "2.1"
services:
  immich:
    image: martabal/qbittorrent-exporter:latest
    container_name: qbittorrent-exporter
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
git clone https://github.com/martabal/qbittorrent-exporter.git
cd qbittorent-exporter
go get -d -v
cd src
go build -o ./qbittorrent-exporter
./qbittorrent-exporter
```

or

```sh
git clone https://github.com/martabal/qbittorrent-exporter.git
cd qbittorent-exporter
go get -d -v
cd src
go run ./src
```

If you want to use an .env file, edit `.env.example` to match your setup, rename it `.env` then run it in the same directory. If you want to force to use the environment variables use `-e` argument like :

```sh
./qbittorrent-exporter -e
```

or

```sh
go run ./src -e
```

## Parameters

### Environment variables

| Parameters | Function | Default Value |
| :-----: | ----- | ----- |
| `-p 8090` | Webservice port |  |
| `-e QBITTORRENT_USERNAME` | qBittorrent username | `admin` |
| `-e QBITTORRENT_PASSWORD` | qBittorrent password | `adminadmin` |
| `-e QBITTORRENT_BASE_URL` | qBittorrent base URL | `http://localhost:8090` |
| `-e LOG_LEVEL` | App log level (`DEBUG`, `INFO`, `WARN` and `ERROR`) | `INFO` |

### Arguments

| Arguments | Function |
| :-----: | ----- |
| -e | If qbittorrent-exporter detects a .env file in the same directory, the values in the .env will be used, `-e` forces the usage of environment variables |
