package logger

import (
	"log/slog"
	"testing"
)

func TestSetLogLevel(t *testing.T) {
	tests := []struct {
		name           string
		inputLogLevel  string
		expectedOutput string
		expectedLevel  slog.Level
	}{
		{"ValidLogLevelTrace", "TRACE", "TRACE", slog.Level(Trace)},
		{"ValidLogLevelDebug", "DEBUG", "DEBUG", slog.Level(Debug)},
		{"ValidLogLevelInfo", "INFO", "INFO", slog.Level(Info)},
		{"ValidLogLevelWarn", "WARN", "WARN", slog.Level(Warn)},
		{"ValidLogLevelError", "ERROR", "ERROR", slog.Level(Error)},
		{"InvalidLogLevel", "INVALID", "INFO", slog.Level(Info)},
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
