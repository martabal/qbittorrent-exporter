package qbit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"

	API "qbit-exp/api"
	"qbit-exp/app"
	"qbit-exp/deltasync"
	"qbit-exp/logger"
	prom "qbit-exp/prometheus"

	"github.com/prometheus/client_golang/prometheus"
)

// syncState holds the persistent state for delta sync.
// Initialized on first scrape, persists between scrapes.
var syncState *deltasync.State

// scrapeCount tracks number of scrapes for periodic full refresh.
var scrapeCount int64

// fullRefreshInterval forces a full sync every N scrapes to prevent state drift.
const fullRefreshInterval = 100

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

const baseAPIRUL string = "/api/v2/"

func newData(url string, handler func(body []byte, r *prometheus.Registry, webUIVersion *string) error) Data {
	return Data{
		Process:    url,
		URL:        baseAPIRUL + url,
		HTTPMethod: http.MethodGet,
		Handle:     handler,
	}
}

var firstAPIRequest = newData("app/webapiVersion", nil)

// staticAPIRequests are requests that don't benefit from delta sync.
// These are small responses that change rarely.
var staticAPIRequests = [...]Data{
	newData("app/version", func(body []byte, r *prometheus.Registry, _ *string) error {
		prom.Version(&body, r)

		return nil
	}),
	newData("app/preferences", func(body []byte, r *prometheus.Registry, _ *string) error {
		result := new(API.Preferences)
		err := json.Unmarshal(body, result)
		if err != nil {
			return err
		}
		prom.Preference(result, r)

		return nil
	}),
}

func createUrl(url string) string {
	return app.QBittorrent.BaseUrl + url
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

	err = data.Handle(body, r, webUIVersion)
	if err != nil {
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

	err = json.Unmarshal(body, &result)
	if err != nil {
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
	for i := range uniqueTrackers {
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

		go processData(&trackerInfo)
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
			logger.Error(err.Error())
		}
	}

	prom.Trackers(*responses, r)
}

func AllRequests(r *prometheus.Registry) error {
	var wg sync.WaitGroup

	firstRequestUrl := createUrl(firstAPIRequest.URL)

	webUIVersionBytes, retry, err := apiRequest(firstRequestUrl, firstAPIRequest.HTTPMethod, firstAPIRequest.QueryParams)
	if retry {
		logger.Debug("Retrying ...")

		webUIVersionBytes, _, err = apiRequest(firstRequestUrl, firstAPIRequest.HTTPMethod, firstAPIRequest.QueryParams)
	}

	webUIVersion := string(webUIVersionBytes)
	logger.Trace("WebUI API version: " + webUIVersion)

	if err != nil {
		return err
	}

	// Initialize sync state if needed
	if syncState == nil {
		syncState = deltasync.NewState()
	}

	// Periodic full refresh to prevent state drift
	scrapeCount++
	if scrapeCount%fullRefreshInterval == 0 {
		logger.Debug("Forcing full sync for state drift prevention")
		syncState.Reset()
	}

	// Fetch delta maindata (replaces both torrents/info and sync/maindata)
	deltaErr := fetchDeltaMainData()
	if deltaErr != nil {
		return deltaErr
	}

	// Get data from sync state for prometheus metrics
	torrents := syncState.GetTorrents()
	mainData := syncState.GetMainData()

	// Register torrent metrics
	prom.Torrent(&torrents, &webUIVersion, r)

	// Register maindata metrics (categories, tags, server state)
	prom.MainData(&mainData, r)

	// Fetch tracker info if enabled
	if app.Exporter.Features.EnableTracker {
		getTrackers(&torrents, r)
	}

	// Fetch static requests in parallel (app/version, app/preferences)
	c := make(chan func() (bool, error), len(staticAPIRequests))
	processData := func(data *Data) {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				logger.Error(fmt.Sprintf("Recovered panic: %s", r))
			}
		}()

		getData(r, data, &webUIVersion, c)
	}

	for _, request := range staticAPIRequests {
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

// fetchDeltaMainData fetches sync/maindata with rid parameter and applies to state.
func fetchDeltaMainData() error {
	rid := syncState.GetRID()
	url := createUrl(baseAPIRUL + "sync/maindata")

	queryParams := &[]QueryParams{
		{Key: "rid", Value: strconv.FormatInt(rid, 10)},
	}

	body, retry, err := apiRequest(url, http.MethodGet, queryParams)
	if retry {
		logger.Debug("Retrying delta maindata request...")

		body, _, err = apiRequest(url, http.MethodGet, queryParams)
	}

	if err != nil {
		return err
	}

	var delta API.DeltaMainData

	err = json.Unmarshal(body, &delta)
	if err != nil {
		errMsg := fmt.Errorf("%s %s: %w", unmarshError, url, err)
		errorHelper(&body, &errMsg, &url)

		return err
	}

	// Log sync mode for debugging
	if delta.FullUpdate || rid == 0 {
		logger.Debug(fmt.Sprintf("Full sync: %d torrents", len(delta.Torrents)))
	} else {
		logger.Trace(fmt.Sprintf("Delta sync: %d torrent updates, %d removed",
			len(delta.Torrents), len(delta.TorrentsRemoved)))
	}

	// Apply delta to state
	syncState.Apply(&delta)

	return nil
}

func errorHelper(body *[]byte, errMsg *error, url *string) {
	logger.Trace(fmt.Sprintf("body from %s: %s", *url, string(*body)))
	logger.Error(fmt.Sprintf("%s %s", unmarshError, *errMsg))
}

// returns:
// - body (content of the http response)
// - retry (if it should retry that query)
// - err (the error if there was one during the request).
func apiRequest(url string, method string, queryParams *[]QueryParams) ([]byte, bool, error) {
	if app.QBittorrent.Cookie == nil {
		logger.Debug("no cookie set")

		err := Auth()
		if err != nil {
			return nil, false, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), app.QBittorrent.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
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
	logger.Trace("New request to " + req.URL.String())
	resp, err := app.HttpClient.Do(req)

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		logger.Error(API.QbittorrentTimeOut)

		return nil, false, context.DeadlineExceeded
	}

	if err != nil {
		err := fmt.Errorf("%s: %w", API.ErrorConnect, err)
		logger.Error(err.Error())

		return nil, false, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Error(fmt.Sprintf("Error closing body %v", err))
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

		logger.Warn("Cookie changed, trying to reconnect ...")

		_ = Auth()

		return nil, true, err
	default:
		err := fmt.Errorf("%d", resp.StatusCode)
		logger.Error(fmt.Sprintf("Error code %d for: %s", resp.StatusCode, url))

		return nil, false, err
	}
}
