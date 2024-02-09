package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"qbit-exp/models"
	qbit "qbit-exp/qbit"
	"strconv"
	"strings"

	logger "qbit-exp/logger"

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
	fmt.Println("Using log level: " + models.GetLogLevel())

	qbit.Auth(true)

	logger.Log.Info("qbittorrent URL: " + models.Getbaseurl())
	logger.Log.Info("username: " + models.GetUsername())
	logger.Log.Info("password: " + models.Getpasswordmasked())
	logger.Log.Info("Started")
	http.HandleFunc("/metrics", metrics)
	addr := ":" + strconv.Itoa(models.GetPort())
	if models.GetPort() != DEFAULTPORT {
		logger.Log.Info("Listening on port" + strconv.Itoa(models.GetPort()))
	}
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

func metrics(w http.ResponseWriter, req *http.Request) {
	logger.Log.Debug("New request")
	registry := prometheus.NewRegistry()
	qbit.Allrequests(registry)
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
		// fmt.Println("Using .env file")
	}

	qbitUsername := getEnv("QBITTORRENT_USERNAME", "admin", true, "Qbittorrent username is not set. Using default username")
	qbitPassword := getEnv("QBITTORRENT_PASSWORD", "adminadmin", true, "Qbittorrent password is not set. Using default password")
	qbitURL := strings.TrimSuffix(getEnv("QBITTORRENT_BASE_URL", "http://localhost:8080", true, "Qbittorrent base_url is not set. Using default base_url"), "/")
	exporterPort := getEnv("EXPORTER_PORT", strconv.Itoa(DEFAULTPORT), false, "")
	disableTracker := getEnv("DISABLE_TRACKER", "false", false, "")

	num, err := strconv.Atoi(exporterPort)

	if err != nil {
		panic("EXPORTER_PORT must be an integer")
	}
	if num < 0 || num > 65353 {
		panic("EXPORTER_PORT must be > 0 and < 65353")
	}

	loglevel := setLogLevel(getEnv("LOG_LEVEL", "INFO", false, ""))
	models.SetApp(num, false, strings.ToLower(disableTracker) == "true", loglevel)
	models.SetQbit(qbitURL, qbitUsername, qbitPassword)
}

func setLogLevel(logLevel string) string {
	upperLogLevel := strings.ToUpper(logLevel)
	level, found := logger.LogLevels[upperLogLevel]
	if !found {
		upperLogLevel = "INFO"
		level = logger.LogLevels[upperLogLevel]
	}

	opts := logger.PrettyHandlerOptions{
		SlogOpts: slog.HandlerOptions{
			Level: slog.Level(level),
		},
	}
	handler := logger.NewPrettyHandler(os.Stdout, opts)
	logger.Log = slog.New(handler)
	return upperLogLevel
}

func getEnv(key string, fallback string, printLog bool, logPrinted string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		if printLog {
			logger.Log.Warn(logPrinted)
		}
		return fallback
	}
	return value
}
