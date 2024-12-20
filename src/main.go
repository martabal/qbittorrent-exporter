package main

import (
	"fmt"
	"net"
	"net/http"
	"runtime"

	"qbit-exp/qbit"

	app "qbit-exp/app"
	logger "qbit-exp/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const devVersion = "dev"

var (
	Version     = devVersion
	Author      = "martabal"
	ProjectName = "qbittorrent-exporter"
)

func main() {
	app.LoadEnv()
	fmt.Printf("%s (version %s)\n", ProjectName, Version)
	fmt.Printf("Author: %s\n", Author)
	fmt.Printf("Using log level: %s%s%s\n", logger.ColorLogLevel[logger.LogLevels[app.Exporter.LogLevel]], app.Exporter.LogLevel, logger.Reset)

	envFileMessage := "Using environment variables"
	if app.UsingEnvFile {
		envFileMessage = "Using .env"
	}
	logger.Log.Debug(envFileMessage)
	logger.Log.Info(fmt.Sprintf("qbittorrent URL: %s", app.QBittorrent.BaseUrl))
	logger.Log.Info(fmt.Sprintf("username: %s", app.QBittorrent.Username))
	logger.Log.Info(fmt.Sprintf("password: %s", app.GetPasswordMasked()))
	logger.Log.Info(fmt.Sprintf("Features enabled: %s", app.GetFeaturesEnabled()))
	logger.Log.Info("Started")

	_ = qbit.Auth()

	http.HandleFunc(app.Exporter.Path, func(w http.ResponseWriter, req *http.Request) {
		metrics(w, req, qbit.AllRequests)
	})
	addr := fmt.Sprintf(":%d", app.Exporter.Port)
	if app.Exporter.Port != app.DefaultExporterPort {
		logger.Log.Info(fmt.Sprintf("Listening on port %d", app.Exporter.Port))
	}

	if Version == devVersion {
		app.Exporter.URL = fmt.Sprintf("http://localhost:%d%s", app.Exporter.Port, app.Exporter.Path)
	}

	if app.Exporter.URL != "" {
		logger.Log.Info(fmt.Sprintf("qbittorrent-exporter URL: %s", app.Exporter.URL))
	}

	logger.Log.Info("Starting the exporter")
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func metrics(w http.ResponseWriter, req *http.Request, allRequestsFunc func(*prometheus.Registry) error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		logger.Log.Trace(fmt.Sprintf("New request from %s", ip))
	} else {
		logger.Log.Trace("New request")
	}

	registry := prometheus.NewRegistry()
	err = allRequestsFunc(registry)
	if err != nil {
		http.Error(w, "", http.StatusServiceUnavailable)
		runtime.GC()
	} else {
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, req)
	}

}
