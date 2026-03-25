package deltasync

import (
	"encoding/json"
	"testing"

	API "qbit-exp/api"
)

const (
	torrent1Name  = "Torrent 1"
	torrent2Name  = "Torrent 2"
	stateDownload = "downloading"
	stateSeeding  = "seeding"
)

// raw is a helper to create json.RawMessage from a value.
func raw(v any) json.RawMessage {
	b, _ := json.Marshal(v)

	return b
}

func TestNewState(t *testing.T) {
	t.Parallel()

	state := NewState()

	if state == nil {
		t.Fatal("NewState returned nil")
	}

	if state.GetRID() != 0 {
		t.Errorf("Initial RID: expected 0, got %d", state.GetRID())
	}

	if state.TorrentCount() != 0 {
		t.Errorf("Initial torrent count: expected 0, got %d", state.TorrentCount())
	}

	torrents := state.GetTorrents()
	if len(torrents) != 0 {
		t.Errorf("Initial torrents: expected empty, got %d", len(torrents))
	}
}

func TestState_ApplyFullUpdate(t *testing.T) {
	t.Parallel()

	state := NewState()

	delta := &API.DeltaMainData{
		Rid:        100,
		FullUpdate: true,
		Torrents: map[string]json.RawMessage{
			"hash1": raw(map[string]any{"name": torrent1Name, "state": stateDownload, "progress": 0.5}),
			"hash2": raw(map[string]any{"name": torrent2Name, "state": stateSeeding, "progress": 1.0}),
		},
		Categories: map[string]API.Category{
			"movies": {Name: "movies", SavePath: "/downloads/movies"},
		},
		Tags:        []string{"tag1", "tag2"},
		ServerState: raw(map[string]any{"dht_nodes": 500, "dl_info_speed": 1000000}),
	}

	state.Apply(delta)

	if state.GetRID() != 100 {
		t.Errorf("RID: expected 100, got %d", state.GetRID())
	}

	if state.TorrentCount() != 2 {
		t.Errorf("Torrent count: expected 2, got %d", state.TorrentCount())
	}

	torrents := state.GetTorrents()

	torrentMap := make(map[string]API.Info)
	for _, torrent := range torrents {
		torrentMap[torrent.Hash] = torrent
	}

	if torrent, ok := torrentMap["hash1"]; !ok {
		t.Error("hash1 not found")
	} else if torrent.Name != torrent1Name {
		t.Errorf("hash1 name: expected %q, got %q", torrent1Name, torrent.Name)
	} else if torrent.State != stateDownload {
		t.Errorf("hash1 state: expected %q, got %q", stateDownload, torrent.State)
	}

	if torrent, ok := torrentMap["hash2"]; !ok {
		t.Error("hash2 not found")
	} else if torrent.Name != torrent2Name {
		t.Errorf("hash2 name: expected %q, got %q", torrent2Name, torrent.Name)
	}

	mainData := state.GetMainData()
	if len(mainData.CategoryMap) != 1 {
		t.Errorf("Categories: expected 1, got %d", len(mainData.CategoryMap))
	}

	if len(mainData.Tags) != 2 {
		t.Errorf("Tags: expected 2, got %d", len(mainData.Tags))
	}

	if mainData.ServerState.DHTNodes != 500 {
		t.Errorf("DHTNodes: expected 500, got %d", mainData.ServerState.DHTNodes)
	}
}

