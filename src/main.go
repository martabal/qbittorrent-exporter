package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"qbit-exp/qbit"

	app "qbit-exp/app"
	logger "qbit-exp/logger"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const DEFAULTPORT = 8090

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
	if app.Port != DEFAULTPORT {
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
	qbit.AllRequests(registry)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, req)
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
	exporterPort := getEnv("EXPORTER_PORT", strconv.Itoa(DEFAULTPORT), "")
	disableTracker := getEnv("DISABLE_TRACKER", "false", "")

	num, err := strconv.Atoi(exporterPort)

	if err != nil {
		panic("EXPORTER_PORT must be an integer")
	}
	if num < 0 || num > 65353 {
		panic("EXPORTER_PORT must be > 0 and < 65353")
	}

	app.SetVar(num, strings.ToLower(disableTracker) == "true", loglevel, qbitURL, qbitUsername, qbitPassword)
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
