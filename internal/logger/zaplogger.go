package logger

import (
    "fmt"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "zylog/internal/config"
    "zylog/internal/writer"
)

var DefaultConfig = config.Config{
    LogConfig: []config.OutputConfig{
        {
            WriterName:    config.OutputConsole,
            Level:     config.LevelDebug,
            Formatter: config.FormatterConsole,
        },
    },
    CallerSkip: 2,
}

// NewZapLog creates a default Logger from zap whose caller skip is set to 2
func NewZapLog(c config.Config) Logger {
    return NewZapLogWithCallerSkip(c, c.CallerSkip)
}

// NewZapLogWithCallerSkip creates a default Logger from zap
func NewZapLogWithCallerSkip(c config.Config, callerSkip int) Logger {
    var (
        cores []zapcore.Core
        levels []zap.AtomicLevel
    )
    for _, o := range c.LogConfig {
        w := writer.GetWriter(o.WriterName)
        if w == nil {
            panic("log: writer core: " + o.WriterName.ToString() + " no registered")
        }
        core, zapLevel, err := w.Setup(&o)
        if err != nil {
            panic("log: writer core: " + o.WriterName.ToString() + " setup fail: " + err.Error())
        }
        cores = append(cores, core)
        levels = append(levels, zapLevel)
    }
    return &zapLog{
        levels: levels,
        logger: zap.New(
            zapcore.NewTee(cores...),
            zap.AddCallerSkip(callerSkip),
            zap.AddCaller()),
    }
}

// -------------------------- zapLog -------------------------------

// zapLog is a Logger implementation based on zapLogger
type zapLog struct {
    levels []zap.AtomicLevel
    logger *zap.Logger
}

func getLogMsg(args ...interface{}) string {
    return fmt.Sprint(args...)
}

func getLogMsgf(format string, args ...interface{}) string {
    return fmt.Sprintf(format, args...)
}

// Debug logs to DEBUG log. Arguments are handled in the manner of fmt.Print
func (l *zapLog) Debug(args ...interface{}) {
    if l.logger.Core().Enabled(zapcore.DebugLevel) {
        l.logger.Debug(getLogMsg(args))
    }
}

