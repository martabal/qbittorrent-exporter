package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Level int

var LogLevels = map[string]Level{
	"TRACE": Trace,
	"DEBUG": Debug,
	"INFO":  Info,
	"WARN":  Warn,
	"ERROR": Error,
}

const (
	Trace Level = -8
	Debug Level = -4
	Info  Level = 0
	Warn  Level = 4
	Error Level = 8
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	White  = "\033[97m"
)

type PrettyHandlerOptions struct {
	SlogOpts slog.HandlerOptions
}

type PrettyHandler struct {
	slog.Handler
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String()
	timeStr := r.Time.Format("[02/01/2023 15:04:05]")

	switch r.Level {
	case slog.LevelDebug:
		level = fmt.Sprintf(Green + level + Reset)
		fmt.Fprintf(os.Stdout, "%s %s %s\n", timeStr, level, r.Message)
	case slog.LevelInfo:
		level = fmt.Sprintf(Blue + level + Reset)
		fmt.Fprintf(os.Stdout, "%s %s %s\n", timeStr, level, r.Message)
	case slog.LevelWarn:
		level = fmt.Sprintf(Yellow + level + Reset)
		fmt.Fprintf(os.Stderr, "%s %s %s\n", timeStr, level, r.Message)
	case slog.LevelError:
		level = fmt.Sprintf(Red + level + Reset)
		fmt.Fprintf(os.Stderr, "%s %s %s\n", timeStr, level, r.Message)
	}

	return nil
}

func NewPrettyHandler(
	out io.Writer,
	opts PrettyHandlerOptions,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewTextHandler(out, &opts.SlogOpts),
	}

	return h
}

var Log *slog.Logger
