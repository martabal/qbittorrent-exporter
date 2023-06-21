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

func getData(r *prometheus.Registry, url string, httpmethod string, ref string) bool {

	resp, retry, err := Apirequest(url, httpmethod)
	if retry == true {
		return retry
	} else if err == nil {
		if models.GetPromptError() {
			models.SetPromptError(false)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		} else {
			switch ref {
			case "preference":
				var result models.TypePreferences
				if err := json.Unmarshal(body, &result); err != nil {
					log.Debug("Can not unmarshal JSON for preferences")
				}
				prom.Sendbackmessagepreference(&result, r)
			case "torrents":
				var result models.TypeTorrents
				if err := json.Unmarshal(body, &result); err != nil {
					log.Debug("Can not unmarshal JSON for torrents info")
				}
				prom.Sendbackmessagetorrent(&result, r)
			case "maindata":
				var result models.TypeMaindata
				if err := json.Unmarshal(body, &result); err != nil {
					log.Debug("Can not unmarshal JSON for maindata")
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
				log.Panicln("Unknown type: ", ref)
			}
		}
	}
	return false
}

func Allrequests(r *prometheus.Registry) {
	array := []Dict{
		{"url": "/api/v2/app/preferences", "httpmethod": "GET", "ref": "preference"},
		{"url": "/api/v2/torrents/info", "httpmethod": "GET", "ref": "torrents"},
		{"url": "/api/v2/sync/maindata", "httpmethod": "GET", "ref": "maindata"},
		{"url": "/api/v2/app/version", "httpmethod": "GET", "ref": "qbitversion"},
	}

	for i := 0; i < len(array); i++ {
		url := array[i]["url"].(string)
		httpmethod := array[i]["httpmethod"].(string)
		structuretype := array[i]["ref"].(string)
		err := getData(r, url, httpmethod, structuretype)
		if err == true {
			getData(r, url, httpmethod, structuretype)
		}
	}
}

func Apirequest(uri string, method string) (*http.Response, bool, error) {

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
		return resp, false, err
	} else {
		models.SetPromptError(false)
		switch resp.StatusCode {
		case 200:
			return resp, false, nil
		case 403:
			err := fmt.Errorf("%d", resp.StatusCode)
			if !models.GetPromptError() {
				models.SetPromptError(true)
				log.Warn("Cookie changed, try to reconnect ...")
			}
			cookie, newerr := Auth(false)
			if newerr == nil {
				models.Setcookie(cookie)
			}
			return resp, true, err
		default:
			err := fmt.Errorf("%d", resp.StatusCode)
			if !models.GetPromptError() {
				models.SetPromptError(true)
				log.Debug("Error code ", resp.StatusCode)
			}
			return resp, false, err
		}
	}
}
