package main

import (
    "zylog"
    "zylog/internal/config"
)

// simpleExample is simple example, outputs logs to the console
func simpleExample() {
    zylog.Info("This is a info log")
}

// complexExample is complex example, you can control the output position of logs
func complexExample() {
    c := config.Config{
        CallerSkip: 2,
        LogConfig: []config.OutputConfig{
            {
               WriterName: config.OutputConsole,    // 标准输出
               Level:      config.LevelDebug,       // 日志的级别
               Formatter:  config.FormatterConsole, // 日志的格式
            },
            {
                WriterName: config.OutputFile,    // 本地文件日志
                Level:      config.LevelDebug,    // 日志的级别
                Formatter:  config.FormatterJson, // 日志的格式
                FormatConfig: config.FormatConfig{ // 日志格式内部字段定义
                    TimeFmt:       "2006-01-02 15:04:05", // 日志时间格式。"2006-01-02 15:04:05"为常规时间格式，"seconds"为秒级时间戳，"milliseconds"为毫秒时间戳，"nanoseconds"为纳秒时间戳
                    TimeKey:       "Time",                // 日志时间字段名称，不填默认"T"
                    LevelKey:      "Level",               // 日志级别字段名称，不填默认"L"
                    NameKey:       "Name",                // 日志名称字段名称， 不填默认"N"
                    CallerKey:     "Caller",              // 日志调用方字段名称， 不填默认"C"
                    FunctionKey:   "Function",            // 日志调用方字段名称， 不填默认不打印函数名
                    MessageKey:    "Message",             // 日志消息体字段名称，不填默认"M"
                    StacktraceKey: "StackTrace",          // 日志堆栈字段名称， 不填默认"S"
                },
                WriterConfig: config.WriteConfig{ // 本地文件输出具体配置
                    FileName:   "./test.log",      // 本地文件滚动日志存放的路径
                    WriteMode:  config.WriteSync,  // 日志写入模式，1-同步，2-异步，3-极速(异步丢弃), 不配置默认极速模式
                    RollType:   config.RollBySize, // 文件滚动类型,size为按大小滚动
                    MaxAge:     7,                 // 最大日志保留天数
                    MaxBackups: 10,                // 最大日志文件数
                    Compress:   false,             // 日志文件是否压缩
                    MaxSize:    10,                // 本地文件滚动日志的大小 单位 MB
                },
            },
        },
    }
    zylog.SetLoggerConfig(c)
    zylog.Info("This is a info log")
}

func main() {
    complexExample()
}
