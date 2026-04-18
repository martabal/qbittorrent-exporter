package prom

import (
	"bytes"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	API "qbit-exp/api"
	app "qbit-exp/app"
	"qbit-exp/logger"

	"github.com/prometheus/client_golang/prometheus"
)

var buff = &bytes.Buffer{}

func init() {
	logger.Log = &logger.Logger{Logger: slog.New(slog.NewTextHandler(buff, &slog.HandlerOptions{}))}
}

func TestMain(t *testing.T) {
	t.Parallel()

	app.QBittorrent = app.QBittorrentSettings{
		BaseUrl:  "http://localhost:8080",
		Username: "admin",
		Password: "adminadmin",
		Timeout:  time.Duration(30) * time.Second,
	}

	result := app.GetPasswordMasked(app.QBittorrent.Password)

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
	t.Parallel()

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

	testMetrics(t, expectedMetrics, registry)
}

func createMockMainData(globalRatio string) *API.MainData {
	return &API.MainData{
		ServerState: API.ServerState{
			GlobalRatio:       globalRatio,
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
}

func runMainDataTest(t *testing.T, data *API.MainData) {
	t.Helper()

	registry := prometheus.NewRegistry()
	MainData(data, registry)

	expectedMetrics := map[string]float64{
		"qbittorrent_global_ratio":                      2.5,
		"qbittorrent_global_categories":                 1.0,
		"qbittorrent_global_tags":                       1.0,
		"qbittorrent_app_alt_rate_limits_enabled":       1.0,
		"qbittorrent_global_alltime_downloaded_bytes":   100000, //nolint:misspell
		"qbittorrent_global_alltime_uploaded_bytes":     100001, //nolint:misspell
		"qbittorrent_global_session_downloaded_bytes":   100002,
		"qbittorrent_global_session_uploaded_bytes":     100003,
		"qbittorrent_global_download_speed_bytes":       100004,
		"qbittorrent_global_upload_speed_bytes":         100005,
		"qbittorrent_global_dht_nodes":                  0.0,
		"qbittorrent_global_average_time_queue":         0.0,
		"qbittorrent_global_free_space_on_disk_bytes":   0.0,
		"qbittorrent_global_queued_io_jobs":             0.0,
		"qbittorrent_global_total_buffers_size_bytes":   0.0,
		"qbittorrent_global_total_peer_connections":     0.0,
		"qbittorrent_global_total_queued_size_bytes":    0.0,
		"qbittorrent_global_total_wasted_session_bytes": 0.0,
		"qbittorrent_transfer_connection_status":        1.0,
	}
	testMetrics(t, expectedMetrics, registry)

	tagMetrics := map[string][]string{
		"qbittorrent_global_tags": {"tag1", "tag2"},
	}
	testMultipleMetrics(t, tagMetrics, registry)

	categoryMetrics := map[string][]string{
		"qbittorrent_global_categories": {"cat1", "cat2"},
	}
	testMultipleMetrics(t, categoryMetrics, registry)
}

func TestMainDataMetrics(t *testing.T) {
	t.Parallel()
	runMainDataTest(t, createMockMainData("2.5"))
	runMainDataTest(t, createMockMainData("2,5"))
}

func TestVersion(t *testing.T) {
	t.Parallel()

	expectedVersion := "v5.0.2"
	version := []byte(expectedVersion)

	registry := prometheus.NewRegistry()
	Version(&version, registry)

	metrics, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	var found bool

	metricName := "qbittorrent_app_version"
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

	if !found {
		t.Fatal("Expected qBittorrent version metric to be registered")
	}
}

func TestTorrent(t *testing.T) {
	t.Parallel()

	mockInfo := &API.SliceInfo{
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
			AddedOn:           1664715487,
			CompletionOn:      1664719487,
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
		"qbittorrent_torrent_added_on":                 1664715487,
		"qbittorrent_torrent_completed_on":             1664719487,
		"qbittorrent_torrent_states":                   0,
		"qbittorrent_torrent_tags":                     1,
	}

	testMetrics(t, expectedMetrics, registry)
}

func TestTrackers(t *testing.T) {
	t.Parallel()

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

	testMetrics(t, expectedMetrics, registry)
}

func testMetrics(t *testing.T, expectedMetrics map[string]float64, registry *prometheus.Registry) {
	t.Helper()

	metricFamilies, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	discovered := make(map[string]bool)

	// Validate all expected metrics exist and match values
	for name, expectedValue := range expectedMetrics {
		found := false

		var actualValue float64

		for _, mf := range metricFamilies {
			if mf.GetName() != name {
				continue
			}

			if len(mf.GetMetric()) == 0 {
				t.Errorf("Metric family %s exists but has no metric instances", name)

				found = true

				break
			}

			m := mf.GetMetric()[0]

			if g := m.GetGauge(); g != nil {
				actualValue = g.GetValue()
				found = true

				break
			}

			t.Errorf("Metric %s exists but is not a gauge", name)

			found = true

			break
		}

		if !found {
			t.Errorf("Expected metric %s not found in registry", name)

			continue
		}

		discovered[name] = true

		if actualValue != expectedValue {
			t.Errorf("Metric %s: expected %f, got %f", name, expectedValue, actualValue)
		}
	}

	// Ensure registry does not contain unexpected metrics
	for _, mf := range metricFamilies {
		name := mf.GetName()

		if _, expected := expectedMetrics[name]; !expected {
			t.Errorf("Registry contains unexpected metric: %s", name)
		}
	}
}

// getMetricValue returns the gauge value of the first instance of the named
// metric family, and whether it was found.
func getMetricValue(t *testing.T, registry *prometheus.Registry, name string) (float64, bool) {
	t.Helper()

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	for _, mf := range families {
		if mf.GetName() == name && len(mf.GetMetric()) > 0 {
			if g := mf.GetMetric()[0].GetGauge(); g != nil {
				return g.GetValue(), true
			}
		}
	}

	return 0, false
}

// hasMetric returns true when the registry contains at least one observed
// instance of the named metric.
func hasMetric(t *testing.T, registry *prometheus.Registry, name string) bool {
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

func testMultipleMetrics(t *testing.T, multipleMetrics map[string][]string, registry *prometheus.Registry) {
	t.Helper()

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

func TestCreateTorrentInfoLabels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		enableHighCardinality bool
		enableLabelWithHash   bool
		enableLabelWithTags   bool
		expectedLabelsCount   int
		mustContainLabels     []string
		mustNotContainLabels  []string
	}{
		{
			name:                  "High cardinality disabled, no extra labels",
			enableHighCardinality: false,
			enableLabelWithHash:   false,
			enableLabelWithTags:   false,
			expectedLabelsCount:   8,
			mustContainLabels:     []string{"added_on", "category", "name", "state"},
			mustNotContainLabels:  []string{"hash", "tags"},
		},
		{
			name:                  "High cardinality enabled, no extra labels",
			enableHighCardinality: true,
			enableLabelWithHash:   false,
			enableLabelWithTags:   false,
			expectedLabelsCount:   23,
			mustContainLabels:     []string{"added_on", "category", "progress", "size"},
			mustNotContainLabels:  []string{"hash", "tags"},
		},
		{
			name:                  "High cardinality disabled with hash label",
			enableHighCardinality: false,
			enableLabelWithHash:   true,
			enableLabelWithTags:   false,
			expectedLabelsCount:   9,
			mustContainLabels:     []string{"hash", "name"},
			mustNotContainLabels:  []string{"tags"},
		},
		{
			name:                  "High cardinality disabled with tags label",
			enableHighCardinality: false,
			enableLabelWithHash:   false,
			enableLabelWithTags:   true,
			expectedLabelsCount:   9,
			mustContainLabels:     []string{"tags", "name"},
			mustNotContainLabels:  []string{"hash"},
		},
		{
			name:                  "High cardinality enabled with both extra labels",
			enableHighCardinality: true,
			enableLabelWithHash:   true,
			enableLabelWithTags:   true,
			expectedLabelsCount:   25,
			mustContainLabels:     []string{"hash", "tags", "progress"},
			mustNotContainLabels:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			labels := createTorrentInfoLabels(tt.enableHighCardinality, tt.enableLabelWithHash, tt.enableLabelWithTags)

			if len(labels) != tt.expectedLabelsCount {
				t.Errorf("expected %d labels, got %d: %v", tt.expectedLabelsCount, len(labels), labels)
			}

			for _, mustContain := range tt.mustContainLabels {
				found := slices.Contains(labels, mustContain)

				if !found {
					t.Errorf("expected labels to contain %q, but it was not found", mustContain)
				}
			}

			for _, mustNotContain := range tt.mustNotContainLabels {
				for _, label := range labels {
					if label == mustNotContain {
						t.Errorf("expected labels not to contain %q, but it was found", mustNotContain)
					}
				}
			}
		})
	}
}

func TestCreateTorrentLabels(t *testing.T) {
	t.Parallel()

	mockTorrent := API.Info{
		Name:              "Test Torrent",
		Hash:              "abc123hash",
		Category:          "movies",
		State:             "downloading",
		Tracker:           "http://tracker.example.com",
		Comment:           "test comment",
		SavePath:          "/downloads",
		AddedOn:           1234567890,
		CompletionOn:      1234567900,
		Size:              1000000000,
		Progress:          0.5,
		NumSeeds:          10,
		NumLeechs:         5,
		Dlspeed:           500000,
		Upspeed:           100000,
		AmountLeft:        500000000,
		TimeActive:        3600,
		Eta:               7200,
		Uploaded:          50000000,
		UploadedSession:   25000000,
		Downloaded:        500000000,
		DownloadedSession: 250000000,
		MaxRatio:          2.0,
		Ratio:             0.1,
		Tags:              "tag1, tag2",
	}

	tests := []struct {
		name                  string
		enableHighCardinality bool
		enableLabelWithHash   bool
		enableLabelWithTags   bool
		expectedLabelsCount   int
		mustContainLabels     map[string]string
		mustNotContainLabels  []string
	}{
		{
			name:                  "Basic labels only",
			enableHighCardinality: false,
			enableLabelWithHash:   false,
			enableLabelWithTags:   false,
			expectedLabelsCount:   8,
			mustContainLabels: map[string]string{
				"name":     "Test Torrent",
				"category": "movies",
				"state":    "downloading",
			},
			mustNotContainLabels: []string{"hash", "tags", "size"},
		},
		{
			name:                  "High cardinality enabled",
			enableHighCardinality: true,
			enableLabelWithHash:   false,
			enableLabelWithTags:   false,
			expectedLabelsCount:   23,
			mustContainLabels: map[string]string{
				"name":     "Test Torrent",
				"size":     "1000000000",
				"progress": "0.5000",
			},
			mustNotContainLabels: []string{"hash", "tags"},
		},
		{
			name:                  "With hash label",
			enableHighCardinality: false,
			enableLabelWithHash:   true,
			enableLabelWithTags:   false,
			expectedLabelsCount:   9,
			mustContainLabels: map[string]string{
				"hash": "abc123hash",
				"name": "Test Torrent",
			},
			mustNotContainLabels: []string{"tags"},
		},
		{
			name:                  "With tags label",
			enableHighCardinality: false,
			enableLabelWithHash:   false,
			enableLabelWithTags:   true,
			expectedLabelsCount:   9,
			mustContainLabels: map[string]string{
				"tags": "tag1, tag2",
				"name": "Test Torrent",
			},
			mustNotContainLabels: []string{"hash"},
		},
		{
			name:                  "All options enabled",
			enableHighCardinality: true,
			enableLabelWithHash:   true,
			enableLabelWithTags:   true,
			expectedLabelsCount:   25,
			mustContainLabels: map[string]string{
				"hash":     "abc123hash",
				"tags":     "tag1, tag2",
				"progress": "0.5000",
			},
			mustNotContainLabels: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			labels := createTorrentLabels(mockTorrent, tt.enableHighCardinality, tt.enableLabelWithHash, tt.enableLabelWithTags)

			if len(labels) != tt.expectedLabelsCount {
				t.Errorf("expected %d labels, got %d", tt.expectedLabelsCount, len(labels))
			}

			for key, expectedValue := range tt.mustContainLabels {
				if actualValue, exists := labels[key]; !exists {
					t.Errorf("expected label %q to exist", key)
				} else if actualValue != expectedValue {
					t.Errorf("label %q: expected %q, got %q", key, expectedValue, actualValue)
				}
			}

			for _, mustNotContain := range tt.mustNotContainLabels {
				if _, exists := labels[mustNotContain]; exists {
					t.Errorf("expected label %q not to exist, but it was found", mustNotContain)
				}
			}
		})
	}
}

