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
	envVar := "EXPORTER_PORT"
	expectedValue := "9090"
	os.Setenv(envVar, expectedValue)
	defer os.Unsetenv(envVar)
	value := getEnv(defaultPort)

	if value != expectedValue {
		t.Errorf("Expected %s, got %s", expectedValue, value)
	}
}

func TestGetEnvReturnsDefaultWhenEnvNotSet(t *testing.T) {
	envVar := "EXPORTER_PORT"
	expectedValue := strconv.Itoa(DefaultExporterPort)
	os.Unsetenv(envVar)
	value := getEnv(defaultPort)

	if value != expectedValue {
		t.Errorf("Expected default %s, got %s", expectedValue, value)
	}
}

func TestGetEnvLogsWarningIfHelpMessagePresent(t *testing.T) {
	envVar := "QBITTORRENT_USERNAME"
	os.Unsetenv(envVar)
	expectedLogMessage := defaultUsername.Help
	getEnv(defaultUsername)

	if !strings.Contains(buff.String(), expectedLogMessage) {
		t.Errorf("Expected log message to contain '%s', got '%s'", expectedLogMessage, buff.String())
	}
}

func TestGetEnvWithDifferentDefaults(t *testing.T) {
	tests := []struct {
		name          string
		env           Env
		expectedValue string
	}{
		{"DefaultLogLevel", defaultLogLevel, "INFO"},
		{"DefaultPort", defaultPort, strconv.Itoa(DefaultExporterPort)},
		{"DefaultTimeout", defaultTimeout, strconv.Itoa(DefaultTimeout)},
		{"DefaultUsername", defaultUsername, "admin"},
		{"DefaultPassword", defaultPassword, "adminadmin"},
		{"DefaultBaseUrl", defaultBaseUrl, "http://localhost:8080"},
		{"DefaultDisableTracker", defaultDisableTracker, "true"},
		{"DefaultHighCardinality", defaultHighCardinality, "false"},
		{"DefaultLabelWithHash", defaultLabelWithHash, "false"},
		{"DefaultExporterPath", defaultExporterPath, DefaultExporterPath},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv(tt.env.Key)
			value := getEnv(tt.env)
			if value != tt.expectedValue {
				t.Errorf("Expected %s, got %s", tt.expectedValue, value)
			}
		})
	}
}

func TestGetEnvReturnsBooleanValues(t *testing.T) {
	tests := []struct {
		envVar    Env
		setValue  string
		expectVal string
	}{
		{defaultDisableTracker, "false", "false"},
		{defaultHighCardinality, "true", "true"},
		{defaultLabelWithHash, "true", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.envVar.Key, func(t *testing.T) {
			cleanup := setAndClearEnv(tt.envVar.Key, tt.setValue)
			defer cleanup()

			value := getEnv(tt.envVar)
			if value != tt.expectVal {
				t.Errorf("Expected %s, got %s", tt.expectVal, value)
			}
		})
	}
}

func TestGetEnvHandlesEmptyEnvVarGracefully(t *testing.T) {
	envVar := "QBITTORRENT_USERNAME"
	os.Setenv(envVar, "")
	defer os.Unsetenv(envVar)
	expectedValue := defaultUsername.DefaultValue

	value := getEnv(defaultUsername)

	if value != expectedValue {
		t.Errorf("Expected %s, got %s", expectedValue, value)
	}
}

func TestGetEnvLogsWarningsCorrectly(t *testing.T) {
	tests := []struct {
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

			os.Unsetenv(tt.env.Key)

			getEnv(tt.env)

			if !strings.Contains(buff.String(), tt.expectedLog) {
				t.Errorf("Expected log message to contain '%s', got '%s'", tt.expectedLog, buff.String())
			}
		})
	}
}

func setAndClearEnv(key, value string) func() {
	os.Setenv(key, value)
	return func() {
		os.Unsetenv(key)
	}
}
