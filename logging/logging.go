/***********************************************
        File Name: logging
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/31/19 7:41 AM
***********************************************/

package logging

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DEBUG = 0
	INFO  = 1
	WARN  = 2
	ERROR = 3
	FATAL = 4
)

type Logger struct {
	FileName     string        `json:"filename" toml:"filename"`
	RollSize     int64         `json:"roll_size" toml:"roll_size"`
	RollInterval time.Duration `json:"roll_interval" toml:"roll_interval"`
	Level        int           `json:"level" toml:"level"`
	PanicOnFatal bool          `json:"panic_on_fatal" toml:"panic_on_fatal"`

	genName     func() string
	format      func(int, string, int) string
	dir         string
	idx         int
	roll        bool
	level       [5]string
	sizeCounter int64
	timePoint   time.Time
	file        *os.File
	mtx         sync.Mutex
}

var backend = Logger{}

func init() {
	backend.roll = false
	backend.PanicOnFatal = true
	backend.level = [5]string{"DEBUG", "INFO ", "WARN ", "ERROR", "FATAL"}
	backend.format = func(level int, file string, line int) string {
		// fuck golang
		return fmt.Sprintf("[%s %s] %s:%d ", time.Now().Format("2006-01-02 15:04:05.000"), backend.level[level], file, line)
	}
	backend.file = os.Stdout
}

func Init(cfg *Logger) error {
	backend.roll = true
	backend.RollSize = cfg.RollSize
	backend.RollInterval = cfg.RollInterval
	backend.Level = cfg.Level
	backend.PanicOnFatal = cfg.PanicOnFatal

	backend.dir = path.Dir(cfg.FileName)
	backend.FileName = path.Base(strings.TrimSuffix(cfg.FileName, filepath.Ext(cfg.FileName)))
	backend.sizeCounter = 0
	backend.timePoint = time.Now()
	backend.idx = 0

	prefix := fmt.Sprintf("%s-%s", backend.FileName, time.Now().Format("20060102"))
	entries, err := ioutil.ReadDir(backend.dir)
	if err != nil {
		return err
	}

	// /path/to/prefix-20190617-pid-1.log
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) {
			tmp, err := strconv.Atoi(strings.Split(strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())), "-")[3])
			if err != nil {
				return err
			}
			if backend.idx < tmp {
				backend.idx = tmp
			}
		}
	}

	backend.genName = func() string {
		cur := time.Now().Format("20060102")
		backend.idx += 1
		return fmt.Sprintf("%s/%s-%s-%d-%d.log", backend.dir, backend.FileName, cur, os.Getpid(), backend.idx)
	}

	backend.file, err = os.Create(backend.genName())
	return err
}

func (b *Logger) dispatch(level int, f string, args ...interface{}) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if level < b.Level {
		return
	}
	if b.roll && (b.sizeCounter > b.RollSize || time.Since(b.timePoint) > b.RollInterval) {
		b.sizeCounter = 0
		b.timePoint = time.Now()
		Release()
		b.file, _ = os.Create(b.genName())
	}

	_, file, line, _ := runtime.Caller(2) // golang is horrible
	data := []byte(b.format(level, path.Base(file), line) + fmt.Sprintf(f, args...))
	n, err := b.file.Write(append(data, '\n'))
	if err != nil {
		panic(err)
	}
	b.sizeCounter += int64(n)
}

func Release() {
	_ = backend.file.Sync()
	_ = backend.file.Close()
}

func Sync() error {
	return backend.file.Sync()
}

func Info(f string, args ...interface{}) {
	backend.dispatch(INFO, f, args...)
}

func Debug(f string, args ...interface{}) {
	backend.dispatch(DEBUG, f, args...)
}

func Warn(f string, args ...interface{}) {
	backend.dispatch(WARN, f, args...)
}

func Error(f string, args ...interface{}) {
	backend.dispatch(ERROR, f, args...)
}

func Fatal(f string, args ...interface{}) {
	backend.dispatch(FATAL, f, args...)
	if backend.PanicOnFatal {
		panic("fatal error")
	}
}
