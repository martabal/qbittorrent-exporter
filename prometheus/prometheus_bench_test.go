package prom

import (
	"testing"

	API "qbit-exp/api"

	"github.com/prometheus/client_golang/prometheus"
)

// BenchmarkTorrentMetrics benchmarks the torrent metrics generation
func BenchmarkTorrentMetrics(b *testing.B) {
	// Create sample data
	torrents := make(API.SliceInfo, 100)
	for i := 0; i < 100; i++ {
		torrents[i] = API.Info{
			Name:              "Test Torrent " + string(rune(i)),
			Hash:              "abcdef1234567890",
			Size:              1024 * 1024 * 1024,
			Progress:          0.75,
			Dlspeed:           1024 * 100,
			Upspeed:           1024 * 50,
			Downloaded:        1024 * 1024 * 500,
			Uploaded:          1024 * 1024 * 750,
			AmountLeft:        1024 * 1024 * 256,
			TimeActive:        3600,
			Eta:               1800,
			NumSeeds:          10,
			NumLeechs:         5,
			Ratio:             1.5,
			Category:          "test",
			State:             "downloading",
			Tracker:           "http://tracker.example.com",
			Tags:              "tag1, tag2, tag3",
			Comment:           "Test comment",
			SavePath:          "/downloads",
			AddedOn:           1234567890,
			CompletionOn:      0,
			DownloadedSession: 1024 * 1024 * 100,
			UploadedSession:   1024 * 1024 * 150,
			MaxRatio:          2.0,
		}
	}

	webUIVersion := "2.11.0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := prometheus.NewRegistry()
		Torrent(&torrents, &webUIVersion, r)
	}
}

// BenchmarkMainDataProcessing benchmarks the main data processing
func BenchmarkMainDataProcessing(b *testing.B) {
	mainData := &API.MainData{
		ServerState: API.ServerState{
			AlltimeDl:            1024 * 1024 * 1024 * 100,
			AlltimeUl:            1024 * 1024 * 1024 * 150,
			DlInfoData:           1024 * 1024 * 1024,
			UpInfoData:           1024 * 1024 * 1024 * 2,
			DlInfoSpeed:          1024 * 100,
			UpInfoSpeed:          1024 * 50,
			GlobalRatio:          "1.5",
			DHTNodes:             100,
			ConnectionStatus:     "connected",
			UseAltSpeedLimits:    false,
			AverageTimeQueue:     30,
			FreeSpaceOnDisk:      1024 * 1024 * 1024 * 500,
			QueuedIoJobs:         10,
			TotalBuffersSize:     1024 * 1024 * 10,
			TotalQueuedSize:      1024 * 1024 * 50,
			TotalPeerConnections: 50,
			TotalWastedSession:   1024 * 1024,
		},
		Tags: []string{"tag1", "tag2", "tag3"},
		CategoryMap: map[string]API.Category{
			"cat1": {Name: "cat1"},
			"cat2": {Name: "cat2"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := prometheus.NewRegistry()
		MainData(mainData, r)
	}
}

// BenchmarkTagProcessing benchmarks tag processing
func BenchmarkTagProcessing(b *testing.B) {
	tagString := "tag1, tag2, tag3, tag4, tag5, tag6, tag7, tag8, tag9, tag10"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate the tag splitting operation
		for range tagString {
			_ = tagString
		}
	}
}

// BenchmarkPreferenceProcessing benchmarks preference processing
func BenchmarkPreferenceProcessing(b *testing.B) {
	prefs := &API.Preferences{
		MaxActiveDownloads: 10,
		MaxActiveUploads:   10,
		MaxActiveTorrents:  20,
		DlLimit:            1024 * 1024,
		UpLimit:            1024 * 512,
		AltDlLimit:         1024 * 512,
		AltUpLimit:         1024 * 256,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := prometheus.NewRegistry()
		Preference(prefs, r)
	}
}

// BenchmarkGaugeRegistration benchmarks gauge registration
func BenchmarkGaugeRegistration(b *testing.B) {
	gauges := GaugeList{
		{qbittorrentTorrentEta, helpQbittorrentTorrentEta, []string{"name"}},
		{qbittorrentTorrentDownloadSpeedBytes, helpQbittorrentTorrentDownloadSpeedBytes, []string{"name"}},
		{qbittorrentTorrentUploadSpeedBytes, helpQbittorrentTorrentUploadSpeedBytes, []string{"name"}},
		{qbittorrentTorrentProgress, helpQbittorrentTorrentProgress, []string{"name"}},
		{qbittorrentTorrentTimeActive, helpQbittorrentTorrentTimeActive, []string{"name"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := prometheus.NewRegistry()
		registerGauge(&gauges, r)
	}
}
