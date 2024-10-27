package app

import (
	"testing"
)

func TestGetFeaturesEnabled(t *testing.T) {
	tests := []struct {
		name                string
		features            Features
		experimentalFeature ExperimentalFeatures

		expectedOutput string
	}{
		{
			name: "All features disabled",
			features: Features{
				EnableHighCardinality: false,
				EnableTracker:         false,
			},
			experimentalFeature: ExperimentalFeatures{
				EnableLabelWithHash: false,
			},
			expectedOutput: "[]",
		},
		{
			name: "Only High Cardinality enabled",
			features: Features{
				EnableHighCardinality: true,
				EnableTracker:         false,
			},
			experimentalFeature: ExperimentalFeatures{
				EnableLabelWithHash: false,
			},
			expectedOutput: "[High cardinality]",
		},
		{
			name: "Only Trackers enabled",
			features: Features{
				EnableHighCardinality: false,
				EnableTracker:         true,
			},
			experimentalFeature: ExperimentalFeatures{
				EnableLabelWithHash: false,
			},
			expectedOutput: "[Trackers]",
		},
		{
			name: "Both High Cardinality and Trackers enabled",
			features: Features{
				EnableHighCardinality: true,
				EnableTracker:         true,
			},
			experimentalFeature: ExperimentalFeatures{
				EnableLabelWithHash: false,
			},
			expectedOutput: "[High cardinality, Trackers]",
		},
		{
			name: "Experimental feature enabled",
			features: Features{
				EnableHighCardinality: false,
				EnableTracker:         false,
			},
			experimentalFeature: ExperimentalFeatures{
				EnableLabelWithHash: true,
			},
			expectedOutput: "[Label with hash (experimental)]",
		},
		{
			name: "All features enabled",
			features: Features{
				EnableHighCardinality: true,
				EnableTracker:         true,
			},
			experimentalFeature: ExperimentalFeatures{
				EnableLabelWithHash: true,
			},
			expectedOutput: "[High cardinality, Trackers, Label with hash (experimental)]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set main features
			Exporter.Feature.EnableHighCardinality = test.features.EnableHighCardinality
			Exporter.Feature.EnableTracker = test.features.EnableTracker

			// Set experimental features
			Exporter.ExperimentalFeature.EnableLabelWithHash = test.experimentalFeature.EnableLabelWithHash

			result := GetFeaturesEnabled()
			if result != test.expectedOutput {
				t.Errorf("expected %s, got %s", test.expectedOutput, result)
			}
		})
	}
}
