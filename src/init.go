package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"qbit-exp/src/models"
	qbit "qbit-exp/src/qbit"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	startup()
	log.Println("qbittorrent URL :", models.Getbaseurl())
	log.Println("username :", models.GetUsername())
	log.Println("password :", models.Getpasswordmasked())
	log.Println("Started")
	http.HandleFunc("/metrics", metrics)
	http.ListenAndServe(":8090", nil)
}

func metrics(w http.ResponseWriter, req *http.Request) {
	registry := prometheus.NewRegistry()
	err := qbit.Allrequests(registry)
	if err != nil {
		err = qbit.Allrequests(registry)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, req)
}

func startup() {
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
	fileContent, err := os.Open("./package.json")

	if err != nil {
		log.Fatal(err)
		return
	}

	defer fileContent.Close()

	byteResult, _ := ioutil.ReadAll(fileContent)

	var res map[string]interface{}
	json.Unmarshal([]byte(byteResult), &res)
	log.Println("Author :", res["author"])
	log.Println(res["name"], "version", res["version"])
}

func useenvfile() {
	myEnv, err := godotenv.Read()
	username := myEnv["QBITTORRENT_USERNAME"]
	password := myEnv["QBITTORRENT_PASSWORD"]
	qbit_url := myEnv["QBITTORRENT_BASE_URL"]
	if myEnv["QBITTORRENT_USERNAME"] == "" {
		log.Println("Qbittorrent username is not set. Using default username")
		username = "admin"
	}
	if myEnv["QBITTORRENT_PASSWORD"] == "" {
		log.Println("Qbittorrent password is not set. Using default password")
		password = "adminadmin"
	}
	if myEnv["QBITTORRENT_BASE_URL"] == "" {
		log.Println("Qbittorrent base_url is not set. Using default base_url")
		qbit_url = "http://localhost:8090"
	}
	models.Setuser(username, password)
	models.Setbaseurl(qbit_url)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println("Using .env file")
}

func initenv() {
	username := os.Getenv("QBITTORRENT_USERNAME")
	password := os.Getenv("QBITTORRENT_PASSWORD")
	qbit_url := os.Getenv("QBITTORRENT_BASE_URL")
	if os.Getenv("QBITTORRENT_USERNAME") == "" {
		log.Println("Qbittorrent username is not set. Using default username")
		username = "admin"
	}
	if os.Getenv("QBITTORRENT_PASSWORD") == "" {
		log.Println("Qbittorrent password is not set. Using default password")
		password = "adminadmin"
	}
	if os.Getenv("QBITTORRENT_BASE_URL") == "" {
		log.Println("Qbittorrent base_url is not set. Using default base_url")
		qbit_url = "http://localhost:8080"
	}
	models.Setuser(username, password)
	models.Setbaseurl(qbit_url)
}
