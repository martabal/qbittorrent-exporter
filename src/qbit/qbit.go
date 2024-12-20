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

func getData(r *prometheus.Registry, data *Data, wg *sync.WaitGroup, c chan func() (bool, error)) {
	if wg != nil {
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

	unmarshErr := fmt.Sprintf("%s %s", unmarshError, data.Ref)

	handleUnmarshal := func(target interface{}, body []byte) bool {
		if err := json.Unmarshal(body, target); err != nil {
			errorHelper(body, unmarshErr)
			return false
		}
		return true
	}
	switch data.Ref {
	case RefInfo:
		result := new(API.Info)
		if handleUnmarshal(result, body) {
			prom.Torrent(result, r)
			if app.Exporter.Feature.EnableTracker {
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
		errormessage := "Unknown reference: " + data.Ref
		panic(errormessage)
	}
	c <- (func() (bool, error) { return false, nil })
}

func getTrackersInfo(data *Data, wg *sync.WaitGroup, c chan func() (*API.Trackers, error)) {
	defer wg.Done()
	body, _, err := apiRequest(data.URL, data.HTTPMethod, data.QueryParams)

	if err != nil {
		c <- (func() (*API.Trackers, error) { return nil, err })
	}

	result := new(API.Trackers)
	if err := json.Unmarshal(body, &result); err != nil {
		errorHelper(body, fmt.Sprintf("%s %s", unmarshError, RefTracker))
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
	tracker := make(chan func() (*API.Trackers, error))
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
		go getTrackersInfo(&trackerInfo, &wg, tracker)
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
	c := make(chan func() (bool, error))

	go getData(r, &info[0], nil, c)
	retry, err := (<-c)()
	if retry {
		logger.Log.Debug("Retrying ...")
		go getData(r, &info[0], nil, c)
		_, err = (<-c)()
	}
	if err != nil {
		return err
	}
	for i := 1; i < len(info); i++ {
		wg.Add(1)
		go getData(r, &info[i], &wg, c)
	}
	go func() {
		wg.Wait()
		close(c)
	}()

	for respFunc := range c {
		_, err := respFunc()
		if err != nil {
			return err
		}
	}
	return nil
}

func errorHelper(body []byte, errMsg string) {
	logger.Log.Debug(string(body))
	logger.Log.Error(fmt.Sprintf("%s %s", unmarshError, errMsg))
}

// returns:
// - body (content of the http response)
// - retry (if it should retry that query)
// - err (the error if there was one during the request)
func apiRequest(uri string, method string, queryParams *[]QueryParams) ([]byte, bool, error) {
	if app.QBittorrent.Cookie == nil {
		logger.Log.Debug("no cookie set")
		err := Auth()
		if err != nil {
			return nil, false, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), app.QBittorrent.Timeout)
	defer cancel()

	req, err := http.NewRequest(method, app.QBittorrent.BaseUrl+uri, nil)
	req = req.WithContext(ctx)
	if err != nil {
		panic(API.ErrorWithUrl + err.Error())
	}
	if queryParams != nil {
		q := req.URL.Query()
		for _, obj := range *queryParams {
			q.Add(obj.Key, obj.Value)
		}
		req.URL.RawQuery = q.Encode()
	}

	req.AddCookie(&http.Cookie{Name: "SID", Value: *app.QBittorrent.Cookie})
	client := &http.Client{}
	logger.Log.Trace("New request to " + req.URL.String())
	resp, err := client.Do(req)
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
