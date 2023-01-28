package main

import (
	"flag"
	"log"
	"os"
	"qbit-prom/src/models"
	qbit "qbit-prom/src/qbit"

	"github.com/joho/godotenv"
)

func startup() {
	var envfile bool
	models.SetPromptError(false)
	flag.BoolVar(&envfile, "e", false, "Use .env file")
	flag.Parse()
	log.Println("Loading all parameters")
	if envfile {
		useenvfile()
	} else {
		initenv()
	}

	qbit.Auth()
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
