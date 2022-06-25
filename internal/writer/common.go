package writer

import (
    "github.com/noahyzhang/zlog/config"
    "go.uber.org/zap/zapcore"
    "time"
)

var LogLevelToZapLevel = map[config.LogLevel]zapcore.Level {
    config.LevelDebug: zapcore.DebugLevel,
    config.LevelInfo:  zapcore.InfoLevel,
    config.LevelWarn:  zapcore.WarnLevel,
    config.LevelError: zapcore.ErrorLevel,
    config.LevelFatal: zapcore.FatalLevel,
}

var ZapLevelToLogLevel = map[zapcore.Level]config.LogLevel{
    zapcore.DebugLevel: config.LevelDebug,
    zapcore.InfoLevel:  config.LevelInfo,
    zapcore.WarnLevel:  config.LevelWarn,
    zapcore.ErrorLevel: config.LevelError,
    zapcore.FatalLevel: config.LevelFatal,
}

var LogLevelToString = map[config.LogLevel]string {
    config.LevelDebug: "debug",
    config.LevelInfo:  "info",
    config.LevelWarn:  "warn",
    config.LevelError: "error",
    config.LevelFatal: "fatal",
}

// GetLogEncoderKey gets user defined log output name, uses defKey if empty.
func GetLogEncoderKey(defKey, key string) string {
    if key == "" {
        return defKey
    }
    return key
}

// DefaultTimeFormat returns the default time format.
func DefaultTimeFormat(t time.Time) []byte {
    t = t.Local()
    year, month, day := t.Date()
    hour, minute, second := t.Clock()
    micros := t.Nanosecond() / 1000

    buf := make([]byte, 23)
    buf[0] = byte((year/1000)%10) + '0'
    buf[1] = byte((year/100)%10) + '0'
    buf[2] = byte((year/10)%10) + '0'
    buf[3] = byte(year%10) + '0'
    buf[4] = '-'
    buf[5] = byte((month)/10) + '0'
    buf[6] = byte((month)%10) + '0'
    buf[7] = '-'
    buf[8] = byte((day)/10) + '0'
    buf[9] = byte((day)%10) + '0'
    buf[10] = ' '
    buf[11] = byte((hour)/10) + '0'
    buf[12] = byte((hour)%10) + '0'
    buf[13] = ':'
    buf[14] = byte((minute)/10) + '0'
    buf[15] = byte((minute)%10) + '0'
    buf[16] = ':'
    buf[17] = byte((second)/10) + '0'
    buf[18] = byte((second)%10) + '0'
    buf[19] = '.'
    buf[20] = byte((micros/100000)%10) + '0'
    buf[21] = byte((micros/10000)%10) + '0'
    buf[22] = byte((micros/1000)%10) + '0'
    return buf
}

// CustomTimeFormat customize time format.
func CustomTimeFormat(t time.Time, format string) string {
    return t.Format(format)
}

// NewTimeEncoder creates a time format encoder.
func NewTimeEncoder(format string) zapcore.TimeEncoder {
    switch format {
    case "":
        return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
            enc.AppendByteString(DefaultTimeFormat(t))
        }
    case "seconds":
        return zapcore.EpochTimeEncoder
    case "milliseconds":
        return zapcore.EpochMillisTimeEncoder
    case "nanoseconds":
        return zapcore.EpochNanosTimeEncoder
    default:
        return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
            enc.AppendString(CustomTimeFormat(t, format))
        }
    }
}

func newEncoder(c *config.OutputConfig) zapcore.Encoder {
    encoderCfg := zapcore.EncoderConfig{
        TimeKey:        GetLogEncoderKey("T", c.FormatConfig.TimeKey),
        LevelKey:       GetLogEncoderKey("L", c.FormatConfig.LevelKey),
        NameKey:        GetLogEncoderKey("N", c.FormatConfig.NameKey),
        CallerKey:      GetLogEncoderKey("C", c.FormatConfig.CallerKey),
        FunctionKey:    GetLogEncoderKey(zapcore.OmitKey, c.FormatConfig.FunctionKey),
        MessageKey:     GetLogEncoderKey("M", c.FormatConfig.MessageKey),
        StacktraceKey:  GetLogEncoderKey("S", c.FormatConfig.StacktraceKey),
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeLevel:    zapcore.CapitalLevelEncoder,
        EncodeTime:     NewTimeEncoder(c.FormatConfig.TimeFmt),
        EncodeDuration: zapcore.StringDurationEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }
    switch c.Formatter {
    case config.FormatterConsole:
        return zapcore.NewConsoleEncoder(encoderCfg)
    case config.FormatterJson:
        return zapcore.NewJSONEncoder(encoderCfg)
    default:
        return zapcore.NewConsoleEncoder(encoderCfg)
    }
}