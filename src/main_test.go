package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"qbit-exp/logger"

	"github.com/prometheus/client_golang/prometheus"
)

var buff = &bytes.Buffer{}

func init() {
	logger.Log = &logger.Logger{Logger: slog.New(slog.NewTextHandler(buff, &slog.HandlerOptions{}))}
}

func TestMetricsFailureResponse(t *testing.T) {

	req, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()

	metrics(rec, req, func(registry *prometheus.Registry) error {
		return fmt.Errorf("mock error")
	})

	if status := rec.Code; status != http.StatusServiceUnavailable {
		t.Errorf("expected status code 503, got %d", status)
	}
}

func TestMetricsReturnMetric(t *testing.T) {

	buff.Reset()
	opts := &slog.HandlerOptions{
		Level: slog.Level(logger.Trace),
	}

	logger.Log = &logger.Logger{Logger: slog.New(slog.NewTextHandler(buff, opts))}

	req, err := http.NewRequest("GET", "/metrics", nil)
	req.RemoteAddr = "127.0.0.1:80"
	if err != nil {
		t.Fatal(err)
	}
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
