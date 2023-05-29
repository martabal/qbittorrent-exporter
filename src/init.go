package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"qbit-exp/src/models"
	qbit "qbit-exp/src/qbit"
	"strings"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

func main() {
	startup()
	log.Info("qbittorrent URL :", models.Getbaseurl())
	log.Info("username :", models.GetUsername())
	log.Info("password :", models.Getpasswordmasked())
	log.Info("Started")
	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(":8090", nil)
}

func metrics(w http.ResponseWriter, req *http.Request) {
	registry := prometheus.NewRegistry()
	err := qbit.Allrequests(registry)
	if err != nil {
		qbit.Allrequests(registry)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, req)
}

func startup() {
	log.SetLevel(log.TraceLevel)
	projectinfo()
	models.SetPromptError(false)
	loadenv()

	qbit.Auth()
}

func projectinfo() {
	fileContent, err := os.ReadFile("./package.json")
	if err != nil {
		log.Fatal(err)
		return
	}

	var res map[string]interface{}
	err = json.Unmarshal(fileContent, &res)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("Author:", res["author"])
	fmt.Println(res["name"], "version", res["version"])
}

func loadenv() {
	var envfile bool
	flag.BoolVar(&envfile, "e", false, "Use .env file")
	flag.Parse()
	if envfile {
		err := godotenv.Load(".env")
		if err != nil {
			log.Panic("Error loading .env file")
		}
		fmt.Println("Using .env file")
	}
	qbitUsername := getEnv("QBITTORRENT_USERNAME", "admin", true, "Qbittorrent username is not set. Using default username")
	qbitPassword := getEnv("QBITTORRENT_PASSWORD", "adminadmin", true, "Qbittorrent password is not set. Using default password")
	qbitURL := getEnv("QBITTORRENT_BASE_URL", "http://localhost:8080", true, "Qbittorrent base_url is not set. Using default base_url")

	setLogLevel(os.Getenv("LOG_LEVEL"))

	models.Init(qbitURL, qbitUsername, qbitPassword)
}

func setLogLevel(logLevel string) {
	logLevels := map[string]log.Level{
		"DEBUG": log.DebugLevel,
		"INFO":  log.InfoLevel,
		"WARN":  log.WarnLevel,
		"ERROR": log.ErrorLevel,
	}

	level, found := logLevels[strings.ToUpper(logLevel)]
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
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	if printLog {
		log.Warn(logPrinted)
	}

	return fallback
}
