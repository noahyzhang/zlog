package rollwriter

import (
    "compress/gzip"
    "fmt"
    "io"
    "os"
    "strings"
    "time"
)

const (
    backupTimeFormat = "bk-20060102-150405.00000"
    compressSuffix   = ".gz"
)

// logInfo is an assistant struct which is used to return file name and last modified time
type logInfo struct {
    timestamp time.Time
    os.FileInfo
}

// byFormatTime sorts by time descending order.
type byFormatTime []logInfo

// Less checks whether the time of b[j] is early than the time of b[i].
func (b byFormatTime) Less(i, j int) bool {
    return b[i].timestamp.After(b[j].timestamp)
}

// Swap swaps b[i] and b[j].
func (b byFormatTime) Swap(i, j int) {
    b[i], b[j] = b[j], b[i]
}

// Len returns the length of list b.
func (b byFormatTime) Len() int {
    return len(b)
}

// filterByMaxBackups filters redundant files that exceeded the limit
func filterByMaxBackups(files []logInfo, remove *[]logInfo, maxBackups int) []logInfo {
    if maxBackups == 0 || len(files) < maxBackups {
        return files
    }
    var remaining []logInfo
    preserved := make(map[string]bool)
    for _, f := range files {
        fn := strings.TrimSuffix(f.Name(), compressSuffix)
        preserved[fn] = true

        if len(preserved) > maxBackups {
            *remove = append(*remove, f)
        } else {
            remaining = append(remaining, f)
        }
    }
    return remaining
}

// filterByMaxAge filters expired files
func filterByMaxAge(files []logInfo, remove *[]logInfo, maxAge int) []logInfo {
    if maxAge <= 0 {
        return files
    }
    var remaining []logInfo
    diff := time.Duration(int64(24 * time.Hour) * int64(maxAge))
    cutoff := time.Now().Add(-1 * diff)
    for _, f := range files {
        if f.timestamp.Before(cutoff) {
            *remove = append(*remove, f)
        } else {
            remaining = append(remaining, f)
        }
    }
    return remaining
}

// filterByCompressExt filters all compressed files
func filterByCompressExt(files []logInfo, compress *[]logInfo, needCompress bool) {
    if !needCompress {
        return
    }
    for _, f := range files {
        if !strings.HasSuffix(f.Name(), compressSuffix) {
            *compress = append(*compress, f)
        }
    }
}

// compressFile compress file src to dst, and removes src on success
func compressFile(src, dst string) error {
    f, err := os.Open(src)
    if err != nil {
        return fmt.Errorf("failed to open file: %v", err)
    }
    defer f.Close()

    gzf, err := os.OpenFile(dst, os.O_CREATE | os.O_TRUNC | os.O_WRONLY, 0666)
    if err != nil {
        return fmt.Errorf("failed to open compressed file: %v", err)
    }
    defer gzf.Close()

    gz := gzip.NewWriter(gzf)
    defer gz.Close()

    if _, err = io.Copy(gz, f); err != nil {
        _ = os.Remove(dst)
        return fmt.Errorf("fialed to compress file: %v", err)
    }
    _ = os.Remove(src)
    return nil
}