func TestState_ApplyDeltaUpdate(t *testing.T) {
	t.Parallel()

	state := NewState()

	// Initial full update
	initialDelta := &API.DeltaMainData{
		Rid:        100,
		FullUpdate: true,
		Torrents: map[string]json.RawMessage{
			"hash1": raw(map[string]any{
				"name": torrent1Name, "state": stateDownload,
				"progress": 0.5, "dlspeed": 1000000,
			}),
		},
		Categories: map[string]API.Category{
			"movies": {Name: "movies", SavePath: "/downloads/movies"},
		},
		Tags:        []string{"tag1"},
		ServerState: raw(map[string]any{"dht_nodes": 500}),
	}
	state.Apply(initialDelta)

	// Delta update: change state and dlspeed, add new torrent
	deltaDelta := &API.DeltaMainData{
		Rid:        101,
		FullUpdate: false,
		Torrents: map[string]json.RawMessage{
			"hash1": raw(map[string]any{"state": stateSeeding, "dlspeed": 0, "progress": 1.0}),
			"hash2": raw(map[string]any{"name": torrent2Name, "state": stateDownload}),
		},
		ServerState: raw(map[string]any{"dht_nodes": 600}),
	}
	state.Apply(deltaDelta)

	if state.GetRID() != 101 {
		t.Errorf("RID: expected 101, got %d", state.GetRID())
	}

	if state.TorrentCount() != 2 {
		t.Errorf("Torrent count: expected 2, got %d", state.TorrentCount())
	}

	torrents := state.GetTorrents()

	torrentMap := make(map[string]API.Info)
	for _, torrent := range torrents {
		torrentMap[torrent.Hash] = torrent
	}

	// hash1 should have merged values
	torrent1, ok := torrentMap["hash1"]
	if !ok {
		t.Fatal("hash1 not found")
	}

	if torrent1.Name != torrent1Name {
		t.Errorf("hash1 name should be preserved: expected %q, got %q", torrent1Name, torrent1.Name)
	}

	if torrent1.State != stateSeeding {
		t.Errorf("hash1 state should be updated: expected %q, got %q", stateSeeding, torrent1.State)
	}

	if torrent1.Dlspeed != 0 {
		t.Errorf("hash1 dlspeed should be updated: expected 0, got %d", torrent1.Dlspeed)
	}

	if torrent1.Progress != 1.0 {
		t.Errorf("hash1 progress should be updated: expected 1.0, got %f", torrent1.Progress)
	}

	// hash2 should exist as new torrent
	torrent2, ok := torrentMap["hash2"]
	if !ok {
		t.Fatal("hash2 not found")
	}

	if torrent2.Name != torrent2Name {
		t.Errorf("hash2 name: expected %q, got %q", torrent2Name, torrent2.Name)
	}

	// Server state should be updated
	mainData := state.GetMainData()
	if mainData.ServerState.DHTNodes != 600 {
		t.Errorf("DHTNodes: expected 600, got %d", mainData.ServerState.DHTNodes)
	}
}

func TestState_DeltaPreservesUnchangedFields(t *testing.T) {
	t.Parallel()

	state := NewState()

	// Full update with many fields
	state.Apply(&API.DeltaMainData{
		Rid:        1,
		FullUpdate: true,
		Torrents: map[string]json.RawMessage{
			"hash1": raw(map[string]any{
				"name": "Original", "state": stateDownload,
				"dlspeed": 100000, "size": 50000000, "progress": 0.5,
			}),
		},
		ServerState: raw(map[string]any{
			"dht_nodes": 100, "dl_info_speed": 50000,
			"global_ratio": "1.5", "connection_status": "connected",
		}),
	})

	// Delta only updates dlspeed — everything else must be preserved
	state.Apply(&API.DeltaMainData{
		Rid: 2,
		Torrents: map[string]json.RawMessage{
			"hash1": raw(map[string]any{"dlspeed": 0}),
		},
		ServerState: raw(map[string]any{"dht_nodes": 150}),
	})

	torrents := state.GetTorrents()
	if len(torrents) != 1 {
		t.Fatalf("Expected 1 torrent, got %d", len(torrents))
	}

	torrent := torrents[0]
	if torrent.Name != "Original" {
		t.Errorf("Name should be preserved: expected %q, got %q", "Original", torrent.Name)
	}

	if torrent.Size != 50000000 {
		t.Errorf("Size should be preserved: expected 50000000, got %d", torrent.Size)
	}

	if torrent.Progress != 0.5 {
		t.Errorf("Progress should be preserved: expected 0.5, got %f", torrent.Progress)
	}

	if torrent.Dlspeed != 0 {
		t.Errorf("Dlspeed should be updated: expected 0, got %d", torrent.Dlspeed)
	}

	mainData := state.GetMainData()
	if mainData.ServerState.DHTNodes != 150 {
		t.Errorf("DHTNodes should be updated: expected 150, got %d", mainData.ServerState.DHTNodes)
	}

	if mainData.ServerState.GlobalRatio != "1.5" {
		t.Errorf("GlobalRatio should be preserved: expected %q, got %q", "1.5", mainData.ServerState.GlobalRatio)
	}

	if mainData.ServerState.ConnectionStatus != "connected" {
		t.Errorf("ConnectionStatus should be preserved: expected %q, got %q", "connected", mainData.ServerState.ConnectionStatus)
	}
}