func TestTorrentEmptyList(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	emptyList := &API.SliceInfo{}
	version := "2.11.2"

	Torrent(emptyList, &version, registry)

	val, found := getMetricValue(t, registry, "qbittorrent_global_torrents")
	if !found {
		t.Fatal("expected qbittorrent_global_torrents to be registered")
	}

	if val != 0 {
		t.Errorf("expected qbittorrent_global_torrents=0 for empty list, got %f", val)
	}
}

func TestTorrentLegacyWebUIVersion(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	mockInfo := &API.SliceInfo{
		{Name: "Paused Torrent", State: "pausedUP"},
	}
	version := "2.10.0" // < 2.11.0

	Torrent(mockInfo, &version, registry)

	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	var pausedUPFound, stoppedUPFound bool

	for _, mf := range families {
		if mf.GetName() != "qbittorrent_torrent_states" {
			continue
		}

		for _, m := range mf.GetMetric() {
			for _, lbl := range m.GetLabel() {
				switch lbl.GetValue() {
				case "pausedUP":
					pausedUPFound = true
				case "stoppedUP":
					stoppedUPFound = true
				}
			}
		}
	}

	if !pausedUPFound {
		t.Error("expected pausedUP state metric for WebUI version < 2.11.0")
	}

	if stoppedUPFound {
		t.Error("expected stoppedUP state metric to be absent for WebUI version < 2.11.0")
	}
}