// Debugf logs to DEBUG log. Arguments are handled in the manner of fmt.Printf
func (l *zapLog) Debugf(format string, args ...interface{}) {
    if l.logger.Core().Enabled(zapcore.DebugLevel) {
        l.logger.Debug(getLogMsgf(format, args...))
    }
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print
func (l *zapLog) Info(args ...interface{}) {
    if l.logger.Core().Enabled(zapcore.InfoLevel) {
        l.logger.Info(getLogMsg(args...))
    }
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf
func (l *zapLog) Infof(format string, args ...interface{}) {
    if l.logger.Core().Enabled(zapcore.InfoLevel) {
        l.logger.Info(getLogMsgf(format, args...))
    }
}

// Warn logs to WARNING log. Arguments are handled in the manner of fmt.Print
func (l *zapLog) Warn(args ...interface{}) {
    if l.logger.Core().Enabled(zapcore.WarnLevel) {
        l.logger.Warn(getLogMsg(args...))
    }
}

// Warnf logs to WARNING log. Arguments are handled in the manner of fmt.Printf
func (l *zapLog) Warnf(format string, args ...interface{}) {
    if l.logger.Core().Enabled(zapcore.WarnLevel) {
        l.logger.Warn(getLogMsgf(format, args...))
    }
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print
func (l *zapLog) Error(args ...interface{}) {
    if l.logger.Core().Enabled(zapcore.ErrorLevel) {
        l.logger.Error(getLogMsg(args...))
    }
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf
func (l *zapLog) Errorf(format string, args ...interface{}) {
    if l.logger.Core().Enabled(zapcore.ErrorLevel) {
        l.logger.Error(getLogMsgf(format, args...))
    }
}

// Fatal logs to FATAL log. Arguments are handled in the manner of fmt.Print
func (l *zapLog) Fatal(args ...interface{}) {
    if l.logger.Core().Enabled(zapcore.FatalLevel) {
        l.logger.Fatal(getLogMsg(args...))
    }
}

// Fatalf logs to FATAL log. Arguments are handled in the manner of fmt.Printf
func (l *zapLog) Fatalf(format string, args ...interface{}) {
    if l.logger.Core().Enabled(zapcore.FatalLevel) {
        l.logger.Fatal(getLogMsgf(format, args...))
    }
}

// Sync calls the zap logger's Sync method, and flushes any buffered log entries
// Applications should take care to call Sync before exiting
func (l *zapLog) Sync() error {
    return l.logger.Sync()
}

// SetLevel sets output log level
func (l *zapLog) SetLevel(level config.LogLevel) {
    for i := 0; i < len(l.levels); i++ {
        l.levels[i].SetLevel(writer.LogLevelToZapLevel[level])
    }
}

// GetLevel gets output log level
func (l *zapLog) GetLevel() config.LogLevel {
    if len(l.levels) <= 0 {
        return config.LevelDebug
    }
    return writer.ZapLevelToLogLevel[l.levels[0].Level()]
}

// WithFields set some user defined data to logs, such as uid, imei, etc.
// Use this function at the beginning of each request. The returned new Logger should be used to
// print logs.
// Fields must be paired.
// Deprecated: use With instead.
func (l *zapLog) WithFields(fields ...string) Logger {
    zapFields := make([]zap.Field, len(fields)/2)
    for i := range zapFields {
        zapFields[i] = zap.String(fields[2*i], fields[2*i+1])
    }
    // By ZapLogWrapper proxy, we can add a layer to the debug series function calls, so that the
    // caller information can be set correctly.
    return &ZapLogWrapper{l: &zapLog{logger: l.logger.With(zapFields...), levels: l.levels}}
}

// With add user defined fields to Logger. Fields support multiple values
func (l *zapLog) With(fields ... Field) Logger {
    zapFields := make([]zap.Field, len(fields))
    for i := range fields {
        zapFields[i] = zap.Any(fields[i].Key, fields[i].Value)
    }
    // By ZapLogWrapper proxy, we can add a layer to the debug series function calls, so that the
    // caller information can be set correctly.
    return &ZapLogWrapper{l: &zapLog{logger: l.logger.With(zapFields...), levels: l.levels}}
}

// --------------------------- ZapLogWrapper ------------------------------------------

// ZapLogWrapper delegates zapLogger which was introduced in this
// [issue](https://git.code.oa.com/trpc-go/trpc-go/issues/260).
// By ZapLogWrapper proxy, we can add a layer to the debug series function calls, so that the caller
// information can be set correctly
type ZapLogWrapper struct {
    l *zapLog
}

// GetLogger returns interval zapLog.
func (z *ZapLogWrapper) GetLogger() Logger {
    return z.l
}

// Debug logs to DEBUG log. Arguments are handled in the manner of fmt.Print.
func (z *ZapLogWrapper) Debug(args ...interface{}) {
    z.l.Debug(args...)
}

// Debugf logs to DEBUG log. Arguments are handled in the manner of fmt.Printf.
func (z *ZapLogWrapper) Debugf(format string, args ...interface{}) {
    z.l.Debugf(format, args...)
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func (z *ZapLogWrapper) Info(args ...interface{}) {
    z.l.Info(args...)
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func (z *ZapLogWrapper) Infof(format string, args ...interface{}) {
    z.l.Infof(format, args...)
}

// Warn logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func (z *ZapLogWrapper) Warn(args ...interface{}) {
    z.l.Warn(args...)
}

// Warnf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func (z *ZapLogWrapper) Warnf(format string, args ...interface{}) {
    z.l.Warnf(format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func (z *ZapLogWrapper) Error(args ...interface{}) {
    z.l.Error(args...)
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func (z *ZapLogWrapper) Errorf(format string, args ...interface{}) {
    z.l.Errorf(format, args...)
}

// Fatal logs to FATAL log. Arguments are handled in the manner of fmt.Print.
func (z *ZapLogWrapper) Fatal(args ...interface{}) {
    z.l.Fatal(args...)
}

// Fatalf logs to FATAL log. Arguments are handled in the manner of fmt.Printf.
func (z *ZapLogWrapper) Fatalf(format string, args ...interface{}) {
    z.l.Fatalf(format, args...)
}

// Sync calls the zap logger's Sync method, and flushes any buffered log entries.
// Applications should take care to call Sync before exiting.
func (z *ZapLogWrapper) Sync() error {
    return z.l.Sync()
}

// SetLevel set output log level.
func (z *ZapLogWrapper) SetLevel(level config.LogLevel) {
    z.l.SetLevel(level)
}

// GetLevel gets output log level.
func (z *ZapLogWrapper) GetLevel() config.LogLevel {
    return z.l.GetLevel()
}

// WithFields set some user defined data to logs, such as uid, imei, etc.
// Use this function at the beginning of each request. The returned new Logger should be used to
// print logs.
// Fields must be paired.
// Deprecated: use With instead.
func (z *ZapLogWrapper) WithFields(fields ...string) Logger {
    return z.l.WithFields(fields...)
}

// With add user defined fields to Logger. Fields support multiple values.
func (z *ZapLogWrapper) With(fields ...Field) Logger {
    return z.l.With(fields...)
}