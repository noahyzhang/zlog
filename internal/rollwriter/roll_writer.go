package rollwriter

import (
    "errors"
    "fmt"
    "github.com/lestrrat-go/strftime"
    "io/ioutil"
    "os"
    "sort"
    "strings"
    "sync"
    "sync/atomic"
    "path/filepath"
    "time"
)

// RollWriter is a file log writer which support rolling by size or datetime.
// It implements io.WriteCloser.
type RollWriter struct {
    filePath string
    opts *Options

    pattern *strftime.Strftime
    currDir string
    currPath string
    currSize int64
    currFile atomic.Value
    openTime int64

    mu sync.Mutex
    once sync.Once
    notifyCh chan bool
    closeCh chan *os.File
}

// NewRollWriter creates a new RollWriter
func NewRollWriter(filePath string, opt ...Option) (*RollWriter, error) {
    opts := &Options{
        MaxSize:    0,     // default no rolling by file size
        MaxAge:     0,     // default no scavenging on expired logs
        MaxBackups: 0,     // default no scavenging on redundant logs
        Compress:   false, // default no compressing
    }

    // opt has the highest priority and should overwrite the original and default one
    for _, o := range opt {
        o(opts)
    }
    if filePath == "" {
        return nil, errors.New("invalid file path")
    }
    patter, err := strftime.New(filePath + opts.TimeFormat)
    if err != nil {
        return nil, errors.New("invalid time pattern")
    }
    w := &RollWriter{
        filePath: filePath,
        opts: opts,
        pattern: patter,
        currDir: filepath.Dir(filePath),
    }
    if err := os.MkdirAll(w.currDir, 0755); err != nil {
        return nil, err
    }
    return w, nil
}

// Write writes logs. It implements io.Writer
func (w *RollWriter) Write(v []byte) (int, error) {
    // reopen file every 10 seconds
    if w.getCurrFile() == nil || time.Now().Unix() - atomic.LoadInt64(&w.openTime) > 10 {
        w.mu.Lock()
        w.reopenFile()
        w.mu.Unlock()
    }
    // return when failed to open the file
    if w.getCurrFile() == nil {
        return 0, errors.New("open file fail")
    }
    // write logs to file
    n, err := w.getCurrFile().Write(v)
    atomic.AddInt64(&w.currSize, int64(n))

    // rolling on full
    if w.opts.MaxSize > 0 && atomic.LoadInt64(&w.currSize) >= w.opts.MaxSize {
        w.mu.Lock()
        w.backupFile()
        w.mu.Unlock()
    }
    return n, err
}

// Close close the current log file. It implements io.Closer
func (w *RollWriter) Close() error {
    if w.getCurrFile() == nil {
        return nil
    }
    err := w.getCurrFile().Close()
    w.setCurrFile(nil)
    return err
}

// reopenFile reopen the file regularly. It notifies the scavenger if file path has changed
func (w *RollWriter) reopenFile() {
    if w.getCurrFile() == nil || time.Now().Unix() - atomic.LoadInt64(&w.openTime) > 10 {
        atomic.StoreInt64(&w.openTime, time.Now().Unix())
        currPath := w.pattern.FormatString(time.Now())
        if w.currPath != currPath {
            w.currPath = currPath
            w.notify()
        }
        _ = w.doReopenFile(w.currPath)
    }
}

// backupFile backs this file up and reopen a new one if file size is too large
func (w *RollWriter) backupFile() {
    if w.opts.MaxSize > 0 && atomic.LoadInt64(&w.currSize) >= w.opts.MaxSize {
        atomic.StoreInt64(&w.currSize, 0)
        // rename the old file
        NewName := w.currPath + "." + time.Now().Format(backupTimeFormat)
        if _, e := os.Stat(w.currPath); !os.IsNotExist(e) {
            _ = os.Rename(w.currPath, NewName)
        }
        // reopen a new one
        _ = w.doReopenFile(w.currPath)
        w.notify()
    }
}

// notify runs scavengers
func (w *RollWriter) notify() {
    w.once.Do(func() {
        w.notifyCh = make(chan bool, 1)
        w.closeCh = make(chan *os.File, 100)
        go w.runCleanFiles()
        go w.runCloseFiles()
    })
    select {
    case w.notifyCh <- true:
    default:
    }
}

