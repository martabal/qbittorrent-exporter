package qbit

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	API "qbit-exp/api"
	"qbit-exp/app"
	"qbit-exp/deltasync"

	"github.com/prometheus/client_golang/prometheus"
)

func resetQbitPackageState() {
	syncState = nil
	scrapeCount = 0
}

func setupQbitTestState() {
	resetQbitPackageState()
	setupMockApp()
	app.QBittorrent.BaseUrl = ""
	app.QBittorrent.FullRefreshInterval = 100
	app.Exporter.Features.EnableTracker = false
	app.Exporter.Features.EnableHighCardinality = false
	app.Exporter.ExperimentalFeatures.EnableLabelWithHash = false
	app.Exporter.ExperimentalFeatures.EnableLabelWithTags = false
}

func TestGetDataRetryOnForbidden(t *testing.T) {
	setupQbitTestState()
	t.Cleanup(resetQbitPackageState)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	calls := 0
	data := Data{
		URL:        "/test",
		HTTPMethod: http.MethodGet,
		Handle: func(_ []byte, _ *prometheus.Registry, _ *string) error {
			calls++

			return nil
		},
	}

	c := make(chan func() (bool, error), 1)

	getData(prometheus.NewRegistry(), &data, new(string), c)

	gotRetryFunc := <-c
	retry, err := gotRetryFunc()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !retry {
		t.Fatal("expected retry=true for forbidden request with legacy auth")
	}

	if calls != 0 {
		t.Fatalf("handler must not be called on retry response, got %d", calls)
	}
}

func TestGetDataHandleErrorIsReturned(t *testing.T) {
	setupQbitTestState()
	t.Cleanup(resetQbitPackageState)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	expected := errors.New("handler failed")
	data := Data{
		URL:        "/test",
		HTTPMethod: http.MethodGet,
		Handle: func(_ []byte, _ *prometheus.Registry, _ *string) error {
			return expected
		},
	}

	c := make(chan func() (bool, error), 1)

	getData(prometheus.NewRegistry(), &data, new(string), c)

	gotRetryFunc := <-c
	retry, err := gotRetryFunc()

	if retry {
		t.Fatal("expected retry=false")
	}

	if !errors.Is(err, expected) {
		t.Fatalf("expected handler error %v, got %v", expected, err)
	}
}

func TestGetDataSuccess(t *testing.T) {
	setupQbitTestState()
	t.Cleanup(resetQbitPackageState)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	calls := 0
	data := Data{
		URL:        "/test",
		HTTPMethod: http.MethodGet,
		Handle: func(body []byte, _ *prometheus.Registry, _ *string) error {
			calls++
			if string(body) != `{"ok":true}` {
				t.Fatalf("unexpected body %q", string(body))
			}

			return nil
		},
	}

	c := make(chan func() (bool, error), 1)
	getData(prometheus.NewRegistry(), &data, new(string), c)

	gotRetryFunc := <-c
	retry, err := gotRetryFunc()

	if retry || err != nil {
		t.Fatalf("expected success response, got retry=%v err=%v", retry, err)
	}

	if calls != 1 {
		t.Fatalf("expected handler to be called once, got %d", calls)
	}
}

func TestGetTrackersDeduplicatesTrackers(t *testing.T) {
	setupQbitTestState()
	t.Cleanup(resetQbitPackageState)

	requestedHashes := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/trackers" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		hash := r.URL.Query().Get("hash")
		requestedHashes[hash]++

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"msg":"ok","num_downloaded":1,"num_leeches":2,"num_peers":3,"num_seeds":4,"status":2,"tier":"1","url":"http://tracker.test/announce"}]`))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	torrents := API.SliceInfo{
		{Hash: "hash-1", Tracker: "http://same.tracker"},
		{Hash: "hash-2", Tracker: "http://same.tracker"},
		{Hash: "hash-3", Tracker: "http://other.tracker"},
	}

	registry := prometheus.NewRegistry()

	getTrackers(&torrents, registry)

	if len(requestedHashes) != 2 {
		t.Fatalf("expected 2 unique tracker requests, got %d (%v)", len(requestedHashes), requestedHashes)
	}

	if requestedHashes["hash-2"] != 0 {
		t.Fatalf("expected duplicated tracker hash not requested, got hash-2=%d", requestedHashes["hash-2"])
	}
}

func TestFetchDeltaMainDataAppliesState(t *testing.T) {
	setupQbitTestState()
	t.Cleanup(resetQbitPackageState)

	syncState = deltasync.NewState()
	app.QBittorrent.FullRefreshInterval = 2

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/sync/maindata" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		if gotRID := r.URL.Query().Get("rid"); gotRID != "0" {
			t.Fatalf("expected rid=0, got %s", gotRID)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"rid": 10,
			"full_update": true,
			"torrents": {
				"hash-a": {"name":"torrent-a","tracker":"http://tracker-a","state":"downloading"}
			},
			"categories": {"movies": {"name":"movies","savePath":"/downloads/movies"}},
			"tags": ["tag-a"],
			"server_state": {"global_ratio":"1.5"}
		}`))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	if err := fetchDeltaMainData(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := syncState.GetRID(); got != 10 {
		t.Fatalf("expected rid=10, got %d", got)
	}

	if got := syncState.TorrentCount(); got != 1 {
		t.Fatalf("expected 1 torrent in state, got %d", got)
	}
}

