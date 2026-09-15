// Package logging 提供库的日志记录器。
//
// 对应 Python 模块 aiotieba.logging，但基于 zerolog 实现。
//
// 控制台输出使用 zerolog 的 ConsoleWriter：彩色、时间格式为 2006-01-02 15:04:05。
// 消息体由本包的 PyArgs / PyErr 渲染，形状与 Python 版一致：
//
//	[sign_forums] (340011, ''). args=() kwargs={}
//	[sign_forum] Succeeded. args=('盗墓笔记',) kwargs={}
package logging

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	mu       sync.RWMutex
	logger   *zerolog.Logger
	fileSink io.Closer
)

// GetLogger 获取日志记录器。
//
// 懒加载创建默认实例。
func GetLogger() *zerolog.Logger {
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
func SetLogger(l *zerolog.Logger) {
	mu.Lock()
	defer mu.Unlock()
	logger = l
}

// SetLevel 设置日志记录器的级别。
//
// 默认级别为 Debug，与 Python 版 StreamHandler 的默认级别一致。
func SetLevel(l zerolog.Level) {
	zerolog.SetGlobalLevel(l)
}

// EnableFileLog 启用文件日志，把日志追加写入 logDir/<程序名>.log。
//
// 参数:
//
//	logDir 用于存放日志文件的文件夹
//
// 把日志追写写入 logDir/<程序名>.log，等价于 Python 的 enable_filelog。
// 只有第一次调用生效，与 Python 行为一致。
//
// 与 Python 的 TimedRotatingFileHandler 不同，文件轮转交由 lumberjack 按大小触发：
// 单文件超过 10MB 时轮转，最多保留 5 份历史。文件日志无颜色且只记录 INFO 及以上，
// 控制台日志则保持彩色与 Debug 级别。
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

	sink := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, programName()+".log"),
		MaxSize:    10,    // 单文件上限，单位 MB
		MaxBackups: 5,     // 保留的历史文件数
		MaxAge:     0,     // 不按天数清理
		LocalTime:  true,  // 轮转文件名使用本地时间
		Compress:   false, // 保持纯文本，与 Python 一致
	}
	fileSink = sink

	l := zerolog.New(zerolog.MultiLevelWriter(
		newConsoleWriter(os.Stderr, false),
		minLevelWriter{w: newConsoleWriter(sink, true), min: zerolog.InfoLevel},
	)).With().Timestamp().Logger()
	logger = &l
	return nil
}

// newDefaultLogger 构建默认记录器：彩色控制台输出到 stderr。
func newDefaultLogger() *zerolog.Logger {
	l := zerolog.New(newConsoleWriter(os.Stderr, false)).With().Timestamp().Logger()
	return &l
}

// newConsoleWriter 构建控制台输出器，noColor 为真时不输出 ANSI 转义序列。
func newConsoleWriter(out io.Writer, noColor bool) zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out:        out,
		NoColor:    noColor,
		TimeFormat: time.DateTime,
	}
}

// minLevelWriter 只把不低于 min 级别的事件写入 w。
//
// zerolog 会优先调用 WriteLevel，因此文件日志可以拥有独立于控制台的级别门槛。
type minLevelWriter struct {
	w   io.Writer
	min zerolog.Level
}

func (m minLevelWriter) Write(p []byte) (int, error) { return m.w.Write(p) }

func (m minLevelWriter) WriteLevel(l zerolog.Level, p []byte) (int, error) {
	if l < m.min {
		return len(p), nil
	}
	return m.w.Write(p)
}

// programName 返回当前可执行文件名（去目录与扩展名），用作日志文件名。
func programName() string {
	name := filepath.Base(os.Args[0])
	if ext := filepath.Ext(name); ext != "" {
		name = name[:len(name)-len(ext)]
	}
	return name
}
