package logger

import (
    "sync"
    "github.com/noahyzhang/zlog/config"
    "github.com/noahyzhang/zlog/internal/writer"
)

func init() {
    writer.RegisterWriter(config.OutputConsole, writer.DefaultConsoleWriterFactory)
    writer.RegisterWriter(config.OutputFile, writer.DefaultFileWriterFactory)
    Register(defaultLoggerName, NewZapLog(DefaultConfig))
}

const defaultLoggerName = "default"

var (
    // DefaultLogger the default Logger. The initial output is console
    DefaultLogger Logger

    mu sync.RWMutex
    loggers = make(map[string]Logger)
)

// Register registers Logger. It supports multiple Logger implementation
func Register(name string, logger Logger) {
    mu.Lock()
    defer mu.Unlock()
    if logger == nil {
        panic("log: Register logger is nil")
    }
    if _, ok := loggers[name]; ok && name != defaultLoggerName {
        panic("log: Register called twined for logger name " + name)
    }
    loggers[name] = logger
    if name == defaultLoggerName {
        DefaultLogger = logger
    }
}

// GetDefaultLogger gets the default Logger
// The console output is the default value
func GetDefaultLogger() Logger {
    mu.RLock()
    logger := DefaultLogger
    mu.RUnlock()
    return logger
}

// SetDefaultLogger set the default Logger
func SetDefaultLogger(logger Logger) {
    mu.Lock()
    DefaultLogger = logger
    mu.Unlock()
}