func TestTorrentUnknownState(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	mockInfo := &API.SliceInfo{
		{Name: "Test", State: "customUnknownState123"},
	}
	version := "2.11.2"

	Torrent(mockInfo, &version, registry)

	if !strings.Contains(buff.String(), "Unknown state: customUnknownState123") {
		t.Errorf("expected error log 'Unknown state: customUnknownState123', got: %s", buff.String())
	}
}

func TestTorrentIncreasedCardinality(t *testing.T) {
	origIncreased := app.Exporter.Features.EnableIncreasedCardinality
	origHigh := app.Exporter.Features.EnableHighCardinality

	t.Cleanup(func() {
		app.Exporter.Features.EnableIncreasedCardinality = origIncreased
		app.Exporter.Features.EnableHighCardinality = origHigh
	})

	app.Exporter.Features.EnableIncreasedCardinality = true
	app.Exporter.Features.EnableHighCardinality = false

	registry := prometheus.NewRegistry()
	mockInfo := &API.SliceInfo{
		{
			Name:     "Test Torrent",
			State:    "downloading",
			Comment:  "some comment",
			SavePath: "/data",
		},
	}
	version := "2.11.2"

	Torrent(mockInfo, &version, registry)

	for _, name := range []string{
		"qbittorrent_torrent_state",
		"qbittorrent_torrent_comment",
		"qbittorrent_torrent_save_path",
		"qbittorrent_torrent_info",
	} {
		if !hasMetric(t, registry, name) {
			t.Errorf("expected metric %s with EnableIncreasedCardinality=true", name)
		}
	}
}

