package writer

import (
    "github.com/noahyzhang/zlog/config"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "os"
)

// DefaultConsoleWriterFactory is the default console output implementation
var DefaultConsoleWriterFactory = &ConsoleWriterFactory{}

// ConsoleWriterFactory is the console internal instance
type ConsoleWriterFactory struct {
}

func (f *ConsoleWriterFactory) Setup(c *config.OutputConfig) (zapcore.Core, zap.AtomicLevel, error)  {
    lvl := zap.NewAtomicLevelAt(LogLevelToZapLevel[c.Level])
    return zapcore.NewCore(newEncoder(c), zapcore.Lock(os.Stdout), lvl), lvl, nil
}


