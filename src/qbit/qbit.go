package qbit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"

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

const unmarshError = "Can not unmarshal JSON for"

const (
	RefQbitVersion = "qbitversion"
	RefPreference  = "preference"
	RefInfo        = "info"
	RefTransfer    = "transfer"
	RefMainData    = "maindata"
	RefTracker     = "tracker"
)

var info = []Data{
	{
		URL:         "/api/v2/app/version",
		HTTPMethod:  http.MethodGet,
		Ref:         RefQbitVersion,
		QueryParams: nil,
	},
	{
		URL:         "/api/v2/app/preferences",
		HTTPMethod:  http.MethodGet,
		Ref:         RefPreference,
		QueryParams: nil,
	},
	{
		URL:         "/api/v2/torrents/info",
		HTTPMethod:  http.MethodGet,
		Ref:         RefInfo,
		QueryParams: nil,
	},
	{
		URL:         "/api/v2/sync/maindata",
		HTTPMethod:  http.MethodGet,
		Ref:         RefMainData,
		QueryParams: nil,
	},
	{
		URL:         "/api/v2/transfer/info",
		HTTPMethod:  http.MethodGet,
		Ref:         RefTransfer,
		QueryParams: nil,
	},
}

var firstAPIRequest = info[0]
var otherAPIRequests = info[1:]

func createUrl(url string) string {
	return app.QBittorrent.BaseUrl + url
}

func getData(r *prometheus.Registry, data *Data, c chan func() (bool, error)) {
	url := createUrl(data.URL)
	body, retry, err := apiRequest(url, data.HTTPMethod, data.QueryParams)
	if retry {
		c <- (func() (bool, error) { return true, nil })
		return
	}
	if err != nil {
		c <- (func() (bool, error) { return false, err })
		return
	}

	unmarshErr := fmt.Errorf("%s %s", unmarshError, url)

	handleUnmarshal := func(target interface{}, body []byte) bool {
		if err := json.Unmarshal(body, target); err != nil {
			errorHelper(&body, &unmarshErr, &url)
			return false
		}
		return true
	}
	switch data.Ref {
	case RefInfo:
		result := new(API.Info)
		if handleUnmarshal(result, body) {
			prom.Torrent(result, r)
			if app.Exporter.Features.EnableTracker {
				getTrackers(result, r)
			}
		}
	case RefMainData:
		result := new(API.MainData)
		if handleUnmarshal(result, body) {
			prom.MainData(result, r)
		}
	case RefPreference:
		result := new(API.Preferences)
		if handleUnmarshal(result, body) {
			prom.Preference(result, r)
		}
	case RefQbitVersion:
		prom.Version(&body, r)
	case "transfer":
		result := new(API.Transfer)
		if handleUnmarshal(result, body) {
			prom.Transfer(result, r)
		}
	default:
		errormessage := fmt.Sprintf("Unknown reference: %s", data.Ref)
		panic(errormessage)
	}
	c <- (func() (bool, error) { return false, nil })
}

func getTrackersInfo(data *Data, c chan func() (*API.Trackers, error)) {
	url := createUrl(data.URL)
	body, _, err := apiRequest(url, data.HTTPMethod, data.QueryParams)

	if err != nil {
		c <- (func() (*API.Trackers, error) { return nil, err })
	}

	result := new(API.Trackers)
	if err := json.Unmarshal(body, &result); err != nil {
		errMsg := fmt.Errorf("%s %s", unmarshError, RefTracker)
		errorHelper(&body, &errMsg, &url)
	} else {
		c <- (func() (*API.Trackers, error) { return result, err })
	}

}

