package writer

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "zylog/internal/config"
)

var (
    writers = make(map[config.WriterNameType]Factory)
)

// RegisterWriter registers log output writer. Writer may have multiple implementations
func RegisterWriter(name config.WriterNameType, writer Factory) {
    writers[name] = writer
}

// GetWriter gets log output writer, returns nil if not exist
func GetWriter(name config.WriterNameType) Factory {
    return writers[name]
}

type Factory interface {
    Setup(c *config.OutputConfig) (zapcore.Core, zap.AtomicLevel, error)
}

