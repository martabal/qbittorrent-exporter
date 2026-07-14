package prom

import (
	"bytes"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	API "qbit-exp/api"
	app "qbit-exp/app"
	"qbit-exp/logger"

	"github.com/VictoriaMetrics/metrics"
)

var buff = &bytes.Buffer{}

func init() {
	logger.Log = &logger.Logger{Logger: slog.New(slog.NewTextHandler(buff, &slog.HandlerOptions{}))} //nolint:exhaustruct
}

var username = "admin"
var password = "adminadmin"

func TestMain(t *testing.T) {
	t.Parallel()

	app.QBittorrent = app.QBittorrentSettings{
		BaseUrl: "http://localhost:8080",
		LegacyAuth: &app.LegacyAuth{
			Username: username,
			Password: password,
			Cookie: app.Cookie{
				Key:   "cookieKey",
				Value: nil,
			},
		},
		Timeout:             time.Duration(30) * time.Second,
		APIKey:              nil,
		FullRefreshInterval: 5,
		BasicAuth:           nil,
	}

	result := app.GetPasswordMasked(app.QBittorrent.LegacyAuth.Password)

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

	registry := metrics.NewSet()

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
			GlobalRatio:          globalRatio,
			UseAltSpeedLimits:    true,
			AlltimeDl:            100000,
			AlltimeUl:            100001,
			DlInfoData:           100002,
			UpInfoData:           100003,
			DlInfoSpeed:          100004,
			UpInfoSpeed:          100005,
			AverageTimeQueue:     0,
			ConnectionStatus:     "",
			DHTNodes:             0,
			FreeSpaceOnDisk:      0,
			QueuedIoJobs:         0,
			TotalBuffersSize:     0,
			TotalQueuedSize:      0,
			TotalPeerConnections: 0,
			TotalWastedSession:   0,
		},
		Tags: []string{"tag1", "tag2"},
		CategoryMap: map[string]API.Category{
			"cat1": {Name: "cat1", SavePath: ""},
			"cat2": {Name: "cat2", SavePath: ""},
		},
	}
}

func runMainDataTest(t *testing.T, data *API.MainData) {
	t.Helper()

	registry := metrics.NewSet()
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

	registry := metrics.NewSet()
	Version(&version, registry)

	families, labels := parseSetMetrics(t, registry)
	values := families["qbittorrent_app_version"]
	if len(values) == 0 {
		t.Fatal("Expected qBittorrent version metric to be registered")
	}
	if values[0] != 1.0 {
		t.Errorf("Expected gauge value to be 1.0, got %v", values[0])
	}
	foundVersionLabel := false
	for _, metricLabels := range labels["qbittorrent_app_version"] {
		if metricLabels["version"] == expectedVersion {
			foundVersionLabel = true
			break
		}
	}
	if !foundVersionLabel {
		t.Errorf("Expected label value to be '%s'", expectedVersion)
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
			Popularity:        3.25,
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
			Category:          "",
			Comment:           "",
			MaxRatio:          0,
			SavePath:          "",
			Tracker:           "",
		},
	}

	registry := metrics.NewSet()

	webuiversion := "2.11.2"

	Torrent(mockInfo, &webuiversion, registry)

	expectedMetrics := map[string]float64{
		"qbittorrent_torrent_eta":                      120,
		"qbittorrent_torrent_download_speed_bytes":     500000,
		"qbittorrent_torrent_upload_speed_bytes":       250000,
		"qbittorrent_torrent_progress":                 0.0315,
		"qbittorrent_torrent_popularity":               3.25,
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

	registry := metrics.NewSet()
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

func testMetrics(t *testing.T, expectedMetrics map[string]float64, registry *metrics.Set) {
	t.Helper()
	metricFamilies, _ := parseSetMetrics(t, registry)

	for name, expectedValue := range expectedMetrics {
		values := metricFamilies[name]
		if len(values) == 0 {
			t.Errorf("Expected metric %s not found in registry", name)
			continue
		}
		if values[0] != expectedValue {
			t.Errorf("Metric %s: expected %f, got %f", name, expectedValue, values[0])
		}
	}
}

func testMultipleMetrics(t *testing.T, multipleMetrics map[string][]string, registry *metrics.Set) {
	t.Helper()
	_, metricLabels := parseSetMetrics(t, registry)

	for name, labels := range multipleMetrics {
		for _, label := range labels {
			found := false

			for _, labels := range metricLabels[name] {
				for _, value := range labels {
					if value == label {
						found = true
						break
					}
				}
			}

			if !found {
				t.Errorf("Metric %s with label %s not found in the registry", name, label)
			}
		}
	}
}

func parseSetMetrics(t *testing.T, set *metrics.Set) (map[string][]float64, map[string][]map[string]string) {
	t.Helper()

	var output bytes.Buffer
	set.WritePrometheus(&output)

	families := make(map[string][]float64)
	labelsByFamily := make(map[string][]map[string]string)
	for _, line := range strings.Split(output.String(), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		value, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			t.Fatalf("Failed to parse metric value from line %q: %v", line, err)
		}

		nameAndLabels := parts[0]
		name := nameAndLabels
		labels := map[string]string{}
		if open := strings.IndexByte(nameAndLabels, '{'); open != -1 {
			name = nameAndLabels[:open]
			labelText := strings.TrimSuffix(nameAndLabels[open+1:], "}")
			labels = parseLabelPairs(labelText)
		}

		families[name] = append(families[name], value)
		labelsByFamily[name] = append(labelsByFamily[name], labels)
	}

	return families, labelsByFamily
}

func parseLabelPairs(raw string) map[string]string {
	labels := make(map[string]string)
	if raw == "" {
		return labels
	}

	for len(raw) > 0 {
		eq := strings.IndexByte(raw, '=')
		if eq == -1 {
			break
		}
		key := raw[:eq]
		raw = raw[eq+1:]
		if len(raw) == 0 {
			break
		}

		quotedEnd := 1
		escaped := false
		for quotedEnd < len(raw) {
			ch := raw[quotedEnd]
			if ch == '\\' && !escaped {
				escaped = true
				quotedEnd++
				continue
			}
			if ch == '"' && !escaped {
				break
			}
			escaped = false
			quotedEnd++
		}

		if quotedEnd >= len(raw) {
			break
		}

		value, err := strconv.Unquote(raw[:quotedEnd+1])
		if err == nil {
			labels[key] = value
		}

		raw = strings.TrimPrefix(raw[quotedEnd+1:], ",")
	}

	return labels
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
		Popularity:        3.25,
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