func TestState_TorrentRemoval(t *testing.T) {
	t.Parallel()

	state := NewState()

	// Initial state with 3 torrents
	initialDelta := &API.DeltaMainData{
		Rid:        100,
		FullUpdate: true,
		Torrents: map[string]json.RawMessage{
			"hash1": raw(map[string]any{"name": torrent1Name, "state": stateSeeding}),
			"hash2": raw(map[string]any{"name": torrent2Name, "state": stateSeeding}),
			"hash3": raw(map[string]any{"name": "Torrent 3", "state": stateSeeding}),
		},
	}
	state.Apply(initialDelta)

	if state.TorrentCount() != 3 {
		t.Fatalf("Initial count: expected 3, got %d", state.TorrentCount())
	}

	// Remove hash2
	deltaDelta := &API.DeltaMainData{
		Rid:             101,
		Torrents:        map[string]json.RawMessage{},
		TorrentsRemoved: []string{"hash2"},
	}
	state.Apply(deltaDelta)

	if state.TorrentCount() != 2 {
		t.Errorf("After removal: expected 2, got %d", state.TorrentCount())
	}

	torrents := state.GetTorrents()
	for _, torrent := range torrents {
		if torrent.Hash == "hash2" {
			t.Error("hash2 should have been removed")
		}
	}
}

func TestState_CategoryUpdates(t *testing.T) {
	t.Parallel()

	state := NewState()

	// Initial categories
	initialDelta := &API.DeltaMainData{
		Rid:        100,
		FullUpdate: true,
		Torrents:   map[string]json.RawMessage{},
		Categories: map[string]API.Category{
			"movies": {Name: "movies", SavePath: "/movies"},
			"music":  {Name: "music", SavePath: "/music"},
		},
	}
	state.Apply(initialDelta)

	mainData := state.GetMainData()
	if len(mainData.CategoryMap) != 2 {
		t.Fatalf("Initial categories: expected 2, got %d", len(mainData.CategoryMap))
	}

	// Add new category, remove old one
	deltaDelta := &API.DeltaMainData{
		Rid:      101,
		Torrents: map[string]json.RawMessage{},
		Categories: map[string]API.Category{
			"games": {Name: "games", SavePath: "/games"},
		},
		CategoriesRemoved: []string{"music"},
	}
	state.Apply(deltaDelta)

	mainData = state.GetMainData()
	if len(mainData.CategoryMap) != 2 {
		t.Errorf("After update: expected 2 categories, got %d", len(mainData.CategoryMap))
	}

	if _, ok := mainData.CategoryMap["movies"]; !ok {
		t.Error("movies category should still exist")
	}

	if _, ok := mainData.CategoryMap["games"]; !ok {
		t.Error("games category should have been added")
	}

	if _, ok := mainData.CategoryMap["music"]; ok {
		t.Error("music category should have been removed")
	}
}

