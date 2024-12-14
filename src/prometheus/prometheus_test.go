package prom

import (
	API "qbit-exp/api"
	app "qbit-exp/app"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

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

func TestIsValidURL(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"https://www.example.com", true},
		{"http://localhost:8080", true},
		{"ftp://ftp.example.com", true},
		{"not_a_url", false},
		{"www.example.com", false},
		{"file:///path/to/file", false},
		{"", false},
		{"http://invalid..url", true},
		{"https://[::1]", true},
		{"http://user:pass@www.example.com", true},
		{"https://www.example.com:8080/path", true},
		{"https://www.example.com?query=value", true},
		{"https://www.example.com#fragment", true},
		{"http://www.ex ample.com", false},
		{"http://www.example.com:10000", true},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := isValidURL(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s to be %v, but got %v", tc.input, tc.expected, result)
			}
		})
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