func TestFetchDeltaMainDataInvalidJSON(t *testing.T) {
	setupQbitTestState()
	t.Cleanup(resetQbitPackageState)

	syncState = deltasync.NewState()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	err := fetchDeltaMainData()
	if err == nil {
		t.Fatal("expected JSON unmarshal error")
	}
}

func TestAllRequestsSuccess(t *testing.T) {
	setupQbitTestState()
	t.Cleanup(resetQbitPackageState)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/app/webapiVersion":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("2.9.0"))
		case "/api/v2/sync/maindata":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"rid": 1,
				"full_update": true,
				"torrents": {
					"hash-a": {"name":"torrent-a","tracker":"http://tracker-a","state":"downloading"}
				},
				"categories": {},
				"tags": [],
				"server_state": {"global_ratio":"1.0"}
			}`))
		case "/api/v2/app/version":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("v4.6.0"))
		case "/api/v2/app/preferences":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"alt_dl_limit": 1,
				"alt_up_limit": 2,
				"dl_limit": 3,
				"max_active_downloads": 4,
				"max_active_torrents": 5,
				"max_active_uploads": 6,
				"up_limit": 7
			}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	registry := prometheus.NewRegistry()

	if err := AllRequests(registry); err != nil {
		t.Fatalf("expected successful scrape, got %v", err)
	}

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("unexpected gather error: %v", err)
	}

	if len(metrics) == 0 {
		t.Fatal("expected metrics to be registered")
	}
}

func TestAllRequestsReturnsStaticRequestError(t *testing.T) {
	setupQbitTestState()
	t.Cleanup(resetQbitPackageState)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/app/webapiVersion":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("2.9.0"))
		case "/api/v2/sync/maindata":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"rid":1,"full_update":true,"torrents":{},"categories":{},"tags":[],"server_state":{"global_ratio":"1.0"}}`))
		case "/api/v2/app/version":
			w.WriteHeader(http.StatusInternalServerError)
		case "/api/v2/app/preferences":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"alt_dl_limit": 0, "alt_up_limit": 0, "dl_limit": 0, "max_active_downloads": 0, "max_active_torrents": 0, "max_active_uploads": 0, "up_limit": 0}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	err := AllRequests(prometheus.NewRegistry())
	if err == nil {
		t.Fatal("expected error when a static request fails")
	}
}

func TestAllRequestsFullRefreshResetsState(t *testing.T) {
	setupQbitTestState()
	t.Cleanup(resetQbitPackageState)

	call := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/app/webapiVersion":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("2.9.0"))
		case "/api/v2/sync/maindata":
			w.WriteHeader(http.StatusOK)
			call++
			if call == 1 {
				_, _ = w.Write([]byte(`{
					"rid": 1,
					"full_update": true,
					"torrents": {"hash-a": {"name":"torrent-a"}},
					"categories": {},
					"tags": [],
					"server_state": {"global_ratio":"1.0"}
				}`))

				return
			}

			_, _ = w.Write([]byte(`{
				"rid": 2,
				"full_update": true,
				"torrents": {"hash-b": {"name":"torrent-b"}},
				"categories": {},
				"tags": [],
				"server_state": {"global_ratio":"1.0"}
			}`))
		case "/api/v2/app/version":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("v4.6.0"))
		case "/api/v2/app/preferences":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"alt_dl_limit": 0, "alt_up_limit": 0, "dl_limit": 0, "max_active_downloads": 0, "max_active_torrents": 0, "max_active_uploads": 0, "up_limit": 0}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL
	app.QBittorrent.FullRefreshInterval = 2

	if err := AllRequests(prometheus.NewRegistry()); err != nil {
		t.Fatalf("unexpected first scrape error: %v", err)
	}

	if err := AllRequests(prometheus.NewRegistry()); err != nil {
		t.Fatalf("unexpected second scrape error: %v", err)
	}

	torrents := syncState.GetTorrents()

	if len(torrents) != 1 || torrents[0].Name != "torrent-b" {
		raw, _ := json.Marshal(torrents)
		t.Fatalf("expected state reset with latest full sync torrent, got %s", string(raw))
	}
}

func TestAllRequestsReturnsErrorWhenVersionRequestFails(t *testing.T) {
	setupQbitTestState()
	t.Cleanup(resetQbitPackageState)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/app/webapiVersion" {
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		t.Fatalf("unexpected path %s", r.URL.Path)
	}))
	defer server.Close()

	app.QBittorrent.BaseUrl = server.URL

	err := AllRequests(prometheus.NewRegistry())
	if err == nil {
		t.Fatal("expected error from initial version request")
	}
}
