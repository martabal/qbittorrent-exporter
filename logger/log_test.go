package logger

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
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

func TestLoggingFunctions(t *testing.T) {
	tests := []struct {
		name    string
		logFunc func(string)
		message string
		level   slog.Level
	}{
		{"Trace", Trace, "trace message", LevelTrace},
		{"Debug", Debug, "debug message", LevelDebug},
		{"Info", Info, "info message", LevelInfo},
		{"Warn", Warn, "warn message", LevelWarn},
		{"Error", Error, "error message", LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a logger that will accept the log level
			var buf bytes.Buffer

			opts := &slog.HandlerOptions{
				Level: tt.level,
			}
			Log = &Logger{Logger: slog.New(slog.NewTextHandler(&buf, opts))}

			// Call the logging function - this will write to os.Stdout via PrettyHandler
			// But we're verifying it doesn't panic and the Log object is initialized
			tt.logFunc(tt.message)

			// Verify the Log object is properly set up
			if Log == nil {
				t.Error("expected Log to be initialized")
			}
		})
	}
}

func TestPrettyHandler_Handle(t *testing.T) {
	tests := []struct {
		name      string
		level     slog.Level
		message   string
		wantColor string
	}{
		{"TraceLevel", LevelTrace, "trace test", Purple},
		{"DebugLevel", LevelDebug, "debug test", Green},
		{"InfoLevel", LevelInfo, "info test", Blue},
		{"WarnLevel", LevelWarn, "warn test", Yellow},
		{"ErrorLevel", LevelError, "error test", Red},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since PrettyHandler writes directly to os.Stdout/os.Stderr,
			// we'll just ensure it doesn't error and verify the level name exists in ReverseLogLevels
			handler := NewPrettyHandler(os.Stdout, slog.HandlerOptions{Level: tt.level})

			record := slog.NewRecord(time.Now(), tt.level, tt.message, 0)

			err := handler.Handle(context.Background(), record)
			if err != nil {
				t.Fatalf("Handle() error = %v", err)
			}

			// Verify that the level exists in our mapping
			levelName, exists := ReverseLogLevels[tt.level]
			if !exists {
				t.Errorf("expected level %v to exist in ReverseLogLevels", tt.level)
			}

			if levelName == "" {
				t.Errorf("expected non-empty level name for level %v", tt.level)
			}
		})
	}
}

func TestColorLogLevel(t *testing.T) {
	expectedColors := map[slog.Level]string{
		LevelTrace: Purple,
		LevelDebug: Green,
		LevelInfo:  Blue,
		LevelWarn:  Yellow,
		LevelError: Red,
	}

	for level, expectedColor := range expectedColors {
		if color, exists := ColorLogLevel[level]; !exists {
			t.Errorf("ColorLogLevel missing entry for level %v", level)
		} else if color != expectedColor {
			t.Errorf("ColorLogLevel[%v] = %q, want %q", level, color, expectedColor)
		}
	}
}

func TestReverseLogLevels(t *testing.T) {
	for name, level := range LogLevels {
		reversedName, exists := ReverseLogLevels[level]
		if !exists {
			t.Errorf("ReverseLogLevels missing entry for level %v", level)
		} else if reversedName != name {
			t.Errorf("ReverseLogLevels[%v] = %q, want %q", level, reversedName, name)
		}
	}
}
