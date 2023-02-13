package main

import (
	"log"
	"net/http"
	"qbit-exp/src/models"
	qbit "qbit-exp/src/qbit"

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
