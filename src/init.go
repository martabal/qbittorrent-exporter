package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"qbit-exp/src/models"
	qbit "qbit-exp/src/qbit"

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
	var envfile bool
	models.SetPromptError(false)
	flag.BoolVar(&envfile, "e", false, "Use .env file")
	flag.Parse()
	if envfile {
		useenvfile()
	} else {
		initenv()
	}

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

func useenvfile() {
	myEnv, err := godotenv.Read()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	username := myEnv["QBITTORRENT_USERNAME"]
	password := myEnv["QBITTORRENT_PASSWORD"]
	qbitURL := myEnv["QBITTORRENT_BASE_URL"]
	logLevel := myEnv["LOG_LEVEL"]

	if username == "" {
		username = "admin"
		log.Warn("Qbittorrent username is not set. Using default username")
	}
	if password == "" {
		password = "adminadmin"
		log.Warn("Qbittorrent password is not set. Using default password")
	}
	if qbitURL == "" {
		qbitURL = "http://localhost:8090"
		log.Warn("Qbittorrent base_url is not set. Using default base_url")
	}

	setLogLevel(logLevel)
	models.Init(qbitURL, username, password)

	log.Debug("Using .env file")
}

func initenv() {
	qbitUsername := os.Getenv("QBITTORRENT_USERNAME")
	qbitPassword := os.Getenv("QBITTORRENT_PASSWORD")
	qbitURL := os.Getenv("QBITTORRENT_BASE_URL")

	if qbitUsername == "" {
		qbitUsername = "admin"
		log.Warn("Qbittorrent username is not set. Using default username")
	}
	if qbitPassword == "" {
		qbitPassword = "adminadmin"
		log.Warn("Qbittorrent password is not set. Using default password")
	}
	if qbitURL == "" {
		qbitURL = "http://localhost:8080"
		log.Warn("Qbittorrent base_url is not set. Using default base_url")
	}

	setLogLevel(os.Getenv("LOG_LEVEL"))

	models.Init(qbitURL, qbitUsername, qbitPassword)
}

func setLogLevel(log_level string) {
	if log_level == "DEBUG" {
		log.SetLevel(log.DebugLevel)
	} else if log_level == "INFO" {
		log.SetLevel(log.InfoLevel)
	} else if log_level == "WARN" {
		log.SetLevel(log.WarnLevel)
	} else if log_level == "ERROR" {
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
}