func TestState_TagUpdates(t *testing.T) {
	t.Parallel()

	state := NewState()

	// Initial tags
	initialDelta := &API.DeltaMainData{
		Rid:        100,
		FullUpdate: true,
		Torrents:   map[string]json.RawMessage{},
		Tags:       []string{"tag1", "tag2"},
	}
	state.Apply(initialDelta)

	mainData := state.GetMainData()
	if len(mainData.Tags) != 2 {
		t.Fatalf("Initial tags: expected 2, got %d", len(mainData.Tags))
	}

	// Add new tag, remove old one
	deltaDelta := &API.DeltaMainData{
		Rid:         101,
		Torrents:    map[string]json.RawMessage{},
		Tags:        []string{"tag3"},
		TagsRemoved: []string{"tag1"},
	}
	state.Apply(deltaDelta)

	mainData = state.GetMainData()

	tagSet := make(map[string]bool)
	for _, tag := range mainData.Tags {
		tagSet[tag] = true
	}

	if !tagSet["tag2"] {
		t.Error("tag2 should still exist")
	}

	if !tagSet["tag3"] {
		t.Error("tag3 should have been added")
	}

	if tagSet["tag1"] {
		t.Error("tag1 should have been removed")
	}
}

func TestState_Reset(t *testing.T) {
	t.Parallel()

	state := NewState()

	delta := &API.DeltaMainData{
		Rid:        100,
		FullUpdate: true,
		Torrents: map[string]json.RawMessage{
			"hash1": raw(map[string]any{"name": "Test", "state": stateSeeding}),
		},
		Categories: map[string]API.Category{
			"movies": {Name: "movies"},
		},
		Tags: []string{"tag1"},
	}
	state.Apply(delta)

	if state.TorrentCount() != 1 {
		t.Fatalf("Before reset: expected 1 torrent, got %d", state.TorrentCount())
	}

	state.Reset()

	if state.GetRID() != 0 {
		t.Errorf("After reset RID: expected 0, got %d", state.GetRID())
	}

	if state.TorrentCount() != 0 {
		t.Errorf("After reset torrent count: expected 0, got %d", state.TorrentCount())
	}

	mainData := state.GetMainData()
	if len(mainData.CategoryMap) != 0 {
		t.Errorf("After reset categories: expected 0, got %d", len(mainData.CategoryMap))
	}

	if len(mainData.Tags) != 0 {
		t.Errorf("After reset tags: expected 0, got %d", len(mainData.Tags))
	}
}

func TestState_FirstDeltaIsTreatedAsFullUpdate(t *testing.T) {
	t.Parallel()

	state := NewState()

	// Even with FullUpdate=false, first update (rid=0) should be treated as full
	delta := &API.DeltaMainData{
		Rid:        50,
		FullUpdate: false, // This would normally mean delta, but rid=0 forces full
		Torrents: map[string]json.RawMessage{
			"hash1": raw(map[string]any{"name": "Test", "state": stateSeeding}),
		},
	}
	state.Apply(delta)

	if state.TorrentCount() != 1 {
		t.Errorf("Expected 1 torrent, got %d", state.TorrentCount())
	}

	if state.GetRID() != 50 {
		t.Errorf("RID: expected 50, got %d", state.GetRID())
	}
}

func TestState_HashSetFromMapKey(t *testing.T) {
	t.Parallel()

	state := NewState()

	// Hash is not in the torrent JSON, should come from map key
	delta := &API.DeltaMainData{
		Rid:        100,
		FullUpdate: true,
		Torrents: map[string]json.RawMessage{
			"expected-hash": raw(map[string]any{"name": "Test Torrent"}),
		},
	}
	state.Apply(delta)

	torrents := state.GetTorrents()
	if len(torrents) != 1 {
		t.Fatalf("Expected 1 torrent, got %d", len(torrents))
	}

	if torrents[0].Hash != "expected-hash" {
		t.Errorf("Hash: expected %q, got %q", "expected-hash", torrents[0].Hash)
	}
}
