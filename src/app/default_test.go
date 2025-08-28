package app

import (
	"bytes"
	"fmt"
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
	if err := os.Setenv(envVar, expectedValue); err != nil {
		panic(fmt.Sprintf("Error setting %s: %s", envVar, expectedValue))
	}
	defer func() {
		if err := os.Unsetenv(envVar); err != nil {
			t.Fatalf("Error unsetting %s: %v", envVar, err)
		}
	}()
	value, _ := getEnv(defaultPort)

	if value != expectedValue {
		t.Errorf("Expected %s, got %s", expectedValue, value)
	}
}

func TestGetEnvReturnsDefaultWhenEnvNotSet(t *testing.T) {
	envVar := "EXPORTER_PORT"
	expectedValue := strconv.Itoa(defaultExporterPort)
	defer func() {
		if err := os.Unsetenv(envVar); err != nil {
			t.Fatalf("Error unsetting %s: %v", envVar, err)
		}
	}()
	value, _ := getEnv(defaultPort)

	if value != expectedValue {
		t.Errorf("Expected default %s, got %s", expectedValue, value)
	}
}

func TestGetEnvLogsWarningIfHelpMessagePresent(t *testing.T) {
	envVar := "QBITTORRENT_USERNAME"
	if err := os.Unsetenv(envVar); err != nil {
		t.Fatalf("Error unsetting %s: %v", envVar, err)
	}
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
			if err := os.Unsetenv(tt.env.Key); err != nil {
				t.Fatalf("Error unsetting %s: %v", tt.env.Key, err)
			}
			value, _ := getEnv(tt.env)
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
		{defaultEnableTracker, "false", "false"},
		{defaultLabelWithTracker, "false", "false"},
		{defaultHighCardinality, "true", "true"},
		{defaultLabelWithHash, "true", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.envVar.Key, func(t *testing.T) {
			cleanup := setAndClearEnv(tt.envVar.Key, tt.setValue, t)
			defer cleanup()

			value, _ := getEnv(tt.envVar)
			if value != tt.expectVal {
				t.Errorf("Expected %s, got %s", tt.expectVal, value)
			}
		})
	}
}

func TestGetEnvHandlesEmptyEnvVarGracefully(t *testing.T) {
	envVar := "QBITTORRENT_USERNAME"
	if err := os.Setenv(envVar, ""); err != nil {
		panic(fmt.Sprintf("Error setting %s: %s", envVar, ""))
	}
	defer func() {
		if err := os.Unsetenv(envVar); err != nil {
			t.Fatalf("Error unsetting %s: %v", envVar, err)
		}
	}()
	expectedValue := defaultUsername.DefaultValue

	value, _ := getEnv(defaultUsername)

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

			if err := os.Unsetenv(tt.env.Key); err != nil {
				t.Fatalf("Error unsetting %s: %v", tt.env.Key, err)
			}

			getEnv(tt.env)

			if !strings.Contains(buff.String(), tt.expectedLog) {
				t.Errorf("Expected log message to contain '%s', got '%s'", tt.expectedLog, buff.String())
			}
		})
	}
}

func setAndClearEnv(key, value string, t *testing.T) func() {
	if err := os.Setenv(key, value); err != nil {
		panic(fmt.Sprintf("Error setting %s: %s", key, value))
	}
	return func() {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Error unsetting %s: %v", key, err)
		}
	}
}
