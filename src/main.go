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
	fmt.Println("Using log level: " + app.LogLevel)

	qbit.Auth()

	logger.Log.Info("qbittorrent URL: " + app.BaseUrl)
	logger.Log.Info("username: " + app.Username)
	logger.Log.Info("password: " + app.GetPasswordMasked())
	logger.Log.Info("Started")
	http.HandleFunc("/metrics", metrics)
	addr := ":" + strconv.Itoa(app.Port)
	if app.Port != app.DEFAULT_PORT {
		logger.Log.Info("Listening on port " + strconv.Itoa(app.Port))
	}
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func metrics(w http.ResponseWriter, req *http.Request) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		logger.Log.Debug("New request from " + ip)
	} else {
		logger.Log.Debug("New request")
	}

	registry := prometheus.NewRegistry()
	err = qbit.AllRequests(registry)
	if err != nil {
		http.Error(w, "", http.StatusServiceUnavailable)
		runtime.GC()
	} else {
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, req)
	}

}
