package main

import (
	"fmt"
	"net"
	"net/http"
	"runtime"

	"qbit-exp/qbit"

	app "qbit-exp/app"
	logger "qbit-exp/logger"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Version     = "dev"
	Author      = "martabal"
	ProjectName = "qbittorrent-exporter"
)

func main() {
	app.LoadEnv()
	fmt.Printf("%s (version %s)\n", ProjectName, Version)
	fmt.Println("Author:", Author)
	fmt.Println("Using log level: " + fmt.Sprintf("%s%s%s", logger.ColorLogLevel[logger.LogLevels[app.Exporter.LogLevel]], app.Exporter.LogLevel, logger.Reset))

	envFileMessage := "Using environment variables"
	if app.UsingEnvFile {
		envFileMessage = "Using .env"
	}
	logger.Log.Debug(envFileMessage)
	logger.Log.Info("qbittorrent URL: " + app.QBittorrent.BaseUrl)
	logger.Log.Info("username: " + app.QBittorrent.Username)
	logger.Log.Info("password: " + app.GetPasswordMasked())
	logger.Log.Info("Features enabled: " + app.GetFeaturesEnabled())
	logger.Log.Info("Started")

	err := qbit.Auth()
	if err != nil && app.ShouldShowError {
		logger.Log.Error(err.Error())
		app.ShouldShowError = false
	}

	http.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		metrics(w, req, qbit.AllRequests)
	})
	addr := ":" + strconv.Itoa(app.Exporter.Port)
	if app.Exporter.Port != app.DEFAULT_PORT {
		logger.Log.Info("Listening on port " + strconv.Itoa(app.Exporter.Port))
	}
	logger.Log.Info("Starting the exporter")
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func metrics(w http.ResponseWriter, req *http.Request, allRequestsFunc func(*prometheus.Registry) error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		logger.Log.Trace("New request from " + ip)
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
