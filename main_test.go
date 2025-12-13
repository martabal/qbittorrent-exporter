package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qbit-exp/app"
	"qbit-exp/logger"

	"github.com/prometheus/client_golang/prometheus"
)

var buff = &bytes.Buffer{}

func init() {
	logger.Log = &logger.Logger{Logger: slog.New(slog.NewTextHandler(buff, &slog.HandlerOptions{}))}
}

func TestMetricsFailureResponse(t *testing.T) {
	t.Parallel()

	retryCtx, cancel := context.WithCancel(context.WithoutCancel(t.Context()))
	defer cancel()

	req, err := http.NewRequestWithContext(retryCtx, http.MethodGet, "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	metrics(rec, req, func(_ *prometheus.Registry) error {
		return errors.New("mock error")
	})

	if status := rec.Code; status != http.StatusServiceUnavailable {
		t.Errorf("expected status code 503, got %d", status)
	}
}

func TestMetricsReturnMetric(t *testing.T) {
	t.Parallel()

	buff.Reset()

	opts := &slog.HandlerOptions{
		Level: logger.LevelTrace,
	}

	logger.Log = &logger.Logger{Logger: slog.New(slog.NewTextHandler(buff, opts))}

	retryCtx, cancel := context.WithCancel(context.WithoutCancel(t.Context()))
	defer cancel()

	req, err := http.NewRequestWithContext(retryCtx, http.MethodGet, "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.RemoteAddr = "127.0.0.1:80"

	rec := httptest.NewRecorder()

	metrics(rec, req, func(registry *prometheus.Registry) error {
		qbittorrent_app_version := prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "qbittorrent_app_version",
			Help: "The current qBittorrent version",
			ConstLabels: map[string]string{
				"version": string("1.0"),
			},
		})
		registry.MustRegister(qbittorrent_app_version)
		qbittorrent_app_version.Set(1)

		return nil
	})

	if status := rec.Code; status != http.StatusOK {
		t.Errorf("expected status code 200, got %d", status)
	}

	expectedBody := "# HELP qbittorrent_app_version The current qBittorrent version\n# TYPE qbittorrent_app_version gauge\nqbittorrent_app_version{version=\"1.0\"} 1\n"

	if rec.Body.String() != expectedBody {
		t.Errorf("expected \n%s, got \n%s", expectedBody, rec.Body.String())
	}

	traceMessage := "New request from"
	if !strings.Contains(buff.String(), traceMessage) {
		t.Errorf("expected %s, got %s", traceMessage, buff.String())
	}
}

func TestBasicAuth_Success(t *testing.T) {
	t.Parallel()

	buff.Reset()

	app.Exporter.BasicAuth = &app.BasicAuth{
		Username: "testuser",
		Password: "testpass",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	wrappedHandler := basicAuth(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.SetBasicAuth("testuser", "testpass")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("expected handler to be called")
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status code 200, got %d", rec.Code)
	}

	if rec.Body.String() != "success" {
		t.Errorf("expected body 'success', got %s", rec.Body.String())
	}
}

func TestBasicAuth_InvalidCredentials(t *testing.T) {
	t.Parallel()

	buff.Reset()

	opts := &slog.HandlerOptions{
		Level: logger.LevelWarn,
	}

	logger.Log = &logger.Logger{Logger: slog.New(slog.NewTextHandler(buff, opts))}

	app.Exporter.BasicAuth = &app.BasicAuth{
		Username: "testuser",
		Password: "testpass",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := basicAuth(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.SetBasicAuth("wronguser", "wrongpass")
	req.RemoteAddr = "127.0.0.1:12345"
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	if handlerCalled {
		t.Error("expected handler not to be called")
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status code 401, got %d", rec.Code)
	}

	if !strings.Contains(buff.String(), "Invalid auth") {
		t.Errorf("expected log to contain 'Invalid auth', got %s", buff.String())
	}
}

func TestBasicAuth_NoCredentials(t *testing.T) {
	t.Parallel()

	app.Exporter.BasicAuth = &app.BasicAuth{
		Username: "testuser",
		Password: "testpass",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := basicAuth(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	if handlerCalled {
		t.Error("expected handler not to be called")
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status code 401, got %d", rec.Code)
	}

	authHeader := rec.Header().Get("WWW-Authenticate")
	expectedHeader := `Basic realm="restricted", charset="UTF-8"`
	if authHeader != expectedHeader {
		t.Errorf("expected WWW-Authenticate header %q, got %q", expectedHeader, authHeader)
	}
}
