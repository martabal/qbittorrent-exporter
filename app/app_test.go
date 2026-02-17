package app

import (
	"os"
	"testing"
)

func TestGetFeaturesEnabled(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
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
				EnableLabelWithHash:    true,
				EnableLabelWithTracker: true,
			},
			expectedOutput: "[High cardinality, Trackers, Label with tracker (experimental), Label with hash (experimental)]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			// Set main features
			Exporter.Features.EnableHighCardinality = test.features.EnableHighCardinality
			Exporter.Features.EnableTracker = test.features.EnableTracker

			// Set experimental features
			Exporter.ExperimentalFeatures.EnableLabelWithHash = test.experimentalFeature.EnableLabelWithHash
			Exporter.ExperimentalFeatures.EnableLabelWithTracker = test.experimentalFeature.EnableLabelWithTracker

			result := getFeaturesEnabled()
			if result != test.expectedOutput {
				t.Errorf("expected %s, got %s", test.expectedOutput, result)
			}
		})
	}
}

func TestEnvSetToTrue(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
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

func setPassFile(t *testing.T, pass string) func() {
	t.Helper()

	tmpPath := t.TempDir() + string(os.PathSeparator) + "qbit_pass_test.txt"

	err := os.WriteFile(tmpPath, []byte(pass), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	return setAndClearEnv(t, defaultPasswordFile, tmpPath)
}

// file: Y
// env: N
func TestPassFileSet(t *testing.T) { //nolint:paralleltest
	expected := "welcome123"

	cleanEnvFile := setPassFile(t, expected)
	defer cleanEnvFile()

	got, usingDefaultValue := getPassword()

	if got != expected || usingDefaultValue {
		t.Errorf("GetPassword() = %q; want %q", got, expected)
	}
}

// file: Y
// env: Y
func TestPassFileAndPassSet(t *testing.T) { //nolint:paralleltest
	expected := "welcome123"

	cleanEnvFile := setPassFile(t, expected)
	defer cleanEnvFile()

	cleanEnv := setAndClearEnv(t, "QBITTORRENT_PASSWORD", "anotherpass")
	defer cleanEnv()

	got, usingDefaultValue := getPassword()

	if got != expected || usingDefaultValue {
		t.Errorf("GetPassword() = %q; want %q", got, expected)
	}
}

func TestGetBasicAuth(t *testing.T) {
	tests := []struct {
		name         string
		username     *string
		password     *string
		wantNil      bool
		wantUsername string
		wantPassword string
	}{
		{
			name:     "Both nil returns nil",
			username: nil,
			password: nil,
			wantNil:  true,
		},
		{
			name:         "Both set returns BasicAuth",
			username:     new("user"),
			password:     new("pass"),
			wantNil:      false,
			wantUsername: "user",
			wantPassword: "pass",
		},
		{
			name:         "Only username set",
			username:     new("user"),
			password:     nil,
			wantNil:      false,
			wantUsername: "user",
			wantPassword: "",
		},
		{
			name:         "Only password set",
			username:     nil,
			password:     new("pass"),
			wantNil:      false,
			wantUsername: "",
			wantPassword: "pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBasicAuth(tt.username, tt.password, "DEFAULT_USER", "DEFAULT_PASS")

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
			} else {
				if result == nil {
					t.Fatal("expected non-nil BasicAuth")
				}

				if result.Username != tt.wantUsername {
					t.Errorf("username: expected %q, got %q", tt.wantUsername, result.Username)
				}

				if result.Password != tt.wantPassword {
					t.Errorf("password: expected %q, got %q", tt.wantPassword, result.Password)
				}
			}
		})
	}
}

func TestGetPasswordMasked(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		want     string
	}{
		{"Empty password", "", ""},
		{"Short password", "abc", "***"},
		{"Normal password", "password123", "***********"},
		{"Long password", "verylongpasswordhere", "********************"},
		{"Special chars", "p@$$w0rd!", "*********"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GetPasswordMasked(tt.password)
			if got != tt.want {
				t.Errorf("GetPasswordMasked(%q) = %q, want %q", tt.password, got, tt.want)
			}

			// Verify all characters are asterisks
			for _, char := range got {
				if char != '*' {
					t.Errorf("GetPasswordMasked result contains non-asterisk character: %c", char)
				}
			}
		})
	}
}