func TestTorrentLabelWithTracker(t *testing.T) {
	origLabelWithTracker := app.Exporter.ExperimentalFeatures.EnableLabelWithTracker

	t.Cleanup(func() {
		app.Exporter.ExperimentalFeatures.EnableLabelWithTracker = origLabelWithTracker
	})

	app.Exporter.ExperimentalFeatures.EnableLabelWithTracker = true

	registry := prometheus.NewRegistry()
	mockInfo := &API.SliceInfo{
		{
			Name:    "Test Torrent",
			Tracker: "http://tracker.example.com",
		},
	}
	version := "2.11.2"

	Torrent(mockInfo, &version, registry)

	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	var trackerLabelFound bool

	for _, mf := range families {
		if mf.GetName() != "qbittorrent_torrent_eta" {
			continue
		}

		for _, m := range mf.GetMetric() {
			for _, lbl := range m.GetLabel() {
				if lbl.GetName() == "tracker" {
					trackerLabelFound = true
				}
			}
		}
	}

	if !trackerLabelFound {
		t.Error("expected 'tracker' label on torrent metrics when EnableLabelWithTracker=true")
	}
}

func TestTrackersEmpty(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	Trackers([]*API.Trackers{}, registry)

	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	for _, mf := range families {
		if strings.HasPrefix(mf.GetName(), "qbittorrent_tracker_") {
			t.Errorf("unexpected tracker metric %s for empty input", mf.GetName())
		}
	}
}

