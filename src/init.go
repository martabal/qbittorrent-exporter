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
	username := myEnv["QBITTORRENT_USERNAME"]
	password := myEnv["QBITTORRENT_PASSWORD"]
	qbit_url := myEnv["QBITTORRENT_BASE_URL"]
	if myEnv["QBITTORRENT_USERNAME"] == "" {
		log.Warn("Qbittorrent username is not set. Using default username")
		username = "admin"
	}
	if myEnv["QBITTORRENT_PASSWORD"] == "" {
		log.Warn("Qbittorrent password is not set. Using default password")
		password = "adminadmin"
	}
	if myEnv["QBITTORRENT_BASE_URL"] == "" {
		log.Warn("Qbittorrent base_url is not set. Using default base_url")
		qbit_url = "http://localhost:8090"
	}

	if myEnv["LOG_LEVEL"] == "DEBUG" {
		log.SetLevel(log.DebugLevel)
	} else if myEnv["LOG_LEVEL"] == "INFO" {
		log.SetLevel(log.InfoLevel)
	} else if myEnv["LOG_LEVEL"] == "WARN" {
		log.SetLevel(log.WarnLevel)
	} else if myEnv["LOG_LEVEL"] == "ERROR" {
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	models.Setuser(username, password)
	models.Setbaseurl(qbit_url)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Info("Using .env file")
}

func initenv() {
	username := os.Getenv("QBITTORRENT_USERNAME")
	password := os.Getenv("QBITTORRENT_PASSWORD")
	qbit_url := os.Getenv("QBITTORRENT_BASE_URL")
	if os.Getenv("QBITTORRENT_USERNAME") == "" {
		log.Warn("Qbittorrent username is not set. Using default username")
		username = "admin"
	}
	if os.Getenv("QBITTORRENT_PASSWORD") == "" {
		log.Warn("Qbittorrent password is not set. Using default password")
		password = "adminadmin"
	}
	if os.Getenv("QBITTORRENT_BASE_URL") == "" {
		log.Warn("Qbittorrent base_url is not set. Using default base_url")
		qbit_url = "http://localhost:8080"
	}

	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		log.SetLevel(log.DebugLevel)
	} else if os.Getenv("LOG_LEVEL") == "INFO" {
		log.SetLevel(log.InfoLevel)
	} else if os.Getenv("LOG_LEVEL") == "WARN" {
		log.SetLevel(log.WarnLevel)
	} else if os.Getenv("LOG_LEVEL") == "ERROR" {
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	models.Setuser(username, password)
	models.Setbaseurl(qbit_url)
}
