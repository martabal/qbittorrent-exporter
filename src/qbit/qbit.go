package qbit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

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
	Process     string
	URL         string
	HTTPMethod  string
	QueryParams *[]QueryParams
	Handle      func(body []byte, r *prometheus.Registry, webUIVersion *string) error
}

type UniqueTracker struct {
	Tracker string
	Hash    string
}

const unmarshError string = "Can not unmarshal JSON for"

const baseAPIRUL = "/api/v2/"

func newData(url string, queryParams *[]QueryParams, handler func(body []byte, r *prometheus.Registry, webUIVersion *string) error) Data {
	return Data{
		Process:     url,
		URL:         baseAPIRUL + url,
		QueryParams: queryParams,
		HTTPMethod:  http.MethodGet,
		Handle:      handler,
	}
}

var firstAPIRequest = newData("app/webapiVersion", nil, nil)

var otherAPIRequests = [...]Data{
	newData("app/version", nil, func(body []byte, r *prometheus.Registry, _ *string) error {
		prom.Version(&body, r)
		return nil
	}),
	newData("app/preferences", nil, func(body []byte, r *prometheus.Registry, _ *string) error {
		result := new(API.Preferences)
		if err := json.Unmarshal(body, result); err != nil {
			return err
		}
		prom.Preference(result, r)
		return nil
	}),
	newData("torrents/info", nil, func(body []byte, r *prometheus.Registry, webUIVersion *string) error {
		result := new(API.SliceInfo)
		if err := json.Unmarshal(body, result); err != nil {
			return err
		}
		prom.Torrent(result, webUIVersion, r)
		if app.Exporter.Features.EnableTracker {
			getTrackers(result, r)
		}
		return nil
	}),
	newData("sync/maindata", nil, func(body []byte, r *prometheus.Registry, _ *string) error {
		result := new(API.MainData)
		if err := json.Unmarshal(body, result); err != nil {
			return err
		}
		prom.MainData(result, r)
		return nil
	}),
	newData("transfer/info", nil, func(body []byte, r *prometheus.Registry, _ *string) error {
		result := new(API.Transfer)
		if err := json.Unmarshal(body, result); err != nil {
			return err
		}
		prom.Transfer(result, r)
		return nil
	}),
}

func createUrl(url string) string {
	return fmt.Sprintf("%s%s", app.QBittorrent.BaseUrl, url)
}

func getData(r *prometheus.Registry, data *Data, webUIVersion *string, c chan func() (bool, error)) {
	url := createUrl(data.URL)
	body, retry, err := apiRequest(url, data.HTTPMethod, data.QueryParams)
	if retry {
		c <- func() (bool, error) { return true, nil }
		return
	}
	if err != nil {
		c <- func() (bool, error) { return false, err }
		return
	}

	if err := data.Handle(body, r, webUIVersion); err != nil {
		errormessage := fmt.Errorf("%s %s: %w", unmarshError, url, err)
		errorHelper(&body, &errormessage, &url)
		c <- func() (bool, error) { return false, err }
		return
	}

	c <- func() (bool, error) { return false, nil }
}

func getTrackersInfo(data *Data, c chan func() (*API.Trackers, error)) {
	url := createUrl(data.URL)
	body, _, err := apiRequest(url, data.HTTPMethod, data.QueryParams)

	if err != nil {
		c <- (func() (*API.Trackers, error) { return nil, err })
	}

	result := new(API.Trackers)
	if err := json.Unmarshal(body, &result); err != nil {
		errMsg := fmt.Errorf("%s %s", unmarshError, data.Process)
		errorHelper(&body, &errMsg, &url)
	} else {
		c <- (func() (*API.Trackers, error) { return result, err })
	}

}

func getTrackers(torrentList *API.SliceInfo, r *prometheus.Registry) {
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
	firstRequestUrl := createUrl(firstAPIRequest.URL)
	webUIVersionBytes, retry, err := apiRequest(firstRequestUrl, firstAPIRequest.HTTPMethod, firstAPIRequest.QueryParams)
	if retry {
		logger.Log.Debug("Retrying ...")
		webUIVersionBytes, _, err = apiRequest(firstRequestUrl, firstAPIRequest.HTTPMethod, firstAPIRequest.QueryParams)
	}
	webUIVersion := string(webUIVersionBytes)
	logger.Log.Trace(fmt.Sprintf("WebUI API version: %s", webUIVersion))
	if err != nil {
		return err
	}
	c := make(chan func() (bool, error), len(otherAPIRequests))
	processData := func(data *Data) {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Error(fmt.Sprintf("Recovered panic: %s", r))
			}
		}()
		getData(r, data, &webUIVersion, c)
	}
	for _, request := range otherAPIRequests {
		wg.Add(1)
		go processData(&request)
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Log.Error(fmt.Sprintf("Error closing body %v", err))
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, false, err
		}
		return body, false, nil
	case http.StatusForbidden:
		err := fmt.Errorf("%d", resp.StatusCode)
		logger.Log.Warn("Cookie changed, trying to reconnect ...")
		_ = Auth()
		return nil, true, err
	default:
		err := fmt.Errorf("%d", resp.StatusCode)
		logger.Log.Error(fmt.Sprintf("Error code %d for: %s", resp.StatusCode, url))
		return nil, false, err
	}
}
