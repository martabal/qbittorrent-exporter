package qbit

import (
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"qbit-exp/src/models"
	prom "qbit-exp/src/prometheus"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
)

type Dict map[string]interface{}

func getData(r *prometheus.Registry, url string, httpmethod string, data string) bool {

	resp, err := Apirequest(url, httpmethod)
	if err == true {
		return err
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
			case "qbitversion":
				qbittorrent_app_version := prometheus.NewGauge(prometheus.GaugeOpts{
					Name: "qbittorrent_app_version",
					Help: "The current qBittorrent version",
					ConstLabels: map[string]string{
						"version": string(body),
					},
				})
				r.MustRegister(qbittorrent_app_version)
				qbittorrent_app_version.Set(1)
			default:
				log.Panicln("Unknown type: ", data)
			}
		}
	}
	return false
}

func Allrequests(r *prometheus.Registry) {
	array := []Dict{
		{"url": "/api/v2/app/preferences", "httpmethod": "GET", "structuretype": "preference"},
		{"url": "/api/v2/torrents/info", "httpmethod": "GET", "structuretype": "response"},
		{"url": "/api/v2/sync/maindata", "httpmethod": "GET", "structuretype": "maindata"},
		{"url": "/api/v2/app/version", "httpmethod": "GET", "structuretype": "qbitversion"},
	}

	for i := 0; i < len(array); i++ {
		url := array[i]["url"].(string)
		httpmethod := array[i]["httpmethod"].(string)
		structuretype := array[i]["structuretype"].(string)
		err := getData(r, url, httpmethod, structuretype)
		if err == true {
			getData(r, url, httpmethod, structuretype)
		}
	}
}

func Apirequest(uri string, method string) (*http.Response, bool) {

	req, err := http.NewRequest(method, models.Getbaseurl()+uri, nil)
	if err != nil {
		log.Fatalln("Error with url")
	}

	req.AddCookie(&http.Cookie{Name: "SID", Value: models.Getcookie()})
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err := fmt.Errorf("Can't connect to server")
		if !models.GetPromptError() {
			log.Debug(err.Error())
			models.SetPromptError(true)
		}
		return resp, false
	} else {
		models.SetPromptError(false)
		if resp.StatusCode == 200 {
			return resp, false
		} else if resp.StatusCode == 403 {

			if !models.GetPromptError() {
				models.SetPromptError(true)
				log.Warn("Cookie changed, try to reconnect ...")
			}
			cookie, newerr := Auth(false)
			if newerr == nil {
				models.Setcookie(cookie)
			}
			return resp, true
		} else {
			if !models.GetPromptError() {
				models.SetPromptError(true)

				log.Debug("Error code ", err.Error())
			}
			return resp, false
		}
	}
}
