package app

import (
	"testing"
)

func TestGetFeaturesEnabled(t *testing.T) {
	tests := []struct {
		name                  string
		enableHighCardinality bool
		disableTracker        bool
		expectedOutput        string
	}{
		{
			name:                  "Both disabled",
			enableHighCardinality: false,
			disableTracker:        true,
			expectedOutput:        "[]",
		},
		{
			name:                  "Only High Cardinality enabled",
			enableHighCardinality: true,
			disableTracker:        true,
			expectedOutput:        "[High cardinality]",
		},
		{
			name:                  "Only Trackers enabled",
			enableHighCardinality: false,
			disableTracker:        false,
			expectedOutput:        "[Trackers]",
		},
		{
			name:                  "Both High Cardinality and Trackers enabled",
			enableHighCardinality: true,
			disableTracker:        false,
			expectedOutput:        "[High cardinality, Trackers]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			EnableHighCardinality = test.enableHighCardinality
			DisableTracker = test.disableTracker

			result := GetFeaturesEnabled()
			if result != test.expectedOutput {
				t.Errorf("expected %s, got %s", test.expectedOutput, result)
			}
		})
	}
}