func getTrackers(torrentList *API.Info, r *prometheus.Registry) {
	var wg sync.WaitGroup
	uniqueValues := make(map[string]struct{})
	var uniqueTrackers []UniqueTracker
	for _, obj := range *torrentList {
		if _, exists := uniqueValues[obj.Tracker]; !exists {
			uniqueValues[obj.Tracker] = struct{}{}
			uniqueTrackers = append(uniqueTrackers, UniqueTracker{Tracker: obj.Tracker, Hash: obj.Hash})
		}
	}

	responses := new([]*API.Trackers)
	tracker := make(chan func() (*API.Trackers, error), len(uniqueTrackers))
	processData := func(trackerInfo *Data) {
		defer wg.Done()
		getTrackersInfo(trackerInfo, tracker)
	}
	for i := 0; i < len(uniqueTrackers); i++ {
		var trackerInfo = Data{
			URL:        "/api/v2/torrents/trackers",
			HTTPMethod: http.MethodGet,
			Ref:        RefTracker,
			QueryParams: &[]QueryParams{
				{
					Key:   "hash",
					Value: uniqueTrackers[i].Hash,
				},
			},
		}
		wg.Add(1)
		processData(&trackerInfo)
	}
	go func() {
		wg.Wait()
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
	var wg sync.WaitGroup
	c := make(chan func() (bool, error), 1)
	defer close(c)

	retry, err := func() (bool, error) {
		getData(r, &firstAPIRequest, c)
		return (<-c)()
	}()
	if retry {
		logger.Log.Debug("Retrying ...")
		_, err = func() (bool, error) {
			getData(r, &info[0], c)
			return (<-c)()
		}()
	}
	if err != nil {
		return err
	}
	newc := make(chan func() (bool, error), len(otherAPIRequests))
	processData := func(data *Data) {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Error(fmt.Sprintf("Recovered panic: %s", r))
			}
		}()
		getData(r, data, newc)
	}
	for _, request := range otherAPIRequests {
		wg.Add(1)
		go processData(&request)
	}
	go func() {
		wg.Wait()
		close(newc)
	}()

	for respFunc := range newc {
		_, err := respFunc()
		if err != nil {
			return err
		}
	}
	return nil
}

func errorHelper(body *[]byte, errMsg *error, url *string) {
	logger.Log.Trace(fmt.Sprintf("body from %s: %s", *url, string(*body)))
	logger.Log.Error(fmt.Sprintf("%s %s", unmarshError, *errMsg))
}

// returns:
// - body (content of the http response)
// - retry (if it should retry that query)
// - err (the error if there was one during the request)
func apiRequest(url string, method string, queryParams *[]QueryParams) ([]byte, bool, error) {
	if app.QBittorrent.Cookie == nil {
		logger.Log.Debug("no cookie set")
		err := Auth()
		if err != nil {
			return nil, false, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), app.QBittorrent.Timeout)
	defer cancel()

	req, err := http.NewRequest(method, url, nil)
	req = req.WithContext(ctx)
	if err != nil {
		panic(fmt.Sprintf("%s %s", API.ErrorWithUrl, err.Error()))
	}
	if queryParams != nil {
		q := req.URL.Query()
		for _, obj := range *queryParams {
			q.Add(obj.Key, obj.Value)
		}
		req.URL.RawQuery = q.Encode()
	}

	if app.QBittorrent.BasicAuth != nil {
		req.SetBasicAuth(app.QBittorrent.BasicAuth.Username, app.QBittorrent.BasicAuth.Password)
	}

	req.AddCookie(&http.Cookie{Name: "SID", Value: *app.QBittorrent.Cookie})
	logger.Log.Trace(fmt.Sprintf("New request to %s", req.URL.String()))
	resp, err := app.HttpClient.Do(req)
	if ctx.Err() == context.DeadlineExceeded {
		logger.Log.Error(API.QbittorrentTimeOut)
		return nil, false, context.DeadlineExceeded
	}

	if err != nil {
		err := fmt.Errorf("%s: %v", API.ErrorConnect, err)
		logger.Log.Error(err.Error())
		return nil, false, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, false, err
		}
		return body, false, nil
	case http.StatusForbidden:
		err := fmt.Errorf("%d", resp.StatusCode)
		logger.Log.Warn("Cookie changed, try to reconnect ...")
		_ = Auth()
		return nil, true, err
	default:
		err := fmt.Errorf("%d", resp.StatusCode)
		logger.Log.Error("Error code " + strconv.Itoa(resp.StatusCode))
		return nil, false, err
	}
}
