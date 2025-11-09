package main

import (
	"fmt"
	"net"
	"net/http"

	app "qbit-exp/app"
	logger "qbit-exp/logger"
	"qbit-exp/qbit"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	app.LoadEnv()

	_ = qbit.Auth()

	metrics := func(w http.ResponseWriter, req *http.Request) {
		metrics(w, req, qbit.AllRequests)
	}
	if app.Exporter.BasicAuth != nil {
		metrics = basicAuth(metrics)
	}
	http.HandleFunc(app.Exporter.Path, metrics)

	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, app.Exporter.Path, http.StatusFound)
	})
	addr := fmt.Sprintf(":%d", app.Exporter.Port)

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
			if username == app.Exporter.BasicAuth.Username && password == app.Exporter.BasicAuth.Password {
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
