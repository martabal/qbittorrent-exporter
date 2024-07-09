package qbit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"net/http"
	API "qbit-exp/api"
	"qbit-exp/app"
	"qbit-exp/logger"

	prom "qbit-exp/prometheus"

	"github.com/prometheus/client_golang/prometheus"
)

type QueryParams struct {
	Key   string
	Value string
}

type Data struct {
	URL         string
	HTTPMethod  string
	Ref         string
	QueryParams *[]QueryParams
}

type UniqueTracker struct {
	Tracker string
	Hash    string
}

const UnmarshError = "Can not unmarshal JSON for "

var (
	wg        sync.WaitGroup
	wgTracker sync.WaitGroup
)

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
	{
		URL:         "/api/v2/transfer/info",
		HTTPMethod:  http.MethodGet,
		Ref:         "transfer",
		QueryParams: nil,
	},
}

func getData(r *prometheus.Registry, data Data, goroutine bool, c chan func() (bool, error)) {
	if goroutine {
		defer wg.Done()
	}
	body, retry, err := apiRequest(data.URL, data.HTTPMethod, data.QueryParams)
	if retry {
		c <- (func() (bool, error) { return true, nil })
		return
	}
	if err != nil {
		c <- (func() (bool, error) { return false, err })
		return
	}

	unmarshErr := UnmarshError + data.Ref

	handleUnmarshal := func(target interface{}, body []byte) bool {
		if err := json.Unmarshal(body, target); err != nil {
			errorHelper(body, err, unmarshErr)
			return false
		}
		return true
	}
	switch data.Ref {
	case "info":
		result := new(API.Info)
		if handleUnmarshal(result, body) {
			prom.Torrent(result, r)
			if !app.DisableTracker {
				getTrackers(result, r)
			}
		}
	case "maindata":
		result := new(API.MainData)
		if handleUnmarshal(result, body) {
			prom.MainData(result, r)
		}
	case "preference":
		result := new(API.Preferences)
		if handleUnmarshal(result, body) {
			prom.Preference(result, r)
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
	case "transfer":
		result := new(API.Transfer)
		if handleUnmarshal(result, body) {
			prom.Transfer(result, r)
		}
	default:
		errormessage := "Unknown reference: " + data.Ref
		panic(errormessage)
	}
	c <- (func() (bool, error) { return false, nil })
}

func getTrackersInfo(data Data, c chan func() (*API.Trackers, error)) {
	defer wgTracker.Done()
	body, _, err := apiRequest(data.URL, data.HTTPMethod, data.QueryParams)

	if err != nil {
		c <- (func() (*API.Trackers, error) { return nil, err })
	}

	if err != nil {
		c <- (func() (*API.Trackers, error) { return nil, err })
	}
	result := new(API.Trackers)
	unmarshErr := UnmarshError + "tracker"
	if err := json.Unmarshal(body, &result); err != nil {
		errorHelper(body, err, unmarshErr)
	} else {
		c <- (func() (*API.Trackers, error) { return result, err })
	}

}

func getTrackers(torrentList *API.Info, r *prometheus.Registry) {
	uniqueValues := make(map[string]struct{})
	var uniqueTrackers []UniqueTracker
	for _, obj := range *torrentList {
		if _, exists := uniqueValues[obj.Tracker]; !exists {
			uniqueValues[obj.Tracker] = struct{}{}
			uniqueTrackers = append(uniqueTrackers, UniqueTracker{Tracker: obj.Tracker, Hash: obj.Hash})
		}
	}

	responses := new([]*API.Trackers)
	tracker := make(chan func() (*API.Trackers, error))
	for i := 0; i < len(uniqueTrackers); i++ {
		var trackerInfo = Data{
			URL:        "/api/v2/torrents/trackers",
			HTTPMethod: http.MethodGet,
			Ref:        "tracker",
			QueryParams: &[]QueryParams{
				{
					Key:   "hash",
					Value: uniqueTrackers[i].Hash,
				},
			},
		}
		wgTracker.Add(1)
		go getTrackersInfo(trackerInfo, tracker)
	}
	go func() {
		wgTracker.Wait()
		close(tracker)
	}()
	for respFunc := range tracker {
		res, err := respFunc()
		if err == nil {
			*responses = append(*responses, res)
		} else {
			logger.Log.Error(err.Error())
		}

	}

	prom.Trackers(*responses, r)
}

func AllRequests(r *prometheus.Registry) error {
	c := make(chan func() (bool, error))

	go getData(r, info[0], false, c)
	retry, err := (<-c)()
	if retry {
		logger.Log.Debug("Retrying ...")
		go getData(r, info[0], false, c)
		_, err = (<-c)()
	}
	if err != nil {
		logger.Log.Debug(err.Error())
		return err
	}
	for i := 1; i < len(info); i++ {
		wg.Add(1)
		go getData(r, info[i], true, c)
	}
	go func() {
		wg.Wait()
		close(c)
	}()

	for respFunc := range c {
		_, err := respFunc()
		if err != nil {
			logger.Log.Debug(err.Error())
			return err
		}
	}
	return nil
}

func errorHelper(body []byte, err error, unmarshErr string) {
	logger.Log.Debug(string(body))
	logger.Log.Debug(err.Error())
	logger.Log.Error(unmarshErr)
}

func apiRequest(uri string, method string, queryParams *[]QueryParams) ([]byte, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*app.QbittorrentTimeout))
	defer cancel()

	req, err := http.NewRequest(method, app.BaseUrl+uri, nil)
	req = req.WithContext(ctx)
	if err != nil {
		panic("Error with url " + err.Error())
	}
	if queryParams != nil {
		q := req.URL.Query()
		for _, obj := range *queryParams {
			q.Add(obj.Key, obj.Value)
		}
		req.URL.RawQuery = q.Encode()
	}

	req.AddCookie(&http.Cookie{Name: "SID", Value: app.Cookie})
	client := &http.Client{}
	logger.Log.Debug("New request to " + req.URL.String())
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.Debug(err.Error())
		err := fmt.Errorf("can't connect to server")
		if app.ShouldShowError {
			logger.Log.Error(err.Error())
			app.ShouldShowError = false
		}
		return nil, false, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if !app.ShouldShowError {
			app.ShouldShowError = true
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, false, err
		}
		return body, false, nil
	case http.StatusForbidden:
		err := fmt.Errorf("%d", resp.StatusCode)
		if app.ShouldShowError {
			app.ShouldShowError = false
			logger.Log.Warn("Cookie changed, try to reconnect ...")
		}
		Auth()
		return nil, true, err
	default:
		err := fmt.Errorf("%d", resp.StatusCode)
		if app.ShouldShowError {
			app.ShouldShowError = false
			logger.Log.Error("Error code " + strconv.Itoa(resp.StatusCode))
		}
		return nil, false, err
	}
}
