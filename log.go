package zylog

import (
    "zylog/internal/config"
    "zylog/internal/logger"
    "zylog/internal/writer"
)

// SetLoggerConfig set logger of use our config
func SetLoggerConfig(c config.Config) {
    l := logger.NewZapLog(c)
    logger.SetDefaultLogger(l)
}

// SetLevel sets log level for different output
func SetLevel(level config.LogLevel) {
    logger.GetDefaultLogger().SetLevel(level)
}

// GetIntLevel gets log level(int) for output
func GetLevel() config.LogLevel {
    return logger.GetDefaultLogger().GetLevel()
}

// GetStringLevel gets log level(string) for output
func GetLevelWithString() string {
    return writer.LogLevelToString[logger.GetDefaultLogger().GetLevel()]
}

// WithFields sets some user defined data to logs, such as, uid, imei. Fields must be paired.
// Deprecated: use With instead.
func WithFields(fields ...string) {
    logger.SetDefaultLogger(logger.GetDefaultLogger().WithFields(fields...))
}

// With adds user defined fields to Logger. Field support multiple values.
func With(fields ...logger.Field) {
    logger.SetDefaultLogger(logger.GetDefaultLogger().With(fields...))
}

// Debug logs to DEBUG log. Arguments are handled in the manner of fmt.Print.
func Debug(args ...interface{}) {
    logger.GetDefaultLogger().Debug(args...)
}

// Debugf logs to DEBUG log. Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, args ...interface{}) {
    logger.GetDefaultLogger().Debugf(format, args...)
}

// Info logs to INFO log. Arguments are handled in the manner of fmt.Print.
func Info(args ...interface{}) {
    logger.GetDefaultLogger().Info(args...)
}

// Infof logs to INFO log. Arguments are handled in the manner of fmt.Printf.
func Infof(format string, args ...interface{}) {
    logger.GetDefaultLogger().Infof(format, args...)
}

// Warn logs to WARNING log. Arguments are handled in the manner of fmt.Print.
func Warn(args ...interface{}) {
    logger.GetDefaultLogger().Warn(args...)
}

// Warnf logs to WARNING log. Arguments are handled in the manner of fmt.Printf.
func Warnf(format string, args ...interface{}) {
    logger.GetDefaultLogger().Warnf(format, args...)
}

// Error logs to ERROR log. Arguments are handled in the manner of fmt.Print.
func Error(args ...interface{}) {
    logger.GetDefaultLogger().Error(args...)
}

// Errorf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, args ...interface{}) {
    logger.GetDefaultLogger().Errorf(format, args...)
}

// Fatal logs to ERROR log. Arguments are handled in the manner of fmt.Print.
// All Fatal logs will exit by calling os.Exit(1).
// Implementations may also call os.Exit() with a non-zero exit code.
func Fatal(args ...interface{}) {
    logger.GetDefaultLogger().Fatal(args...)
}

// Fatalf logs to ERROR log. Arguments are handled in the manner of fmt.Printf.
func Fatalf(format string, args ...interface{}) {
    logger.GetDefaultLogger().Fatalf(format, args...)
}

// Sync writes logs that are still in the cache to disk
func Sync() error {
    return logger.GetDefaultLogger().Sync()
}