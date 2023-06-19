package qbit

import (
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"qbit-exp/src/models"
	prom "qbit-exp/src/prometheus"

	log "github.com/sirupsen/logrus"

	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Dict map[string]interface{}

var wg sync.WaitGroup

func getData(r *prometheus.Registry, url string, httpmethod string, data string) {
	defer wg.Done()
	resp, err := Apirequest(url, httpmethod)
	if err != nil {
		if err.Error() == "403" {
			log.Debug("Cookie changed, try to reconnect ...")

		} else {
			if !models.GetPromptError() {
				log.Debug("Error : ", err)
			}
		}
	} else {
		if models.GetPromptError() {
			models.SetPromptError(false)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		} else {
			switch data {
			case "preference":
				var result models.TypePreferences
				if err := json.Unmarshal(body, &result); err != nil {
					log.Debug("Can not unmarshal JSON")
				}
				prom.Sendbackmessagepreference(&result, r)
			case "response":
				var result models.TypeResponse
				if err := json.Unmarshal(body, &result); err != nil {
					log.Debug("Can not unmarshal JSON")
				}
				prom.Sendbackmessagetorrent(&result, r)
			case "maindata":
				var result models.TypeMaindata
				if err := json.Unmarshal(body, &result); err != nil {
					log.Debug("Can not unmarshal JSON")
				}
				prom.Sendbackmessagemaindata(&result, r)
			default:
				log.Panicln("Unknown type: ", data)
			}

		}
	}

}

func Handlerequest(uri string, method string) (string, error) {

	resp, err := Apirequest(uri, method)
	if err != nil {

		if err.Error() == "403" {
			log.Debug("Cookie changed, try to reconnect ...")
			Auth()
		} else {
			if !models.GetPromptError() {
				log.Debug("Error : ", err)
			}
		}
		return "", err
	} else {
		if models.GetPromptError() {
			models.SetPromptError(false)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
			return "", err
		} else {
			sb := string(body)

			return sb, nil
		}
	}
}

func qbitversion(r *prometheus.Registry) {
	defer wg.Done()
	version, err := Handlerequest("/api/v2/app/version", "GET")
	if err != nil {
		if err.Error() == "403" {
			log.Debug("Cookie changed, try to reconnect ...")

		} else {
			if !models.GetPromptError() {
				log.Debug("Error : ", err)
			}
		}
	} else {
		qbittorrent_app_version := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "qbittorrent_app_version",
			Help: "The current qBittorrent version",
			ConstLabels: map[string]string{
				"version": version,
			},
		})
		r.MustRegister(qbittorrent_app_version)
		qbittorrent_app_version.Set(1)
	}
}

func Allrequests(r *prometheus.Registry) error {

	wg.Add(1)
	go qbitversion(r)

	array := []Dict{
		{"url": "/api/v2/app/preferences", "httpmethod": "GET", "structuretype": "preference"},
		{"url": "/api/v2/torrents/info", "httpmethod": "GET", "structuretype": "response"},
		{"url": "/api/v2/sync/maindata", "httpmethod": "GET", "structuretype": "maindata"},
	}

	for i := 0; i < len(array); i++ {
		url := array[i]["url"].(string)
		httpmethod := array[i]["httpmethod"].(string)
		structuretype := array[i]["structuretype"].(string)
		wg.Add(1)
		go getData(r, url, httpmethod, structuretype)
	}
	wg.Wait()
	return nil
}

func Apirequest(uri string, method string) (*http.Response, error) {

	req, err := http.NewRequest(method, models.Getbaseurl()+uri, nil)
	if err != nil {
		log.Fatalln("Error with url")
	}

	req.AddCookie(&http.Cookie{Name: "SID", Value: models.Getcookie()})
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err := fmt.Errorf("can't connect to server")
		if !models.GetPromptError() {
			log.Debug(err.Error())
			models.SetPromptError(true)
		}
		return resp, err
	} else {
		models.SetPromptError(false)
		if resp.StatusCode == 200 {
			return resp, nil
		} else {
			err := fmt.Errorf("%d", resp.StatusCode)
			if !models.GetPromptError() {
				models.SetPromptError(true)

				log.Debug("Error code", err.Error())

			}
			return resp, err
		}
	}
}
