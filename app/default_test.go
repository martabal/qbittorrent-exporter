package app

import (
	"bytes"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"

	"qbit-exp/logger"
)

var buff = &bytes.Buffer{}

func init() {
	logger.Log = &logger.Logger{Logger: slog.New(slog.NewTextHandler(buff, &slog.HandlerOptions{}))}
}

func TestGetEnvReturnsEnvValue(t *testing.T) {
	t.Setenv(defaultPort.Key, strconv.Itoa(defaultExporterPort))

	defer func() {
		err := os.Unsetenv(defaultPort.Key)
		if err != nil {
			t.Fatalf("Error unsetting %s: %v", defaultPort.Key, err)
		}
	}()

	value, _ := getEnv(defaultPort)

	if value != strconv.Itoa(defaultExporterPort) {
		t.Errorf("Expected %d, got %s", defaultExporterPort, value)
	}
}

func TestGetEnvReturnsDefaultWhenEnvNotSet(t *testing.T) {
	t.Parallel()

	expectedValue := strconv.Itoa(defaultExporterPort)

	defer func() {
		err := os.Unsetenv(defaultPort.Key)
		if err != nil {
			t.Fatalf("Error unsetting %s: %v", defaultPort.Key, err)
		}
	}()

	value, _ := getEnv(defaultPort)

	if value != expectedValue {
		t.Errorf("Expected default %s, got %s", expectedValue, value)
	}
}

func TestGetEnvLogsWarningIfHelpMessagePresent(t *testing.T) {
	t.Parallel()

	err := os.Unsetenv(defaultUsername.Key)
	if err != nil {
		t.Fatalf("Error unsetting %s: %v", defaultUsername.Key, err)
	}

	expectedLogMessage := defaultUsername.Help
	getEnv(defaultUsername)

	if !strings.Contains(buff.String(), expectedLogMessage) {
		t.Errorf("Expected log message to contain '%s', got '%s'", expectedLogMessage, buff.String())
	}
}

func TestGetEnvWithDifferentDefaults(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name          string
		env           Env
		expectedValue string
	}{
		{"DefaultLogLevel", defaultLogLevel, "INFO"},
		{"DefaultPort", defaultPort, strconv.Itoa(defaultExporterPort)},
		{"DefaultTimeout", defaultTimeout, strconv.Itoa(DefaultTimeout)},
		{"DefaultUsername", defaultUsername, "admin"},
		{"DefaultPassword", defaultPassword, "adminadmin"},
		{"DefaultBaseUrl", defaultBaseUrl, "http://localhost:8080"},
		{"DefaultEnableTracker", defaultEnableTracker, "true"},
		{"DefaultTrackerLabel", defaultLabelWithTracker, "false"},
		{"DefaultHighCardinality", defaultHighCardinality, "false"},
		{"DefaultLabelWithHash", defaultLabelWithHash, "false"},
		{"DefaultExporterPath", defaultExporterPathEnv, defaultExporterPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := os.Unsetenv(tt.env.Key)
			if err != nil {
				t.Fatalf("Error unsetting %s: %v", tt.env.Key, err)
			}

			value, _ := getEnv(tt.env)
			if value != tt.expectedValue {
				t.Errorf("Expected %s, got %s", tt.expectedValue, value)
			}
		})
	}
}

func TestGetEnvReturnsBooleanValues(t *testing.T) { //nolint:paralleltest
	tests := [...]struct {
		envVar    Env
		setValue  string
		expectVal string
	}{
		{defaultEnableTracker, "false", "false"},
		{defaultLabelWithTracker, "false", "false"},
		{defaultHighCardinality, "true", "true"},
		{defaultLabelWithHash, "true", "true"},
	}

	for _, tt := range tests { //nolint:paralleltest
		t.Run(tt.envVar.Key, func(t *testing.T) {
			cleanup := setAndClearEnv(t, tt.envVar.Key, tt.setValue)
			defer cleanup()

			value, _ := getEnv(tt.envVar)
			if value != tt.expectVal {
				t.Errorf("Expected %s, got %s", tt.expectVal, value)
			}
		})
	}
}

func TestGetEnvHandlesEmptyEnvVarGracefully(t *testing.T) {
	t.Setenv(defaultUsername.Key, "")

	defer func() {
		err := os.Unsetenv(defaultUsername.Key)
		if err != nil {
			t.Fatalf("Error unsetting %s: %v", defaultUsername.Key, err)
		}
	}()

	expectedValue := defaultUsername.DefaultValue

	value, _ := getEnv(defaultUsername)

	if value != expectedValue {
		t.Errorf("Expected %s, got %s", expectedValue, value)
	}
}

func TestGetEnvLogsWarningsCorrectly(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name        string
		env         Env
		expectedLog string
	}{
		{"UsernameWarning", defaultUsername, "qBittorrent username is not set. Using default username"},
		{"PasswordWarning", defaultPassword, "qBittorrent password is not set. Using default password"},
		{"BaseUrlWarning", defaultBaseUrl, "qBittorrent base_url is not set. Using default base_url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := os.Unsetenv(tt.env.Key)
			if err != nil {
				t.Fatalf("Error unsetting %s: %v", tt.env.Key, err)
			}

			getEnv(tt.env)

			if !strings.Contains(buff.String(), tt.expectedLog) {
				t.Errorf("Expected log message to contain '%s', got '%s'", tt.expectedLog, buff.String())
			}
		})
	}
}

func setAndClearEnv(t *testing.T, key, value string) func() {
	t.Helper()

	t.Setenv(key, value)

	return func() {
		err := os.Unsetenv(key)
		if err != nil {
			t.Fatalf("Error unsetting %s: %v", key, err)
		}
	}
}
