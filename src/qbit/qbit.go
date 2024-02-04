package qbit

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"net/http"
	"qbit-exp/models"
	prom "qbit-exp/prometheus"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	wg        sync.WaitGroup
	wgTracker sync.WaitGroup
)

type QueryParams struct {
	Key   string
	Value *string
}

type Data struct {
	URL         string
	HTTPMethod  string
	Ref         string
	QueryParams *[]QueryParams
}

var info = []Data{
	{
		URL:         "/api/v2/app/version",
		HTTPMethod:  http.MethodGet,
		Ref:         "qbitversion",
		QueryParams: nil,
	},
	{
		URL:         "/api/v2/app/preferences",
		HTTPMethod:  http.MethodGet,
		Ref:         "preference",
		QueryParams: nil,
	},
	{
		URL:         "/api/v2/torrents/info",
		HTTPMethod:  http.MethodGet,
		Ref:         "info",
		QueryParams: nil,
	},
	{
		URL:         "/api/v2/sync/maindata",
		HTTPMethod:  http.MethodGet,
		Ref:         "maindata",
		QueryParams: nil,
	},
}

func getData(r *prometheus.Registry, data Data, goroutine bool) bool {
	if goroutine {
		defer wg.Done()
	}
	resp, retry, err := Apirequest(data.URL, data.HTTPMethod, data.QueryParams)
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
	case "info":
		var result models.TypeInfo
		if err := json.Unmarshal(body, &result); err != nil {
			log.Error("Can not unmarshal JSON for torrents info")
		} else {
			prom.Sendbackmessagetorrent(&result, r)
			if !models.GetFeatureFlag() {
				getTrackers(&result, r)
			}

		}
	case "maindata":
		var result models.TypeMaindata
		if err := json.Unmarshal(body, &result); err != nil {
			log.Error("Can not unmarshal JSON for maindata")
		} else {
			prom.Sendbackmessagemaindata(&result, r)
		}
	case "preference":
		var result models.TypePreferences
		if err := json.Unmarshal(body, &result); err != nil {
			log.Error("Can not unmarshal JSON for preferences")
		} else {
			prom.Sendbackmessagepreference(&result, r)
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

func getTrackersInfo(data Data, c chan func() (*models.TypeTrackers, error)) {
	defer wgTracker.Done()
	resp, _, err := Apirequest(data.URL, data.HTTPMethod, data.QueryParams)

	if err != nil {
		c <- (func() (*models.TypeTrackers, error) { return nil, err })
	}
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		c <- (func() (*models.TypeTrackers, error) { return nil, err })
	}
	result := new(models.TypeTrackers)
	if err := json.Unmarshal(body, &result); err != nil {
		log.Error("Can not unmarshal JSON for preferences")
	} else {
		c <- (func() (*models.TypeTrackers, error) { return result, err })
	}

}

type UniqueObject struct {
	Tracker string
	Hash    string
}

func getTrackers(torrentList *models.TypeInfo, r *prometheus.Registry) {
	fmt.Println("test")
	uniqueValues := make(map[string]struct{})
	var uniqueObjects []UniqueObject
	for _, obj := range *torrentList {

		if _, exists := uniqueValues[obj.Tracker]; !exists {

			uniqueValues[obj.Tracker] = struct{}{}
			uniqueObjects = append(uniqueObjects, UniqueObject{Tracker: obj.Tracker, Hash: obj.Hash})
		}
	}

	responses := new([]*models.TypeTrackers)
	for i := 1; i < len(uniqueObjects); i++ {
		tracker := make(chan func() (*models.TypeTrackers, error))
		var trackerInfo = Data{
			URL:        "/api/v2/torrents/trackers",
			HTTPMethod: http.MethodGet,
			Ref:        "tracker",
			QueryParams: &[]QueryParams{
				{
					Key:   "hash",
					Value: &uniqueObjects[i].Hash,
				},
			},
		}
		wgTracker.Add(1)
		go getTrackersInfo(trackerInfo, tracker)
		res, err := (<-tracker)()
		if err == nil {
			*responses = append(*responses, res)
		}

	}
	wgTracker.Wait()

	prom.Sendbackmessagetrackers(*responses, r)

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

func Apirequest(uri string, method string, queryParams *[]QueryParams) (*http.Response, bool, error) {

	req, err := http.NewRequest(method, models.Getbaseurl()+uri, nil)
	if err != nil {
		log.Fatalln("Error with url")
	}
	if queryParams != nil {
		q := req.URL.Query()
		for _, obj := range *queryParams {
			if obj.Value != nil {
				q.Add(obj.Key, *obj.Value)
			}

		}
		req.URL.RawQuery = q.Encode()
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
