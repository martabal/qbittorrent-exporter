package deltasync

import (
	"encoding/json"
	"maps"
	"sync"

	API "qbit-exp/api"
)

// State holds the synchronized torrent state between scrapes.
// It maintains a complete view of all torrents by applying deltas.
type State struct {
	mu          sync.RWMutex
	rid         int64
	torrents    map[string]API.Info
	categories  map[string]API.Category
	tags        []string
	serverState API.ServerState
}

// NewState creates a new empty sync state.
func NewState() *State {
	return &State{
		torrents:   make(map[string]API.Info),
		categories: make(map[string]API.Category),
		tags:       []string{},
	}
}

// GetRID returns the current response ID for delta requests.
func (s *State) GetRID() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.rid
}

// GetTorrents returns a slice of all torrents.
func (s *State) GetTorrents() API.SliceInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(API.SliceInfo, 0, len(s.torrents))
	for _, info := range s.torrents {
		result = append(result, info)
	}

	return result
}

// GetMainData returns the current MainData (categories, tags, server state).
func (s *State) GetMainData() API.MainData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	categories := make(map[string]API.Category, len(s.categories))
	maps.Copy(categories, s.categories)

	tags := make([]string, len(s.tags))
	copy(tags, s.tags)

	return API.MainData{
		CategoryMap: categories,
		ServerState: s.serverState,
		Tags:        tags,
	}
}

// TorrentCount returns the number of torrents in state.
func (s *State) TorrentCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.torrents)
}

// Apply updates the state with delta data from sync/maindata response.
// If fullUpdate is true or this is the first update (rid=0), state is replaced.
// Otherwise, changes are merged into existing state.
func (s *State) Apply(delta *API.DeltaMainData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Full update: replace all state
	if delta.FullUpdate || s.rid == 0 {
		s.applyFullUpdate(delta)

		return
	}

	// Delta update: merge changes
	s.applyDeltaUpdate(delta)
}

// Reset clears all state and resets rid to 0, forcing a full sync on next request.
func (s *State) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rid = 0
	s.torrents = make(map[string]API.Info)
	s.categories = make(map[string]API.Category)
	s.tags = []string{}
	s.serverState = API.ServerState{}
}

func (s *State) applyFullUpdate(delta *API.DeltaMainData) {
	// Clear and rebuild torrents
	s.torrents = make(map[string]API.Info, len(delta.Torrents))
	for hash, raw := range delta.Torrents {
		var info API.Info
		err := json.Unmarshal(raw, &info)
		if err != nil {
			continue
		}

		info.Hash = hash
		s.torrents[hash] = info
	}

	// Clear and rebuild categories
	s.categories = make(map[string]API.Category, len(delta.Categories))
	maps.Copy(s.categories, delta.Categories)

	// Replace tags
	s.tags = make([]string, len(delta.Tags))
	copy(s.tags, delta.Tags)

	// Replace server state (full update includes all fields)
	s.serverState = API.ServerState{}
	if len(delta.ServerState) > 0 {
		_ = json.Unmarshal(delta.ServerState, &s.serverState)
	}

	// Update rid
	s.rid = delta.Rid
}

func (s *State) applyDeltaUpdate(delta *API.DeltaMainData) {
	// Apply torrent updates — json.Unmarshal into an existing struct
	// only overwrites fields present in the JSON, providing merge semantics.
	for hash, raw := range delta.Torrents {
		existing := s.torrents[hash] // zero value if new torrent

		err := json.Unmarshal(raw, &existing)
		if err != nil {
			continue
		}

		existing.Hash = hash
		s.torrents[hash] = existing
	}

	// Remove deleted torrents
	for _, hash := range delta.TorrentsRemoved {
		delete(s.torrents, hash)
	}

	// Apply category updates
	maps.Copy(s.categories, delta.Categories)

	// Remove deleted categories
	for _, name := range delta.CategoriesRemoved {
		delete(s.categories, name)
	}

	// Apply tag updates (tags in delta are additions)
	if len(delta.Tags) > 0 {
		tagSet := make(map[string]struct{}, len(s.tags)+len(delta.Tags))
		for _, tag := range s.tags {
			tagSet[tag] = struct{}{}
		}

		for _, tag := range delta.Tags {
			tagSet[tag] = struct{}{}
		}

		s.tags = make([]string, 0, len(tagSet))
		for tag := range tagSet {
			s.tags = append(s.tags, tag)
		}
	}

	// Remove deleted tags
	if len(delta.TagsRemoved) > 0 {
		removeSet := make(map[string]struct{}, len(delta.TagsRemoved))
		for _, tag := range delta.TagsRemoved {
			removeSet[tag] = struct{}{}
		}

		filtered := make([]string, 0, len(s.tags))
		for _, tag := range s.tags {
			if _, remove := removeSet[tag]; !remove {
				filtered = append(filtered, tag)
			}
		}

		s.tags = filtered
	}

	// Merge server state (only update fields present in delta)
	if len(delta.ServerState) > 0 {
		_ = json.Unmarshal(delta.ServerState, &s.serverState)
	}

	// Update rid
	s.rid = delta.Rid
}
