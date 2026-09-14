// Package logging 提供库的日志记录器。
//
// 对应 Python 模块 aiotieba.logging，但基于 log/slog 实现。
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

// GetLogger 获取日志记录器。
//
// 懒加载创建默认实例。
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

// SetLogger 更换 aiotieba 的日志记录器。
func SetLogger(l *slog.Logger) {
	mu.Lock()
	defer mu.Unlock()
	logger = l
}

// SetLevel 设置默认日志记录器的级别。
func SetLevel(l slog.Level) {
	mu.Lock()
	defer mu.Unlock()
	level = l
	if fileSink == nil {
		logger = newDefaultLogger()
	}
}

// EnableFileLog 启用文件日志，把日志追加写入 logDir/<程序名>.log。
//
// 参数:
//
//	logDir 用于存放日志文件的文件夹
//
// 把日志追写写入 logDir/<程序名>.log，等价于 Python 的 enable_filelog。
// 只有第一次调用生效，与 Python 行为一致。
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
