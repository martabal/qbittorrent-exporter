package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Logger struct {
	*slog.Logger
}

type PrettyHandler struct {
	slog.Handler
}

var LogLevels = map[string]slog.Level{
	"TRACE": LevelTrace,
	"DEBUG": LevelDebug,
	"INFO":  LevelInfo,
	"WARN":  LevelWarn,
	"ERROR": LevelError,
}

var ReverseLogLevels = func(m map[string]slog.Level) map[slog.Level]string {
	rev := make(map[slog.Level]string, len(m))
	for k, v := range m {
		rev[v] = k
	}

	return rev
}(LogLevels)

var ColorLogLevel = map[slog.Level]string{
	LevelTrace: Purple,
	LevelDebug: Green,
	LevelInfo:  Blue,
	LevelWarn:  Yellow,
	LevelError: Red,
}

const (
	LevelTrace slog.Level = -8
	LevelDebug slog.Level = -4
	LevelInfo  slog.Level = 0
	LevelWarn  slog.Level = 4
	LevelError slog.Level = 8
)

const (
	Reset  string = "\033[0m"
	Red    string = "\033[31m"
	Green  string = "\033[32m"
	Yellow string = "\033[33m"
	Blue   string = "\033[34m"
	Purple string = "\033[35m"
)

func NewPrettyHandler(
	out io.Writer,
	opts slog.HandlerOptions,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewTextHandler(out, &opts),
	}

	return h
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	level := ReverseLogLevels[r.Level]
	timeStr := fmt.Sprintf("[%02d-%02d-%02d %02d:%02d:%02d]", r.Time.Year(), r.Time.Month(), r.Time.Day(), r.Time.Hour(), r.Time.Minute(), r.Time.Second())

	color := ColorLogLevel[r.Level]

	coloredLevel := fmt.Sprintf("%s%s%s", color, level, Reset)

	output := os.Stdout
	if r.Level >= slog.LevelWarn {
		output = os.Stderr
	}

	_, err := fmt.Fprintf(output, "%s %s %s\n", timeStr, coloredLevel, r.Message)
	if err != nil {
		fmt.Printf("Can't write log %s\n", err)
	}

	return nil
}

func SetLogLevel(logLevel string) string {
	upperLogLevel := strings.ToUpper(logLevel)

	level, found := LogLevels[upperLogLevel]
	if !found {
		upperLogLevel = "INFO"
		level = LevelInfo
	}

	opts := slog.HandlerOptions{
		Level: level,
	}

	handler := NewPrettyHandler(os.Stdout, opts)
	Log = &Logger{slog.New(handler)}

	return upperLogLevel
}

var Log *Logger

func Trace(msg string) {
	Log.Log(context.Background(), LevelTrace, msg)
}

func Debug(msg string) {
	Log.Log(context.Background(), LevelDebug, msg)
}

func Info(msg string) {
	Log.Log(context.Background(), LevelInfo, msg)
}
func Warn(msg string) {
	Log.Log(context.Background(), LevelWarn, msg)
}

func Error(msg string) {
	Log.Log(context.Background(), LevelError, msg)
}
