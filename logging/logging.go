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

type Config struct {
	Stderr       bool          `toml:"stdout"`
	FileName     string        `toml:"filename"`
	RollSize     int64         `toml:"roll_size"`
	RollInterval time.Duration `toml:"roll_interval"`
	Level        int           `toml:"level"`
}

func normalize(cfg Config) Config {
	return Config{
		Stderr:       cfg.Stderr,
		FileName:     path.Base(strings.TrimSuffix(cfg.FileName, filepath.Ext(cfg.FileName))),
		RollSize:     cfg.RollSize * (1 << 20),
		RollInterval: cfg.RollInterval * time.Hour,
		Level:        cfg.Level,
	}
}

type Ef struct {
	offset      int
	cfg         Config
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
	pid         int
}

var backend = Ef{}

func panicIf(err error, s string) {
	if err != nil {
		panic(fmt.Sprintf("%v: %v", s, err))
	}
}

func init() {
	backend.offset = 0
	backend.roll = false
	backend.pid = os.Getpid()
	backend.level = [5]string{"DEBUG", "INFO ", "WARN ", "ERROR", "FATAL"}
	backend.format = func(level int, path string, line int) string {
		// fuck golang
		return fmt.Sprintf("[%s %s]<%d> %s:%d ", time.Now().Format("2006-01-02 15:04:05.000"),
			backend.level[level], backend.pid, path[backend.offset:], line)
	}
	backend.file = os.Stderr
}

func Init(cfg Config, offset int) {
	if cfg.Stderr {
		return
	}
	var err error
	backend.offset = offset

	backend.cfg = normalize(cfg)
	backend.roll = true
	backend.dir = path.Dir(cfg.FileName)
	backend.sizeCounter = 0
	backend.timePoint = time.Now()
	backend.idx = 1

	_ = os.Mkdir(backend.dir, os.ModePerm)
	prefix := fmt.Sprintf("%s-%s", backend.cfg.FileName, time.Now().Format("20060102"))
	entries, err := ioutil.ReadDir(backend.dir)
	panicIf(err, "read log dir")

	// /path/to/prefix-20190617-pid-1.log
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) {
			sp := strings.Split(strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())), "-")
			if len(sp) < 4 {
				continue
			}
			tmp, err := strconv.Atoi(sp[3])
			if err != nil {
				continue
			}
			if backend.idx < tmp {
				backend.idx = tmp
			}
		}
	}

	backend.genName = func() string {
		cur := time.Now().Format("20060102")
		f := fmt.Sprintf("%s/%s-%s-%d.log", backend.dir, backend.cfg.FileName, cur, backend.idx)
		info, err := os.Stat(f)
		if os.IsNotExist(err) {
			return f
		}
		if info.Size() > backend.cfg.RollSize {
			backend.idx += 1
		}
		return fmt.Sprintf("%s/%s-%s-%d.log", backend.dir, backend.cfg.FileName, cur, backend.idx)
	}

	backend.file, err = os.OpenFile(backend.genName(), os.O_APPEND|os.O_WRONLY, 0644)
	panicIf(err, "create log file")
}

func (b *Ef) dispatch(level int, f string, args ...interface{}) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if level < b.cfg.Level {
		return
	}
	if b.roll && (b.sizeCounter > b.cfg.RollSize || time.Since(b.timePoint) > b.cfg.RollInterval) {
		b.sizeCounter = 0
		b.timePoint = time.Now()
		Release()
		b.file, _ = os.Create(b.genName())
	}

	_, file, line, _ := runtime.Caller(2) // golang is horrible
	data := []byte(b.format(level, file, line) + fmt.Sprintf(f, args...))
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
	panic("fatal error:" + fmt.Sprintf(f, args...))
}
