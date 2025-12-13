package API

import (
	"encoding/json"
	"testing"
)

func TestInfoUnmarshal(t *testing.T) {
	t.Parallel()

	jsonData := `{
		"amount_left": 1000000,
		"added_on": 1234567890,
		"category": "movies",
		"comment": "test comment",
		"completion_on": 1234567900,
		"dlspeed": 500000,
		"downloaded": 10000000,
		"downloaded_session": 5000000,
		"eta": 3600,
		"hash": "abc123",
		"max_ratio": 2.5,
		"name": "Test Torrent",
		"num_leechs": 10,
		"num_seeds": 50,
		"progress": 0.75,
		"ratio": 1.5,
		"save_path": "/downloads",
		"size": 50000000,
		"state": "downloading",
		"tags": "tag1, tag2",
		"tracker": "http://tracker.example.com",
		"time_active": 7200,
		"uploaded": 15000000,
		"uploaded_session": 7500000,
		"upspeed": 250000
	}`

	var info Info

	err := json.Unmarshal([]byte(jsonData), &info)
	if err != nil {
		t.Fatalf("Failed to unmarshal Info: %v", err)
	}

	if info.Name != "Test Torrent" {
		t.Errorf("Name: expected %q, got %q", "Test Torrent", info.Name)
	}

	if info.Hash != "abc123" {
		t.Errorf("Hash: expected %q, got %q", "abc123", info.Hash)
	}

	if info.Size != 50000000 {
		t.Errorf("Size: expected %d, got %d", 50000000, info.Size)
	}

	if info.Progress != 0.75 {
		t.Errorf("Progress: expected %f, got %f", 0.75, info.Progress)
	}

	if info.State != "downloading" {
		t.Errorf("State: expected %q, got %q", "downloading", info.State)
	}
}

func TestSliceInfoUnmarshal(t *testing.T) {
	t.Parallel()

	jsonData := `[
		{
			"amount_left": 1000000,
			"name": "Torrent 1",
			"hash": "hash1",
			"size": 10000000,
			"progress": 0.5,
			"state": "downloading"
		},
		{
			"amount_left": 0,
			"name": "Torrent 2",
			"hash": "hash2",
			"size": 20000000,
			"progress": 1.0,
			"state": "uploading"
		}
	]`

	var sliceInfo SliceInfo

	err := json.Unmarshal([]byte(jsonData), &sliceInfo)
	if err != nil {
		t.Fatalf("Failed to unmarshal SliceInfo: %v", err)
	}

	if len(sliceInfo) != 2 {
		t.Fatalf("Expected 2 torrents, got %d", len(sliceInfo))
	}

	if sliceInfo[0].Name != "Torrent 1" {
		t.Errorf("First torrent name: expected %q, got %q", "Torrent 1", sliceInfo[0].Name)
	}

	if sliceInfo[1].State != "uploading" {
		t.Errorf("Second torrent state: expected %q, got %q", "uploading", sliceInfo[1].State)
	}
}

func TestPreferencesUnmarshal(t *testing.T) {
	t.Parallel()

	jsonData := `{
		"alt_dl_limit": 100000,
		"alt_up_limit": 50000,
		"dl_limit": 200000,
		"max_active_downloads": 5,
		"max_active_torrents": 10,
		"max_active_uploads": 3,
		"up_limit": 100000
	}`

	var prefs Preferences

	err := json.Unmarshal([]byte(jsonData), &prefs)
	if err != nil {
		t.Fatalf("Failed to unmarshal Preferences: %v", err)
	}

	if prefs.MaxActiveDownloads != 5 {
		t.Errorf("MaxActiveDownloads: expected %d, got %d", 5, prefs.MaxActiveDownloads)
	}

	if prefs.DlLimit != 200000 {
		t.Errorf("DlLimit: expected %d, got %d", 200000, prefs.DlLimit)
	}

	if prefs.UpLimit != 100000 {
		t.Errorf("UpLimit: expected %d, got %d", 100000, prefs.UpLimit)
	}
}

func TestMainDataUnmarshal(t *testing.T) {
	t.Parallel()

	jsonData := `{
		"categories": {
			"movies": {
				"name": "movies",
				"savePath": "/downloads/movies"
			},
			"music": {
				"name": "music",
				"savePath": "/downloads/music"
			}
		},
		"server_state": {
			"all-time_dl": 1000000000,
			"all-time_ul": 500000000,
			"average_time_queue": 100,
			"connection_status": "connected",
			"dht_nodes": 500,
			"dl_info_data": 50000000,
			"dl_info_speed": 1000000,
			"free_space_on_disk": 100000000000,
			"global_ratio": "0.5",
			"queued_io_jobs": 5,
			"up_info_data": 25000000,
			"up_info_speed": 500000,
			"total_buffers_size": 1000000,
			"total_queued_size": 2000000,
			"total_peer_connections": 100,
			"total_wasted_session": 1000,
			"use_alt_speed_limits": true
		},
		"tags": ["tag1", "tag2", "tag3"]
	}`

	var mainData MainData

	err := json.Unmarshal([]byte(jsonData), &mainData)
	if err != nil {
		t.Fatalf("Failed to unmarshal MainData: %v", err)
	}

	if len(mainData.CategoryMap) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(mainData.CategoryMap))
	}

	if mainData.CategoryMap["movies"].Name != "movies" {
		t.Errorf("Movies category name: expected %q, got %q", "movies", mainData.CategoryMap["movies"].Name)
	}

	if mainData.ServerState.GlobalRatio != "0.5" {
		t.Errorf("GlobalRatio: expected %q, got %q", "0.5", mainData.ServerState.GlobalRatio)
	}

	if !mainData.ServerState.UseAltSpeedLimits {
		t.Error("UseAltSpeedLimits: expected true, got false")
	}

	if len(mainData.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(mainData.Tags))
	}
}

