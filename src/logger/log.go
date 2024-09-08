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

var LogLevels = map[string]int{
	"TRACE": Trace,
	"DEBUG": Debug,
	"INFO":  Info,
	"WARN":  Warn,
	"ERROR": Error,
}

var ReverseLogLevels = map[int]string{
	Trace: "TRACE",
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Error: "ERROR",
}

const (
	Trace int = -8
	Debug int = -4
	Info  int = 0
	Warn  int = 4
	Error int = 8
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
)

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	level := ReverseLogLevels[int(r.Level)]
	timeStr := fmt.Sprintf("[%02d-%02d-%02d %02d:%02d:%02d]", r.Time.Year(), r.Time.Month(), r.Time.Day(), r.Time.Hour(), r.Time.Minute(), r.Time.Second())

	var color string
	switch r.Level {
	case slog.Level(Trace):
		color = Purple
	case slog.LevelDebug:
		color = Green
	case slog.LevelInfo:
		color = Blue
	case slog.LevelWarn:
		color = Yellow
	case slog.LevelError:
		color = Red
	}

	coloredLevel := fmt.Sprintf("%s%s%s", color, level, Reset)

	output := os.Stdout
	if r.Level >= slog.LevelWarn {
		output = os.Stderr
	}

	fmt.Fprintf(output, "%s %s %s\n", timeStr, coloredLevel, r.Message)

	return nil
}

func NewPrettyHandler(
	out io.Writer,
	opts slog.HandlerOptions,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewTextHandler(out, &opts),
	}

	return h
}

func SetLogLevel(logLevel string) string {
	upperLogLevel := strings.ToUpper(logLevel)
	level, found := LogLevels[upperLogLevel]
	if !found {
		upperLogLevel = "INFO"
		level = LogLevels[upperLogLevel]
	}

	opts := slog.HandlerOptions{
		Level: slog.Level(level),
	}

	handler := NewPrettyHandler(os.Stdout, opts)
	Log = &Logger{slog.New(handler)}
	return upperLogLevel
}

func (l *Logger) Trace(msg string) {
	l.Log(context.Background(), slog.Level(Trace), msg)
}

var Log *Logger