func TestTrackersHighCardinality(t *testing.T) {
	origHigh := app.Exporter.Features.EnableHighCardinality

	t.Cleanup(func() {
		app.Exporter.Features.EnableHighCardinality = origHigh
	})

	app.Exporter.Features.EnableHighCardinality = true

	mockTrackers := []*API.Trackers{
		{
			{
				URL:           "http://tracker.example.com",
				NumDownloaded: 10,
				NumLeeches:    2,
				NumSeeds:      5,
				NumPeers:      7,
				Status:        2,
				Tier:          []byte("0"),
				Message:       "working",
			},
		},
	}

	registry := prometheus.NewRegistry()
	Trackers(mockTrackers, registry)

	if !hasMetric(t, registry, "qbittorrent_tracker_info") {
		t.Error("expected qbittorrent_tracker_info metric with EnableHighCardinality=true")
	}
}

func TestTrackersInvalidURL(t *testing.T) {
	t.Parallel()

	mockTrackers := []*API.Trackers{
		{
			{
				URL:    "not-a-valid-url",
				Status: 2,
				Tier:   []byte("0"),
			},
			{
				URL:    "http://valid.tracker.com",
				Status: 1,
				Tier:   []byte("0"),
			},
		},
	}

	registry := prometheus.NewRegistry()
	Trackers(mockTrackers, registry)

	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	for _, mf := range families {
		if mf.GetName() != "qbittorrent_tracker_status" {
			continue
		}

		if count := len(mf.GetMetric()); count != 1 {
			t.Errorf("expected 1 tracker_status instance (invalid URL filtered), got %d", count)
		}
	}
}

func TestTrackersNonNumericTier(t *testing.T) {
	t.Parallel()

	mockTrackers := []*API.Trackers{
		{
			{
				URL:  "http://tracker.example.com",
				Tier: []byte("abc"),
			},
		},
	}

	registry := prometheus.NewRegistry()
	Trackers(mockTrackers, registry)

	families, err := registry.Gather()
	if err != nil {
		t.Fatal(err)
	}

	for _, mf := range families {
		if mf.GetName() != "qbittorrent_tracker_tier" {
			continue
		}

		if len(mf.GetMetric()) == 0 {
			t.Fatal("expected qbittorrent_tracker_tier to have an instance")
		}

		if val := mf.GetMetric()[0].GetGauge().GetValue(); val != 0 {
			t.Errorf("expected tier=0 for non-numeric value, got %f", val)
		}
	}
}

func TestMainDataInvalidRatio(t *testing.T) {
	t.Parallel()

	data := &API.MainData{
		ServerState: API.ServerState{
			GlobalRatio: "not-a-valid-ratio",
		},
	}

	registry := prometheus.NewRegistry()
	MainData(data, registry)

	if hasMetric(t, registry, "qbittorrent_global_ratio") {
		t.Error("expected qbittorrent_global_ratio to be absent for invalid ratio string")
	}
}

func TestMainDataAltSpeedLimitsDisabled(t *testing.T) {
	t.Parallel()

	data := &API.MainData{
		ServerState: API.ServerState{
			GlobalRatio:       "1.5",
			UseAltSpeedLimits: false,
		},
	}

	registry := prometheus.NewRegistry()
	MainData(data, registry)

	val, found := getMetricValue(t, registry, "qbittorrent_app_alt_rate_limits_enabled")
	if !found {
		t.Fatal("expected qbittorrent_app_alt_rate_limits_enabled to be registered")
	}

	if val != 0 {
		t.Errorf("expected qbittorrent_app_alt_rate_limits_enabled=0 when disabled, got %f", val)
	}
}
