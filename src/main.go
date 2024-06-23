package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"runtime"

	"qbit-exp/qbit"

	app "qbit-exp/app"
	logger "qbit-exp/logger"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const DEFAULT_PORT = 8090
const DEFAULT_TIMEOUT = 30

var (
	Version     = "dev"
	Author      = "martabal"
	ProjectName = "qbittorrent-exporter"
)

func main() {
	loadenv()
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
	if app.Port != DEFAULT_PORT {
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

func loadenv() {
	var envfile bool
	flag.BoolVar(&envfile, "e", false, "Use .env file")
	flag.Parse()
	_, err := os.Stat(".env")
	if !os.IsNotExist(err) && !envfile {
		err := godotenv.Load(".env")
		if err != nil {
			errormessage := "Error loading .env file:" + err.Error()
			panic(errormessage)
		}
	}

	loglevel := setLogLevel(getEnv("LOG_LEVEL", "INFO", ""))
	qbitUsername := getEnv("QBITTORRENT_USERNAME", "admin", "Qbittorrent username is not set. Using default username")
	qbitPassword := getEnv("QBITTORRENT_PASSWORD", "adminadmin", "Qbittorrent password is not set. Using default password")
	qbitURL := strings.TrimSuffix(getEnv("QBITTORRENT_BASE_URL", "http://localhost:8080", "Qbittorrent base_url is not set. Using default base_url"), "/")
	exporterPortEnv := getEnv("EXPORTER_PORT", strconv.Itoa(DEFAULT_PORT), "")
	timeoutDurationEnv := getEnv("QBITTORRENT_TIMEOUT", strconv.Itoa(DEFAULT_TIMEOUT), "")
	disableTracker := getEnv("DISABLE_TRACKER", "false", "")

	exporterPort, errExporterPort := strconv.Atoi(exporterPortEnv)

	if errExporterPort != nil {
		panic("EXPORTER_PORT must be an integer")
	}
	if exporterPort < 0 || exporterPort > 65353 {
		panic("EXPORTER_PORT must be > 0 and < 65353")
	}

	timeoutDuration, errTimeoutDuration := strconv.Atoi(timeoutDurationEnv)
	if errTimeoutDuration != nil {
		panic("QBITTORRENT_TIMEOUT must be an integer")
	}

	app.SetVar(exporterPort, strings.ToLower(disableTracker) == "true", loglevel, qbitURL, qbitUsername, qbitPassword, timeoutDuration)
}

func setLogLevel(logLevel string) string {
	upperLogLevel := strings.ToUpper(logLevel)
	level, found := logger.LogLevels[upperLogLevel]
	if !found {
		upperLogLevel = "INFO"
		level = logger.LogLevels[upperLogLevel]
	}

	opts := slog.HandlerOptions{
		Level: slog.Level(level),
	}

	handler := logger.NewPrettyHandler(os.Stdout, opts)
	logger.Log = slog.New(handler)
	return upperLogLevel
}

func getEnv(key string, fallback string, logMessage string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		if logMessage != "" {
			logger.Log.Warn(logMessage)
		}
		return fallback
	}
	return value
}
