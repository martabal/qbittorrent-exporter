package qbit

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"net/http"
	"qbit-exp/src/models"
	prom "qbit-exp/src/prometheus"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	wg sync.WaitGroup
)

type Data struct {
	URL        string
	HTTPMethod string
	Ref        string
}

var info = []Data{
	{
		URL:        "/api/v2/app/version",
		HTTPMethod: http.MethodGet,
		Ref:        "qbitversion",
	},
	{
		URL:        "/api/v2/app/preferences",
		HTTPMethod: http.MethodGet,
		Ref:        "preference",
	},
	{
		URL:        "/api/v2/torrents/info",
		HTTPMethod: http.MethodGet,
		Ref:        "info",
	},
	{
		URL:        "/api/v2/sync/maindata",
		HTTPMethod: http.MethodGet,
		Ref:        "maindata",
	},
}

func getData(r *prometheus.Registry, data Data, goroutine bool) bool {
	if goroutine {
		defer wg.Done()
	}
	resp, retry, err := Apirequest(data.URL, data.HTTPMethod)
	if retry {
		return retry
	}
	if err != nil {
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	switch data.Ref {
	case "preference":
		var result models.TypePreferences
		if err := json.Unmarshal(body, &result); err != nil {
			log.Error("Can not unmarshal JSON for preferences")
		} else {
			prom.Sendbackmessagepreference(&result, r)
		}
	case "info":
		var result models.TypeInfo
		if err := json.Unmarshal(body, &result); err != nil {
			log.Error("Can not unmarshal JSON for torrents info")
		} else {
			prom.Sendbackmessagetorrent(&result, r)
		}
	case "maindata":
		var result models.TypeMaindata
		if err := json.Unmarshal(body, &result); err != nil {
			log.Error("Can not unmarshal JSON for maindata")
		} else {
			prom.Sendbackmessagemaindata(&result, r)
		}
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
		log.Panicln("Unknown reference: ", data.Ref)
	}
	return false
}

func Allrequests(r *prometheus.Registry) {
	retry := getData(r, info[0], false)
	if retry {
		log.Debug("Retrying ...")
		getData(r, info[0], false)
	}
	wg.Add(len(info) - 1)
	for i := 1; i < len(info); i++ {
		go getData(r, info[i], true)
	}
	wg.Wait()
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
	}

	switch resp.StatusCode {
	case http.StatusOK:
		if models.GetPromptError() {
			models.SetPromptError(false)
		}
		return resp, false, nil
	case http.StatusForbidden:
		err := fmt.Errorf("%d", resp.StatusCode)
		if !models.GetPromptError() {
			models.SetPromptError(true)
			log.Warn("Cookie changed, try to reconnect ...")
		}
		Auth(false)
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
