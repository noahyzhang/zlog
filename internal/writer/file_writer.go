package writer

import (
    "github.com/noahyzhang/zlog/config"
    "github.com/noahyzhang/zlog/internal/rollwriter"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// DefaultFileWriterFactory is the default file output implementation
var DefaultFileWriterFactory = &FileWriterFactory{}

// FileWriterFactory is the file writer instance Factory
type FileWriterFactory struct {
}

func (f *FileWriterFactory) Setup(c *config.OutputConfig) (zapcore.Core, zap.AtomicLevel, error) {
    opts := []rollwriter.Option{
        rollwriter.WithMaxAge(c.WriterConfig.MaxAge),
        rollwriter.WithMaxBackups(c.WriterConfig.MaxBackups),
        rollwriter.WithCompress(c.WriterConfig.Compress),
        rollwriter.WithMaxSize(c.WriterConfig.MaxSize),
    }
    // roll by time
    if c.WriterConfig.RollType != config.RollBySize {
        opts = append(opts, rollwriter.WithRotationTime(c.WriterConfig.TimeUnit.Format()))
    }
    writer, err := rollwriter.NewRollWriter(c.WriterConfig.FileName, opts...)
    if err != nil {
        return nil, zap.AtomicLevel{}, err
    }
    // write mod
    var ws zapcore.WriteSyncer
    if c.WriterConfig.WriteMode == config.WriteSync {
        ws = zapcore.AddSync(writer)
    } else {
        dropLog := c.WriterConfig.WriteMode == config.WriteFast
        ws = rollwriter.NewAsyncRollWriter(writer, rollwriter.WithDropLog(dropLog))
    }
    // log level
    lvl := zap.NewAtomicLevelAt(LogLevelToZapLevel[c.Level])
    return zapcore.NewCore(newEncoder(c), ws, lvl), lvl, nil
}
