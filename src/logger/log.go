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
	"TRACE": Trace,
	"DEBUG": Debug,
	"INFO":  Info,
	"WARN":  Warn,
	"ERROR": Error,
}

var ReverseLogLevels = map[slog.Level]string{
	Trace: "TRACE",
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Error: "ERROR",
}

var ColorLogLevel = map[slog.Level]string{
	Trace: Purple,
	Debug: Green,
	Info:  Blue,
	Warn:  Yellow,
	Error: Red,
}

const (
	Trace slog.Level = -8
	Debug slog.Level = -4
	Info  slog.Level = 0
	Warn  slog.Level = 4
	Error slog.Level = 8
)

const (
	Reset  string = "\033[0m"
	Red    string = "\033[31m"
	Green  string = "\033[32m"
	Yellow string = "\033[33m"
	Blue   string = "\033[34m"
	Purple string = "\033[35m"
)

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	level := ReverseLogLevels[slog.Level(r.Level)]
	timeStr := fmt.Sprintf("[%02d-%02d-%02d %02d:%02d:%02d]", r.Time.Year(), r.Time.Month(), r.Time.Day(), r.Time.Hour(), r.Time.Minute(), r.Time.Second())

	color := ColorLogLevel[r.Level]

	coloredLevel := fmt.Sprintf("%s%s%s", color, level, Reset)

	output := os.Stdout
	if r.Level >= slog.LevelWarn {
		output = os.Stderr
	}

	if _, err := fmt.Fprintf(output, "%s %s %s\n", timeStr, coloredLevel, r.Message); err != nil {
		fmt.Printf("Can't write log %s\n", err)
	}

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
		level = Info
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
