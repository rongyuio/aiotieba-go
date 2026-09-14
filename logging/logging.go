// Package logging provides the library logger.
//
// It mirrors the Python module aiotieba.logging but is built on log/slog.
package logging

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

var (
	mu        sync.RWMutex
	logger    *slog.Logger
	level     = slog.LevelDebug
	fileSink  io.WriteCloser
	formatter = func(w io.Writer) slog.Handler {
		return slog.NewTextHandler(w, &slog.HandlerOptions{Level: &level})
	}
)

// GetLogger returns the library logger, lazily creating a default one.
func GetLogger() *slog.Logger {
	mu.RLock()
	l := logger
	mu.RUnlock()
	if l != nil {
		return l
	}
	mu.Lock()
	defer mu.Unlock()
	if logger == nil {
		logger = newDefaultLogger()
	}
	return logger
}

// SetLogger replaces the library logger.
func SetLogger(l *slog.Logger) {
	mu.Lock()
	defer mu.Unlock()
	logger = l
}

// SetLevel sets the level of the default logger.
func SetLevel(l slog.Level) {
	mu.Lock()
	defer mu.Unlock()
	level = l
	if fileSink == nil {
		logger = newDefaultLogger()
	}
}

// EnableFileLog appends the log stream to logDir/<program>.log.
//
// It is the equivalent of Python's enable_filelog. Only the first call has an
// effect, mirroring the Python behaviour.
func EnableFileLog(logDir string) error {
	mu.Lock()
	defer mu.Unlock()

	if fileSink != nil {
		return nil
	}
	if logDir == "" {
		logDir = "log"
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}
	name := filepath.Base(os.Args[0])
	if ext := filepath.Ext(name); ext != "" {
		name = name[:len(name)-len(ext)]
	}
	f, err := os.OpenFile(filepath.Join(logDir, name+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	fileSink = f
	logger = slog.New(formatter(io.MultiWriter(os.Stdout, f)))
	return nil
}

func newDefaultLogger() *slog.Logger {
	return slog.New(formatter(os.Stdout))
}
