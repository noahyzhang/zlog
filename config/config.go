package config

import "time"

// Config is the log config
type Config struct {
    // LogConfig is the output config. Each log may have multiple outputs
    LogConfig []OutputConfig
    // CallerSkip controls the nesting depth of log function
    CallerSkip int
}

// WriteConfig is the local file config
type WriteConfig struct {
    // LogPath is the log path like /tmp/log/
    LogPath string
    // FileName is the file name like test.log
    FileName string
    // WriteWayMode is the log write mod. 1: sync, 2: async, 3: fast(maybe dropped)
    WriteMode WriteWayMode
    // RollType is the log rolling type. split files by size/time, default by size
    RollType RollType
    // MaxAge is the max expire times(day)
    MaxAge int
    // MaxSize is the max size of log file(MB)
    MaxSize int
    // MaxBackups is the max backup files
    MaxBackups int
    // Compress defines whether log should be compressed
    Compress bool
    // TimeUnit splits files by time unit, like year/month/hour/minute, default day.
    // It takes effect only when split by time.
    TimeUnit TimeUnit
}

// FormatConfig is the log format config.
type FormatConfig struct {
    // TimeFmt is the time format of log output, default as "2006-01-02 15:04:05.000" on empty.
    TimeFmt string

    // TimeKey is the time key of log output, default as "T".
    TimeKey string
    // LevelKey is the level key of log output, default as "L".
    LevelKey string
    // NameKey is the name key of log output, default as "N".
    NameKey string
    // CallerKey is the caller key of log output, default as "C".
    CallerKey string
    // FunctionKey is the function key of log output, default as "", which means not to print
    // function name.
    FunctionKey string
    // MessageKey is the message key of log output, default as "M".
    MessageKey string
    // StackTraceKey is the stack trace key of log output, default as "S".
    StacktraceKey string
}

// OutputConfig is the output config, includes console and file
type OutputConfig struct {
    // Writer is the output of log, such as console or file
    WriterName   WriterNameType
    WriterConfig WriteConfig

    // Formatter is the format of log, such as console or json
    Formatter    FormatterMode
    FormatConfig FormatConfig

    // LogLevel controls the log level, like trace, debug, info or error
    Level LogLevel
}

// ------------------- define ----------------------------

// WriterMode output name, default support console and file
type WriterNameType int

const (
    // OutputConsole write console
    OutputConsole WriterNameType = 1
    // OutputFile write file
    OutputFile WriterNameType = 2
)

func (n WriterNameType) ToString() string {
    switch n {
    case OutputConsole:
        return "console"
    case OutputFile:
        return "file"
    default:
        return "unknown"
    }
}

// FormatterType is the log format type
type FormatterMode int

const (
    // FormatterConsole formatter of console
    FormatterConsole FormatterMode = 1
    // FormatterJson formatter of json
    FormatterJson FormatterMode = 2
)

// WriteWayMode is the log write mode, one of 1, 2, 3
type WriteWayMode int

const (
    // WriteSync writes synchronously
    WriteSync WriteWayMode = 1
    // WriteAsync writes asynchronously
    WriteAsync WriteWayMode = 2
    // WriteFast writes fast(may drop logs asynchronously)
    WriteFast WriteWayMode = 3
)

// RollType is the log rolling type, one of 1, 2
type RollType int

const (
    // RollBySize rolls logs by file size
    RollBySize RollType = 1
    // RollByTime rolls logs by time
    RollByTime RollType = 2
)

// LogLevel is the log level
type LogLevel int

// Enums log level constants
const (
    LevelNil LogLevel = iota
    LevelDebug
    LevelInfo
    LevelWarn
    LevelError
    LevelFatal
)

// Some common used time formats.
const (
    // TimeFormatMinute is accurate to the minute.
    TimeFormatMinute = "%Y%m%d%H%M"
    // TimeFormatHour is accurate to the hour.
    TimeFormatHour = "%Y%m%d%H"
    // TimeFormatDay is accurate to the day.
    TimeFormatDay = "%Y%m%d"
    // TimeFormatMonth is accurate to the month.
    TimeFormatMonth = "%Y%m"
    // TimeFormatYear is accurate to the year.
    TimeFormatYear = "%Y"
)

// TimeUnit is the time unit by which files are split, one of minute/hour/day/month/year.
type TimeUnit string

const (
    // Minute splits by the minute.
    Minute = "minute"
    // Hour splits by the hour.
    Hour = "hour"
    // Day splits by the day.
    Day = "day"
    // Month splits by the month.
    Month = "month"
    // Year splits by the year.
    Year = "year"
)

// Format returns a string preceding with '.', Use TimeFormatDay as default
func (t TimeUnit) Format() string {
    var timeFmt string
    switch t {
    case Minute:
        timeFmt = TimeFormatMinute
    case Hour:
        timeFmt = TimeFormatHour
    case Day:
        timeFmt = TimeFormatDay
    case Month:
        timeFmt = TimeFormatMonth
    case Year:
        timeFmt = TimeFormatYear
    default:
        timeFmt = TimeFormatDay
    }
    return "." + timeFmt
}

// RotationGap returns the time.Duration for time unit. Use one day as the default.
func (t TimeUnit) RotationGap() time.Duration {
    switch t {
    case Minute:
        return time.Minute
    case Hour:
        return time.Hour
    case Day:
        return time.Hour * 24
    case Month:
        return time.Hour * 24 * 30
    case Year:
        return time.Hour * 24 * 365
    default:
        return time.Hour * 24
    }
}