// runCloseFiles delay closing file in a new goroutine
func (w *RollWriter) runCloseFiles() {
    for f := range w.closeCh {
        // 延迟 20ms 关闭
        time.Sleep(20*time.Millisecond)
        _ = f.Close()
    }
}

// runCleanFiles cleans redundant or expired (compressed) logs in a new goroutine
func (w *RollWriter) runCleanFiles() {
    for range w.notifyCh {
        if w.opts.MaxBackups == 0 && w.opts.MaxAge == 0 && !w.opts.Compress {
            continue
        }
        w.cleanFiles()
    }
}

// cleanFiles cleans redundant or expired (compressed) logs
func (w *RollWriter) cleanFiles() {
    // get the file list of current log
    files, err := w.getOldLogFiles()
    if err != nil || len(files) == 0 {
        return
    }

    // find the oldest files to scavenge
    var compress, remove []logInfo
    files = filterByMaxBackups(files, &remove, w.opts.MaxBackups)
    // find the expired files by last modified time
    files = filterByMaxAge(files, &remove, w.opts.MaxAge)
    // find files to compress by file extension ".gz"
    filterByCompressExt(files, &compress, w.opts.Compress)
    // delete expired or redundant
    w.removeFiles(remove)
    // compress log files
    w.compressFile(compress)
}

// removeFiles deletes expired or redundant log files
func (w *RollWriter) removeFiles(remove []logInfo) {
    // clean expired or redundant files
    for _, f := range remove {
        _ = os.Remove(filepath.Join(w.currDir, f.Name()))
    }
}

// compressFiles compresses demanded log files.
func (w *RollWriter) compressFile(compress []logInfo) {
    // compress log files
    for _, f := range compress {
        fn := filepath.Join(w.currDir, f.Name())
        _ = compressFile(fn, fn+compressSuffix)
    }
}

// getOldLogFiles returns the log file list ordered by modified time
func (w *RollWriter) getOldLogFiles() ([]logInfo, error) {
    files, err := ioutil.ReadDir(w.currDir)
    if err != nil {
        return nil, fmt.Errorf("can't read log file directory: %s", err)
    }
    var logFiles []logInfo
    fileName := filepath.Base(w.filePath)
    for _, f := range files {
        if f.IsDir() {
            continue
        }
        if modTime, err := w.matchLogFile(f.Name(), fileName); err == nil {
            logFiles = append(logFiles, logInfo{modTime, f})
        }
    }
    sort.Sort(byFormatTime(logFiles))
    return logFiles, nil
}

// matchLogFile checks whether current log file matches all relative log files, if matched, returns the modified time
func (w *RollWriter) matchLogFile(filename, filePrefix string) (time.Time, error) {
    // exclude current log file
    // a.log
    // a.log.20220624
    if filepath.Base(w.currPath) == filename {
        return time.Time{}, errors.New("ignore current logfile")
    }
    // check if there is the same prefix
    if !strings.HasPrefix(filename, filePrefix) {
        return time.Time{}, errors.New("mismatched prefix")
    }
    if st, _ := os.Stat(filepath.Join(w.currDir, filename)); st != nil {
        return st.ModTime(), nil
    }
    return time.Time{}, errors.New("file stat fail")
}

// getCurrFile returns the current log file
func (w *RollWriter) getCurrFile() *os.File {
    if file, ok := w.currFile.Load().(*os.File); ok {
        return file
    }
    return nil
}

// doReopenFile reopen the file
func (w *RollWriter) doReopenFile(path string) error {
    atomic.StoreInt64(&w.openTime, time.Now().Unix())
    lastFile := w.getCurrFile()
    of, err := os.OpenFile(path, os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0666)
    if err != nil {
        return err
    }
    w.setCurrFile(of)
    if lastFile != nil {
        // delay closing until not used
        w.closeCh <- lastFile
    }
    st, _ := os.Stat(path)
    if st != nil {
        atomic.StoreInt64(&w.currSize, st.Size())
    }
    return nil
}

// setCurrFile sets the current log file
func (w *RollWriter) setCurrFile(file *os.File) {
    w.currFile.Store(file)
}













