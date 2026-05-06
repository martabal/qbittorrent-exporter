package deltasync

import (
	"encoding/json"
	"fmt"
	"testing"

	API "qbit-exp/api"
)

func BenchmarkApplyFullUpdate(b *testing.B) {
	for _, count := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("torrents_%d", count), func(b *testing.B) {
			delta := buildFullDelta(count)

			b.ResetTimer()

			for b.Loop() {
				state := NewState()
				state.Apply(delta)
			}
		})
	}
}

func BenchmarkApplyDeltaUpdate(b *testing.B) {
	for _, changed := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("changed_%d_of_1000", changed), func(b *testing.B) {
			state := NewState()
			state.Apply(buildFullDelta(1000))

			delta := buildPartialDelta(changed, 1000)

			b.ResetTimer()

			for b.Loop() {
				state.Apply(delta)
			}
		})
	}
}

func buildFullDelta(count int) *API.DeltaMainData {
	torrents := make(map[string]json.RawMessage, count)

	for i := range count {
		hash := fmt.Sprintf("hash_%d", i)
		data := map[string]any{
			"name":     fmt.Sprintf("Torrent %d", i),
			"state":    "downloading",
			"progress": 0.5,
			"dlspeed":  1000000,
			"upspeed":  500000,
			"size":     50000000,
		}

		raw, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}

		torrents[hash] = raw
	}

	serverState, err := json.Marshal(map[string]any{
		"dht_nodes": 500, "dl_info_speed": 1000000,
		"global_ratio": "1.5", "connection_status": "connected",
	})
	if err != nil {
		panic(err)
	}

	return &API.DeltaMainData{ //nolint:exhaustruct
		Rid:         1,
		FullUpdate:  true,
		Torrents:    torrents,
		ServerState: serverState,
	}
}

func buildPartialDelta(changed, ridOffset int) *API.DeltaMainData {
	torrents := make(map[string]json.RawMessage, changed)

	for i := range changed {
		hash := fmt.Sprintf("hash_%d", i)
		data := map[string]any{"dlspeed": 999, "state": "seeding"}

		raw, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}

		torrents[hash] = raw
	}

	return &API.DeltaMainData{ //nolint:exhaustruct
		Rid:         int64(ridOffset + 1),
		Torrents:    torrents,
		ServerState: json.RawMessage(`{"dht_nodes": 600}`),
	}
}
