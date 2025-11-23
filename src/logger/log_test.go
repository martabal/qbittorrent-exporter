package logger

import (
	"log/slog"
	"testing"
)

func TestSetLogLevel(t *testing.T) {
	tests := [...]struct {
		name           string
		inputLogLevel  string
		expectedOutput string
		expectedLevel  slog.Level
	}{
		{"ValidLogLevelTrace", "TRACE", "TRACE", slog.Level(LevelTrace)},
		{"ValidLogLevelTraceLower", "Trace", "TRACE", slog.Level(LevelTrace)},
		{"ValidLogLevelFullLower", "trace", "TRACE", slog.Level(LevelTrace)},
		{"ValidLogLevelDebug", "DEBUG", "DEBUG", slog.Level(LevelDebug)},
		{"ValidLogLevelInfo", "INFO", "INFO", slog.Level(LevelInfo)},
		{"ValidLogLevelWarn", "WARN", "WARN", slog.Level(LevelWarn)},
		{"ValidLogLevelError", "ERROR", "ERROR", slog.Level(LevelError)},
		{"InvalidLogLevel", "INVALID", "INFO", slog.Level(LevelInfo)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logLevel := SetLogLevel(tt.inputLogLevel)
			if logLevel != tt.expectedOutput {
				t.Errorf("expected %v, got %v", tt.expectedOutput, logLevel)
			}

		})
	}
}