func TestTrackersUnmarshal(t *testing.T) {
	t.Parallel()

	jsonData := `[
		{
			"msg": "Working",
			"num_downloaded": 100,
			"num_leeches": 10,
			"num_peers": 50,
			"num_seeds": 40,
			"status": 2,
			"tier": 0,
			"url": "http://tracker1.example.com"
		},
		{
			"msg": "Not contacted",
			"num_downloaded": 0,
			"num_leeches": 0,
			"num_peers": 0,
			"num_seeds": 0,
			"status": 0,
			"tier": 1,
			"url": "http://tracker2.example.com"
		}
	]`

	var trackers Trackers

	err := json.Unmarshal([]byte(jsonData), &trackers)
	if err != nil {
		t.Fatalf("Failed to unmarshal Trackers: %v", err)
	}

	if len(trackers) != 2 {
		t.Fatalf("Expected 2 trackers, got %d", len(trackers))
	}

	if trackers[0].URL != "http://tracker1.example.com" {
		t.Errorf("First tracker URL: expected %q, got %q", "http://tracker1.example.com", trackers[0].URL)
	}

	if trackers[0].Status != 2 {
		t.Errorf("First tracker status: expected %d, got %d", 2, trackers[0].Status)
	}

	if trackers[1].NumPeers != 0 {
		t.Errorf("Second tracker NumPeers: expected %d, got %d", 0, trackers[1].NumPeers)
	}
}

func TestServerStateUnmarshal(t *testing.T) {
	t.Parallel()

	//nolint:misspell
	jsonData := `{
		"alltime_dl": 1234567890,
		"alltime_ul": 9876543210, 
		"average_time_queue": 50,
		"connection_status": "firewalled",
		"dht_nodes": 1000,
		"dl_info_data": 100000000,
		"dl_info_speed": 2000000,
		"free_space_on_disk": 500000000000,
		"global_ratio": "2.5",
		"queued_io_jobs": 10,
		"up_info_data": 250000000,
		"up_info_speed": 1500000,
		"total_buffers_size": 5000000,
		"total_queued_size": 10000000,
		"total_peer_connections": 200,
		"total_wasted_session": 5000,
		"use_alt_speed_limits": false
	}`

	var serverState ServerState

	err := json.Unmarshal([]byte(jsonData), &serverState)
	if err != nil {
		t.Fatalf("Failed to unmarshal ServerState: %v", err)
	}

	if serverState.AlltimeDl != 1234567890 {
		t.Errorf("AlltimeDl: expected %d, got %d", 1234567890, serverState.AlltimeDl)
	}

	if serverState.GlobalRatio != "2.5" {
		t.Errorf("GlobalRatio: expected %q, got %q", "2.5", serverState.GlobalRatio)
	}

	if serverState.ConnectionStatus != "firewalled" {
		t.Errorf("ConnectionStatus: expected %q, got %q", "firewalled", serverState.ConnectionStatus)
	}

	if serverState.UseAltSpeedLimits {
		t.Error("UseAltSpeedLimits: expected false, got true")
	}
}

func TestCategoryUnmarshal(t *testing.T) {
	t.Parallel()

	jsonData := `{
		"name": "downloads",
		"savePath": "/home/user/downloads"
	}`

	var category Category

	err := json.Unmarshal([]byte(jsonData), &category)
	if err != nil {
		t.Fatalf("Failed to unmarshal Category: %v", err)
	}

	if category.Name != "downloads" {
		t.Errorf("Name: expected %q, got %q", "downloads", category.Name)
	}

	if category.SavePath != "/home/user/downloads" {
		t.Errorf("SavePath: expected %q, got %q", "/home/user/downloads", category.SavePath)
	}
}

func TestConstantValues(t *testing.T) {
	t.Parallel()

	// Test constant values are as expected
	if QbittorrentTimeOut != "qBittorrent is timing out" {
		t.Errorf("QbittorrentTimeOut: expected %q, got %q", "qBittorrent is timing out", QbittorrentTimeOut)
	}

	if ErrorWithUrl != "Error with url" {
		t.Errorf("ErrorWithUrl: expected %q, got %q", "Error with url", ErrorWithUrl)
	}

	if ErrorConnect != "Can't connect to qBittorrent" {
		t.Errorf("ErrorConnect: expected %q, got %q", "Can't connect to qBittorrent", ErrorConnect)
	}
}
