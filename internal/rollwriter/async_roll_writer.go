package rollwriter

import (
    "bytes"
    "errors"
    "io"
    "time"
)

// AsyncRollWriter is the asynchronous rolling log writer which implements zapcore.WriteSyncer
type AsyncRollWriter struct {
    logger io.Writer
    opts *AsyncOptions

    logQueue chan []byte
    syncChan chan struct{}
}

// NewAsyncRollWriter create a new AsyncRollWriter
func NewAsyncRollWriter(logger io.Writer, opt ...AsyncOption) *AsyncRollWriter {
    opts := &AsyncOptions{
        LogQueueSize:     10000,    // default queue size as 10000
        WriteLogSize:     4 * 1024, // default write log size as 4K
        WriteLogInterval: 100,      // default sync interval as 100ms
        DropLog:          false,    // default do not drop logs
    }
    for _, o := range opt {
        o(opts)
    }
    w := &AsyncRollWriter{
        logger: logger,
        opts: opts,
        logQueue: make(chan []byte, opts.LogQueueSize),
        syncChan: make(chan struct{}),
    }
    go w.batchWriteLog()
    return w
}

// Write writes logs. It implements io.Writer
func (w *AsyncRollWriter) Write(data []byte) (int, error) {
    log := make([]byte, len(data))
    copy(log, data)
    if w.opts.DropLog {
        select {
        case w.logQueue <- log:
        default:
            return 0, errors.New("log queue is full")
        }
    } else {
        w.logQueue <- log
    }
    return len(data), nil
}

// Sync syncs logs. It implements zapcore.WriteSyncer
func (w *AsyncRollWriter) Sync() error {
    w.syncChan <- struct{}{}
    return nil
}

// Close closes current log file. It implement io.Closer
func (w *AsyncRollWriter) Close() error {
    return w.Sync()
}

// batchWriteLog asynchronously writers logs in batches
func (w *AsyncRollWriter) batchWriteLog() {
    buffer := bytes.NewBuffer(make([]byte, 0, w.opts.WriteLogSize * 2))
    ticker := time.NewTicker(time.Millisecond * time.Duration(w.opts.WriteLogInterval))
    for {
        select {
        case <- ticker.C:
            if buffer.Len() > 0 {
                _, _ = w.logger.Write(buffer.Bytes())
                buffer.Reset()
            }
        case data := <-w.logQueue:
            buffer.Write(data)
            if buffer.Len() >= w.opts.WriteLogSize {
                _, _ = w.logger.Write(buffer.Bytes())
                buffer.Reset()
            }
        case <-w.syncChan:
            if buffer.Len() > 0 {
                _, _ = w.logger.Write(buffer.Bytes())
                buffer.Reset()
            }
            size := len(w.logQueue)
            for i := 0; i < size; i++ {
                v := <- w.logQueue
                _, _ = w.logger.Write(v)
            }
        }
    }
}
