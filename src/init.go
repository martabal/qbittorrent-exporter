package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"qbit-exp/models"
	qbit "qbit-exp/qbit"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const DEFAULTPORT = 8090

var (
	Version     = "dev"
	Author      = "martabal"
	ProjectName = "qbittorrent-exporter"
)

var logLevels = map[string]log.Level{
	"TRACE": log.TraceLevel,
	"DEBUG": log.DebugLevel,
	"INFO":  log.InfoLevel,
	"WARN":  log.WarnLevel,
	"ERROR": log.ErrorLevel,
}

func main() {
	loadenv()
	fmt.Printf("%s (version %s)\n", ProjectName, Version)
	fmt.Println("Author: ", Author)
	fmt.Println("Using log level: ", log.GetLevel())

	qbit.Auth(true)

	log.Info("qbittorrent URL: ", models.Getbaseurl())
	log.Info("username: ", models.GetUsername())
	log.Info("password: ", models.Getpasswordmasked())
	log.Info("Started")
	http.HandleFunc("/metrics", metrics)
	addr := ":" + strconv.Itoa(models.GetPort())
	if models.GetPort() != DEFAULTPORT {
		log.Info("Listening on port", models.GetPort())
	}
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func metrics(w http.ResponseWriter, req *http.Request) {
	log.Trace("New request")
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
			log.Panic("Error loading .env file:", err)
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
		log.Panic("EXPORTER_PORT must be an integer")
	}
	if num < 0 || num > 65353 {
		log.Panic("EXPORTER_PORT must be > 0 and < 65353")
	}

	setLogLevel(getEnv("LOG_LEVEL", "INFO", false, ""))
	models.SetApp(num, false, strings.ToLower(disableTracker) == "true")
	models.SetQbit(qbitURL, qbitUsername, qbitPassword)
}

func setLogLevel(logLevel string) {
	upperLogLevel := strings.ToUpper(logLevel)
	level, found := logLevels[upperLogLevel]
	if !found {
		level = log.InfoLevel
	}

	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
}

func getEnv(key string, fallback string, printLog bool, logPrinted string) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		if printLog {
			log.Warn(logPrinted)
		}
		return fallback
	}
	return value
}
