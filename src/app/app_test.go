package app

import (
	"strings"
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
		{
			name: "Only Tracker Label enabled",
			features: Features{
				EnableTrackerLabel: true,
			},
			experimentalFeature: ExperimentalFeatures{},
			expectedOutput:      "[Tracker Label]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Set main features
			Exporter.Features.EnableHighCardinality = test.features.EnableHighCardinality
			Exporter.Features.EnableTracker = test.features.EnableTracker
			Exporter.Features.EnableTrackerLabel = test.features.EnableTrackerLabel

			// Set experimental features
			Exporter.ExperimentalFeatures.EnableLabelWithHash = test.experimentalFeature.EnableLabelWithHash

			result := getFeaturesEnabled()
			if result != test.expectedOutput {
				t.Errorf("expected %s, got %s", test.expectedOutput, result)
			}
		})
	}
}

func TestEnvSetToTrue(t *testing.T) {
	tests := []struct {
		input  string
		output bool
	}{
		{"true", true},
		{"TRUE", true},
		{"False", false},
		{"false", false},
		{"1", false},
		{"0", false},
		{"", false},
		{"randomstring", false},
	}

	for _, test := range tests {
		got := envSetToTrue(test.input)
		if got != test.output {
			t.Errorf("envSetToTrue(%q) = %v; want %v", test.input, got, test.output)
		}
	}
}

func TestGetPasswordMasked(t *testing.T) {
	QBittorrent.Password = "mysecretpassword"
	expected := strings.Repeat("*", len(QBittorrent.Password))
	got := GetPasswordMasked()

	if got != expected {
		t.Errorf("GetPasswordMasked() = %q; want %q", got, expected)
	}
}
