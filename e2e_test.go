package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"qbit-exp/app"
	"qbit-exp/qbit"
)

// buildMockQBittorrentServer creates a minimal mock qBittorrent HTTP server that
// returns valid responses for all endpoints called by qbit.AllRequests.
func buildMockQBittorrentServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v2/app/webapiVersion", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("2.11.2"))
	})

	mux.HandleFunc("/api/v2/sync/maindata", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"rid": 1,
			"full_update": true,
			"torrents": {},
			"categories": {},
			"tags": [],
			"server_state": {
				"global_ratio": "1.0",
				"connection_status": "connected",
				"use_alt_speed_limits": false
			}
		}`))
	})

	mux.HandleFunc("/api/v2/app/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("v5.0.2"))
	})

	mux.HandleFunc("/api/v2/app/preferences", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"alt_dl_limit": 10000,
			"alt_up_limit": 10000,
			"dl_limit": 0,
			"max_active_downloads": 3,
			"max_active_torrents": 5,
			"max_active_uploads": 3,
			"up_limit": 0
		}`))
	})

	return httptest.NewServer(mux)
}

func TestE2E_FullScrapeSuccess(t *testing.T) {
	server := buildMockQBittorrentServer()
	defer server.Close()

	origBaseURL := app.QBittorrent.BaseUrl
	origCookie := app.QBittorrent.Cookie
	origTimeout := app.QBittorrent.Timeout
	origRefreshInterval := app.QBittorrent.FullRefreshInterval
	origFeatures := app.Exporter.Features

	t.Cleanup(func() {
		app.QBittorrent.BaseUrl = origBaseURL
		app.QBittorrent.Cookie = origCookie
		app.QBittorrent.Timeout = origTimeout
		app.QBittorrent.FullRefreshInterval = origRefreshInterval
		app.Exporter.Features = origFeatures
	})

	cookie := "test-session-cookie"
	app.QBittorrent.BaseUrl = server.URL
	app.QBittorrent.Cookie = &cookie
	app.QBittorrent.Timeout = 5 * time.Second
	app.QBittorrent.FullRefreshInterval = 100
	app.Exporter.Features.EnableTracker = false

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.RemoteAddr = "127.0.0.1:9000"
	rec := httptest.NewRecorder()

	metrics(rec, req, qbit.AllRequests)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d", rec.Code)
	}

	body := rec.Body.String()

	expectedMetrics := []string{
		"qbittorrent_global_torrents",
		"qbittorrent_app_version",
		"qbittorrent_global_ratio",
		"qbittorrent_global_max_active_downloads",
		"qbittorrent_transfer_connection_status",
	}

	for _, name := range expectedMetrics {
		if !strings.Contains(body, name) {
			t.Errorf("expected metric %s in /metrics output", name)
		}
	}
}

func TestE2E_QBittorrentDown(t *testing.T) {
	// Server that returns 503 for every request, simulating qBittorrent being down.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	origBaseURL := app.QBittorrent.BaseUrl
	origCookie := app.QBittorrent.Cookie
	origTimeout := app.QBittorrent.Timeout
	origRefreshInterval := app.QBittorrent.FullRefreshInterval

	t.Cleanup(func() {
		app.QBittorrent.BaseUrl = origBaseURL
		app.QBittorrent.Cookie = origCookie
		app.QBittorrent.Timeout = origTimeout
		app.QBittorrent.FullRefreshInterval = origRefreshInterval
	})

	app.QBittorrent.BaseUrl = server.URL
	app.QBittorrent.Cookie = nil // Force auth attempt which will fail.
	app.QBittorrent.Timeout = 5 * time.Second
	app.QBittorrent.FullRefreshInterval = 100

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	metrics(rec, req, qbit.AllRequests)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected HTTP 503 when qBittorrent is down, got %d", rec.Code)
	}
}

func TestE2E_RootRedirectsToMetrics(t *testing.T) {
	origPath := app.Exporter.Path

	t.Cleanup(func() {
		app.Exporter.Path = origPath
	})

	app.Exporter.Path = "/metrics"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	http.Redirect(rec, req, app.Exporter.Path, http.StatusFound)

	if rec.Code != http.StatusFound {
		t.Errorf("expected HTTP 302 redirect from /, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/metrics" {
		t.Errorf("expected redirect to /metrics, got %q", location)
	}
}

func TestE2E_BasicAuthProtectsMetrics(t *testing.T) {
	server := buildMockQBittorrentServer()
	defer server.Close()

	origBaseURL := app.QBittorrent.BaseUrl
	origCookie := app.QBittorrent.Cookie
	origTimeout := app.QBittorrent.Timeout
	origRefreshInterval := app.QBittorrent.FullRefreshInterval
	origExporterBasicAuth := app.Exporter.BasicAuth
	origFeatures := app.Exporter.Features

	t.Cleanup(func() {
		app.QBittorrent.BaseUrl = origBaseURL
		app.QBittorrent.Cookie = origCookie
		app.QBittorrent.Timeout = origTimeout
		app.QBittorrent.FullRefreshInterval = origRefreshInterval
		app.Exporter.BasicAuth = origExporterBasicAuth
		app.Exporter.Features = origFeatures
	})

	cookie := "test-session-cookie"
	app.QBittorrent.BaseUrl = server.URL
	app.QBittorrent.Cookie = &cookie
	app.QBittorrent.Timeout = 5 * time.Second
	app.QBittorrent.FullRefreshInterval = 100
	app.Exporter.Features.EnableTracker = false
	app.Exporter.BasicAuth = &app.BasicAuth{
		Username: "admin",
		Password: "secret",
	}

	metricsHandler := func(w http.ResponseWriter, req *http.Request) {
		metrics(w, req, qbit.AllRequests)
	}
	protected := basicAuth(metricsHandler)

	tests := []struct {
		name         string
		user         string
		pass         string
		expectedCode int
	}{
		{"valid credentials", "admin", "secret", http.StatusOK},
		{"wrong password", "admin", "wrong", http.StatusUnauthorized},
		{"no credentials", "", "", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/metrics", nil)

			if tt.user != "" || tt.pass != "" {
				req.SetBasicAuth(tt.user, tt.pass)
			}

			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, req)

			if rec.Code != tt.expectedCode {
				t.Errorf("expected HTTP %d, got %d", tt.expectedCode, rec.Code)
			}
		})
	}
}
