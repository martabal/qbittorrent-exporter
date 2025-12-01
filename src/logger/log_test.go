package logger

import (
	"log/slog"
	"testing"
)

func TestSetLogLevel(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name           string
		inputLogLevel  string
		expectedOutput string
		expectedLevel  slog.Level
	}{
		{"ValidLogLevelTrace", "TRACE", "TRACE", LevelTrace},
		{"ValidLogLevelTraceLower", "Trace", "TRACE", LevelTrace},
		{"ValidLogLevelFullLower", "trace", "TRACE", LevelTrace},
		{"ValidLogLevelDebug", "DEBUG", "DEBUG", LevelDebug},
		{"ValidLogLevelInfo", "INFO", "INFO", LevelInfo},
		{"ValidLogLevelWarn", "WARN", "WARN", LevelWarn},
		{"ValidLogLevelError", "ERROR", "ERROR", LevelError},
		{"InvalidLogLevel", "INVALID", "INFO", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logLevel := SetLogLevel(tt.inputLogLevel)
			if logLevel != tt.expectedOutput {
				t.Errorf("expected %v, got %v", tt.expectedOutput, logLevel)
			}
		})
	}
}
