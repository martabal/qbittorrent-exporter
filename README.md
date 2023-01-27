# qbittorrent-prometheus

[![manual push](https://github.com/martabal/qbittorrent-prometheus/actions/workflows/push_docker.yml/badge.svg)](https://github.com/martabal/qbittorrent-prometheus/actions/workflows/push_docker.yml)

<p align="center">
<img src="img/qbittorrent.png" width=100> <img src="img/prometheus.png" width=100><img src="img/golang.png" width=100>
</p>

This app is a Prometheus exporter for qBittorrent.  
You must have version 4.1.0 of qBittorrent or higher.  
This app is made for to be integrated with the [qbittorrent-grafana-dashboard](https://github.com/caseyscarborough/qbittorrent-grafana-dashboard)  

## About this App

### Ressources

This app running uses between 5-12Mo of RAM and uses bearly no CPU power.

## Run it

## docker-cli ([click here for more info](https://docs.docker.com/engine/reference/commandline/cli/))

```sh
docker run --name=qbit \
    -e QBITTORENT_URL=http://192.168.1.10:8080 \
    -e QBITTORENT_PASSWORD='<your_password>' \
    -e QBITTORENT_USERNAME=admin \
    -p 8090:8090 \
    martabal/qbittorrent-prometheus
```

## docker-compose

```yaml
version: "2.1"
services:
  immich:
    image: martabal/qbittorrent-prometheus:latest
    container_name: qbittorrent-prometheus
    environment:
      - QBITTORENT_URL=http://192.168.1.10:8080
      - QBITTORENT_PASSWORD='<your_password>'
      - QBITTORENT_USERNAME=admin
    ports:
      - 8090:8090
    restart: unless-stopped
```

## Parameters

| Parameters | Function | Default Value |
| :-----: | ----- | ----- |
| `-p 8090` | Webservice port |  |
| `-e QBITTORENT_USERNAME` | qBittorrent username | `admin` |
| `-e QBITTORENT_PASSWORD` | qBittorrent password | `adminadmin` |
| `-e QBITTORENT_BASE_URL` | qBittorrent base URL | `http://localhost:8090` |
