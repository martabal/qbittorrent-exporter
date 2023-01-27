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
	flag.BoolVar(&envfile, "e", false, "Use .env file")
	flag.Parse()

	if envfile {
		useenvfile()
	} else {
		initenv()
	}
	log.Println("Loaded all env... Starting")
	qbit.Auth()
}
func useenvfile() {

	myEnv, err := godotenv.Read()
	username := myEnv["QBITTORENT_USERNAME"]
	password := myEnv["QBITTORENT_PASSWORD"]
	qbit_url := myEnv["QBITTORENT_BASE_URL"]
	if myEnv["QBITTORENT_USERNAME"] == "" {
		log.Println("Qbittorrent username is not set. Using default username")
		username = "admin"
	}
	if myEnv["QBITTORENT_PASSWORD"] == "" {
		log.Println("Qbittorrent password is not set. Using default password")
		password = "adminadmin"
	}
	if myEnv["QBITTORENT_BASE_URL"] == "" {
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
	username := os.Getenv("QBITTORENT_USERNAME")
	password := os.Getenv("QBITTORENT_PASSWORD")
	qbit_url := os.Getenv("QBITTORENT_BASE_URL")
	if os.Getenv("QBITTORENT_USERNAME") == "" {
		log.Println("Qbittorrent username is not set. Using default username")
		username = "admin"
	}
	if os.Getenv("QBITTORENT_PASSWORD") == "" {
		log.Println("Qbittorrent password is not set. Using default password")
		password = "adminadmin"
	}
	if os.Getenv("QBITTORENT_BASE_URL") == "" {
		log.Println("Qbittorrent base_url is not set. Using default base_url")
		qbit_url = "http://localhost:8090"
	}

	models.Setuser(username, password)
	models.Setbaseurl(qbit_url)

}
