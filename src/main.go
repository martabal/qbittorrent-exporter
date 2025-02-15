package main

import (
	"fmt"
	"net"
	"net/http"

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
	if app.Exporter.Features.EnableBasicAuthRequestHeader {
		logger.Log.Info("Enabling qBittorrent Basic Auth request header.")
	}
	logger.Log.Info(fmt.Sprintf("qBittorrent URL: %s", app.QBittorrent.BaseUrl))
	logger.Log.Info(fmt.Sprintf("username: %s", app.QBittorrent.Username))
	password := app.GetPasswordMasked()
	if app.Exporter.Features.ShowPassword {
		password = app.QBittorrent.Password
	}
	logger.Log.Info(fmt.Sprintf("password: %s", password))
	logger.Log.Info(fmt.Sprintf("Features enabled: %s", app.GetFeaturesEnabled()))
	logger.Log.Info("Started")

	_ = qbit.Auth()

	metrics := func(w http.ResponseWriter, req *http.Request) {
		metrics(w, req, qbit.AllRequests)
	}
	if app.Exporter.BasicAuth.Username != nil && app.Exporter.BasicAuth.Password != nil {
		metrics = basicAuth(metrics)
		logger.Log.Info("Using basic auth to protect the exporter instance")
	} else {
		logger.Log.Trace("Not using basic auth to protect the exporter instance")
	}
	http.HandleFunc(app.Exporter.Path, metrics)
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
	logMsg := "New request"
	if err == nil {
		logMsg = fmt.Sprintf("%s from %s", logMsg, ip)
	}
	logger.Log.Trace(logMsg)

	registry := prometheus.NewRegistry()
	err = allRequestsFunc(registry)
	if err != nil {
		http.Error(w, "", http.StatusServiceUnavailable)
	} else {
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, req)
	}

}

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			if username == *app.Exporter.BasicAuth.Username && password == *app.Exporter.BasicAuth.Password {
				h.ServeHTTP(w, r)
				return
			}
		}
		logErr := "Invalid auth"
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			logErr = fmt.Sprintf("%s from %s", logErr, ip)
		}
		logger.Log.Warn(logErr)

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
