package qbit

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	API "qbit-exp/api"
	"qbit-exp/app"
	"qbit-exp/deltasync"

	"github.com/prometheus/client_golang/prometheus"
)

func TestGetData_Success(t *testing.T) {
	setupMockApp()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key": "value"}`))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	handlerCalled := false
	data := &Data{
		Process:    "test",
		URL:        "/test",
		HTTPMethod: http.MethodGet,
		Handle: func(_ []byte, _ *prometheus.Registry, _ *string) error {
			handlerCalled = true
			return nil
		},
	}

	c := make(chan func() (bool, error), 1)
	version := "2.11.2"
	registry := prometheus.NewRegistry()

	getData(registry, data, &version, c)

	result := <-c
	retry, err := result()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if retry {
		t.Error("expected retry=false")
	}

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestGetData_Retry(t *testing.T) {
	setupMockApp()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	handlerCalled := false
	data := &Data{
		Process:    "test",
		URL:        "/test",
		HTTPMethod: http.MethodGet,
		Handle: func(_ []byte, _ *prometheus.Registry, _ *string) error {
			handlerCalled = true
			return nil
		},
	}

	c := make(chan func() (bool, error), 1)
	version := "2.11.2"
	registry := prometheus.NewRegistry()

	getData(registry, data, &version, c)

	result := <-c
	retry, _ := result()

	if !retry {
		t.Error("expected retry=true for 403 response")
	}

	if handlerCalled {
		t.Error("expected handler not to be called on retry")
	}
}

func TestGetData_HandleError(t *testing.T) {
	setupMockApp()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	expectedErr := errors.New("handler error")
	data := &Data{
		Process:    "test",
		URL:        "/test",
		HTTPMethod: http.MethodGet,
		Handle: func(_ []byte, _ *prometheus.Registry, _ *string) error {
			return expectedErr
		},
	}

	c := make(chan func() (bool, error), 1)
	version := "2.11.2"
	registry := prometheus.NewRegistry()

	getData(registry, data, &version, c)

	result := <-c
	_, err := result()

	if err == nil {
		t.Fatal("expected error from Handle, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped handler error, got %v", err)
	}
}

func TestFetchDeltaMainData_Success(t *testing.T) {
	setupMockApp()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"rid": 42,
			"full_update": true,
			"torrents": {
				"abc123": {"name": "Test Torrent", "state": "downloading", "progress": 0.5}
			},
			"server_state": {"dht_nodes": 100}
		}`))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL
	syncState = deltasync.NewState()

	err := fetchDeltaMainData()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if count := syncState.TorrentCount(); count != 1 {
		t.Errorf("expected 1 torrent in state, got %d", count)
	}

	if rid := syncState.GetRID(); rid != 42 {
		t.Errorf("expected RID=42, got %d", rid)
	}
}

func TestFetchDeltaMainData_InvalidJSON(t *testing.T) {
	setupMockApp()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL
	syncState = deltasync.NewState()

	err := fetchDeltaMainData()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestGetTrackers(t *testing.T) {
	setupMockApp()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"url": "http://tracker.example.com",
				"num_seeds": 5,
				"num_leeches": 2,
				"num_peers": 7,
				"num_downloaded": 10,
				"status": 2,
				"tier": 0,
				"msg": ""
			}
		]`))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	torrentList := &API.SliceInfo{
		{Name: "Torrent1", Hash: "hash1", Tracker: "http://tracker.example.com"},
	}

	registry := prometheus.NewRegistry()
	getTrackers(torrentList, registry)

	if !hasTrackerMetric(t, registry, "qbittorrent_tracker_seeders") {
		t.Error("expected qbittorrent_tracker_seeders after getTrackers")
	}
}

func TestGetTrackers_Deduplication(t *testing.T) {
	setupMockApp()

	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{
				"url": "http://tracker.example.com",
				"num_seeds": 5,
				"num_leeches": 0,
				"num_peers": 5,
				"num_downloaded": 0,
				"status": 2,
				"tier": 0,
				"msg": ""
			}
		]`))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	// Two torrents sharing the same tracker — should trigger only one API call.
	torrentList := &API.SliceInfo{
		{Name: "Torrent1", Hash: "hash1", Tracker: "http://tracker.example.com"},
		{Name: "Torrent2", Hash: "hash2", Tracker: "http://tracker.example.com"},
	}

	registry := prometheus.NewRegistry()
	getTrackers(torrentList, registry)

	if count := int(requestCount.Load()); count != 1 {
		t.Errorf("expected 1 API request after deduplication, got %d", count)
	}
}

// hasTrackerMetric is a local helper for qbit package tests.
func hasTrackerMetric(t *testing.T, registry *prometheus.Registry, name string) bool {
	t.Helper()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, mf := range families {
		if mf.GetName() == name && len(mf.GetMetric()) > 0 {
			return true
		}
	}

	return false
}
