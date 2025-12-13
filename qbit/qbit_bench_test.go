package qbit

import (
	"testing"

	API "qbit-exp/api"
)

// BenchmarkUniqueTrackerBuilding benchmarks unique tracker building
func BenchmarkUniqueTrackerBuilding(b *testing.B) {
	// Create sample torrents with trackers
	torrents := make(API.SliceInfo, 100)
	for i := 0; i < 100; i++ {
		torrents[i] = API.Info{
			Hash:    "hash" + string(rune(i)),
			Tracker: "http://tracker" + string(rune(i%10)) + ".example.com",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uniqueValues := make(map[string]struct{})
		var uniqueTrackers []UniqueTracker

		for _, obj := range torrents {
			if _, exists := uniqueValues[obj.Tracker]; !exists {
				uniqueValues[obj.Tracker] = struct{}{}
				uniqueTrackers = append(uniqueTrackers, UniqueTracker{Tracker: obj.Tracker, Hash: obj.Hash})
			}
		}
	}
}

// BenchmarkCreateUrl benchmarks URL creation
func BenchmarkCreateUrl(b *testing.B) {
	url := "/api/v2/torrents/info"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = createUrl(url)
	}
}
