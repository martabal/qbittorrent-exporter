package prom

import (
	"bytes"
	"log/slog"
	API "qbit-exp/api"
	app "qbit-exp/app"
	"qbit-exp/logger"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var buff = &bytes.Buffer{}

func init() {
	logger.Log = &logger.Logger{Logger: slog.New(slog.NewTextHandler(buff, &slog.HandlerOptions{}))}
}

func TestMain(t *testing.T) {
	app.QBittorrent = app.QBittorrentSettings{
		BaseUrl:  "http://localhost:8080",
		Username: "admin",
		Password: "adminadmin",
		Timeout:  time.Duration(30) * time.Second,
	}

	result := app.GetPasswordMasked()

	if !isValidMaskedPassword(result) {
		t.Errorf("Invalid masked password. Expected only asterisks, got: %s", result)
	}
}

func isValidMaskedPassword(password string) bool {
	for _, char := range password {
		if char != '*' {
			return false
		}
	}
	return true
}

func TestPreference(t *testing.T) {

	mockPrefs := &API.Preferences{
		MaxActiveDownloads: 5,
		MaxActiveUploads:   3,
		MaxActiveTorrents:  10,
		DlLimit:            100000,
		UpLimit:            100001,
		AltDlLimit:         50000,
		AltUpLimit:         50001,
	}

	registry := prometheus.NewRegistry()

	Preference(mockPrefs, registry)

	expectedMetrics := map[string]float64{
		"qbittorrent_global_max_active_downloads":          5,
		"qbittorrent_global_max_active_uploads":            3,
		"qbittorrent_global_max_active_torrents":           10,
		"qbittorrent_global_download_rate_limit_bytes":     100000,
		"qbittorrent_global_upload_rate_limit_bytes":       100001,
		"qbittorrent_global_alt_download_rate_limit_bytes": 50000,
		"qbittorrent_global_alt_upload_rate_limit_bytes":   50001,
	}

	testMetrics(expectedMetrics, registry, t)
}

func TestMainData(t *testing.T) {
	mockMainData := &API.MainData{
		ServerState: API.ServerState{
			GlobalRatio:       "2.5",
			UseAltSpeedLimits: true,
			AlltimeDl:         100000,
			AlltimeUl:         100001,
			DlInfoData:        100002,
			UpInfoData:        100003,
			DlInfoSpeed:       100004,
			UpInfoSpeed:       100005,
		},
		Tags: []string{"tag1", "tag2"},
		CategoryMap: map[string]API.Category{
			"cat1": {Name: "cat1"},
			"cat2": {Name: "cat2"},
		},
	}

	registry := prometheus.NewRegistry()
	MainData(mockMainData, registry)

	expectedMetrics := map[string]float64{
		"qbittorrent_app_alt_rate_limits_enabled":     1.0,
		"qbittorrent_global_alltime_downloaded_bytes": 100000,
		"qbittorrent_global_alltime_uploaded_bytes":   100001,
		"qbittorrent_global_session_downloaded_bytes": 100002,
		"qbittorrent_global_session_uploaded_bytes":   100003,
		"qbittorrent_global_download_speed_bytes":     100004,
		"qbittorrent_global_upload_speed_bytes":       100005,
	}

	testMetrics(expectedMetrics, registry, t)

	tagMetrics := map[string][]string{
		"qbittorrent_global_tags": {"tag1", "tag2"},
	}
	testMultipleMetrics(tagMetrics, registry, t)

	categoryMetrics := map[string][]string{
		"qbittorrent_global_categories": {"cat1", "cat2"},
	}
	testMultipleMetrics(categoryMetrics, registry, t)
}
func TestTransfer(t *testing.T) {

	mockTransfer := &API.Transfer{
		DhtNodes:         5,
		ConnectionStatus: "connected",
	}

	registry := prometheus.NewRegistry()
	Transfer(mockTransfer, registry)

	expectedMetrics := map[string]float64{
		"qbittorrent_global_dht_nodes":           5,
		"qbittorrent_transfer_connection_status": 1,
	}
	testMetrics(expectedMetrics, registry, t)
}

func TestVersion(t *testing.T) {

	expectedVersion := "v5.0.2"
	version := []byte(expectedVersion)

	registry := prometheus.NewRegistry()
	Version(&version, registry)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	var found bool

	expectedMetricName := "qbittorrent_app_version"
	metricName := createMetricName(metricNameApp, "version")
	for _, m := range metrics {
		if m.GetName() == metricName {
			found = true
			if len(m.GetMetric()) == 0 {
				t.Fatal("Expected metrics to have at least one entry")
			}
			if m.GetMetric()[0].GetGauge().GetValue() != 1.0 {
				t.Errorf("Expected gauge value to be 1.0, got %v", m.GetMetric()[0].GetGauge().GetValue())
			}
			if len(m.GetMetric()[0].GetLabel()) == 0 || m.GetMetric()[0].GetLabel()[0].GetValue() != expectedVersion {
				t.Errorf("Expected label value to be '%s', got %v", expectedVersion, m.GetMetric()[0].GetLabel())
			}
			break
		}
	}

	if expectedMetricName != metricName {
		t.Errorf("Error with metric name, expected %s got %s", expectedMetricName, metricName)
	}

	if !found {
		t.Fatal("Expected qBittorrent version metric to be registered")
	}
}

func TestTorrent(t *testing.T) {

	mockInfo := &API.Info{
		{
			Name:              "Torrent",
			Hash:              "hash",
			Eta:               120,
			Dlspeed:           500000,
			Upspeed:           250000,
			Progress:          0.031524459437698917,
			TimeActive:        3600,
			NumSeeds:          10,
			NumLeechs:         5,
			Ratio:             1.5,
			AmountLeft:        1000000000,
			Size:              5000000000,
			DownloadedSession: 250000000,
			UploadedSession:   100000000,
			Downloaded:        1000000000,
			Uploaded:          500000000,
			State:             "stalledUP",
			Tags:              "tag1, tag2",
		},
	}

	registry := prometheus.NewRegistry()

	webuiversion := "2.11.2"

	Torrent(mockInfo, &webuiversion, registry)

	expectedMetrics := map[string]float64{
		"qbittorrent_torrent_eta":                      120,
		"qbittorrent_torrent_download_speed_bytes":     500000,
		"qbittorrent_torrent_upload_speed_bytes":       250000,
		"qbittorrent_torrent_progress":                 0.0315,
		"qbittorrent_torrent_time_active":              3600,
		"qbittorrent_torrent_seeders":                  10,
		"qbittorrent_torrent_leechers":                 5,
		"qbittorrent_torrent_ratio":                    1.5,
		"qbittorrent_torrent_amount_left_bytes":        1000000000,
		"qbittorrent_torrent_size_bytes":               5000000000,
		"qbittorrent_torrent_session_downloaded_bytes": 250000000,
		"qbittorrent_torrent_session_uploaded_bytes":   100000000,
		"qbittorrent_torrent_total_downloaded_bytes":   1000000000,
		"qbittorrent_torrent_total_uploaded_bytes":     500000000,
		"qbittorrent_global_torrents":                  1,
	}

	testMetrics(expectedMetrics, registry, t)
}

func TestTrackers(t *testing.T) {

	mockTrackers := []*API.Trackers{
		{
			{
				URL:           "http://tracker",
				NumDownloaded: 100,
				NumLeeches:    50,
				NumSeeds:      10,
				NumPeers:      60,
				Status:        1,
				Tier:          []byte("1"),
				Message:       "Active",
			},
		},
	}

	registry := prometheus.NewRegistry()
	Trackers(mockTrackers, registry)

	expectedMetrics := map[string]float64{
		"qbittorrent_tracker_downloaded": 100,
		"qbittorrent_tracker_leeches":    50,
		"qbittorrent_tracker_peers":      60,
		"qbittorrent_tracker_seeders":    10,
		"qbittorrent_tracker_status":     1,
		"qbittorrent_tracker_tier":       1,
	}

	testMetrics(expectedMetrics, registry, t)
}

func testMetrics(expectedMetrics map[string]float64, registry *prometheus.Registry, t *testing.T) {

	for name, expectedValue := range expectedMetrics {
		mf, err := registry.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		var actualValue float64
		found := false
		for _, metricFamily := range mf {
			if metricFamily.GetName() == name {
				for _, metric := range metricFamily.GetMetric() {
					actualValue = metric.GetGauge().GetValue()
					found = true
					break
				}
			}
		}

		if !found {
			t.Errorf("Metric %s not found in the registry", name)
			continue
		}

		if actualValue != expectedValue {
			t.Errorf("Metric %s: expected %f, got %f", name, expectedValue, actualValue)
		}
	}
}

func testMultipleMetrics(multipleMetrics map[string][]string, registry *prometheus.Registry, t *testing.T) {

	for name, labels := range multipleMetrics {
		mf, err := registry.Gather()
		if err != nil {
			t.Fatalf("Failed to gather metrics: %v", err)
		}

		for _, label := range labels {
			found := false
			for _, metricFamily := range mf {
				if metricFamily.GetName() == name {
					for _, metric := range metricFamily.GetMetric() {
						for _, lbl := range metric.GetLabel() {
							if lbl.GetValue() == label {
								found = true
								break
							}
						}
					}
				}
			}

			if !found {
				t.Errorf("Metric %s with label %s not found in the registry", name, label)
			}
		}
	}
}
