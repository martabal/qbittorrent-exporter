package main

import (
	"fmt"
	"log"
	"net/http"
	"qbit-prom/src/models"
	qbit "qbit-prom/src/qbit"
)

func main() {
	startup()
	http.HandleFunc("/metrics", metrics)
	log.Println("qbittorrent URL :", models.Getbaseurl())
	log.Println("username :", models.GetUsername())
	log.Println("password :", models.Getpasswordmasked())
	log.Println("Started")
	http.ListenAndServe(":8090", nil)
}

func metrics(w http.ResponseWriter, req *http.Request) {
	value := qbit.Allrequests()
	if value == "" {
		value = qbit.Allrequests()
	}
	fmt.Fprintf(w, value)
}
