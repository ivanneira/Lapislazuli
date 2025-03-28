package logger

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"sync"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var (
	enabled = true
	level   = DEBUG
	logger  = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	bufPool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

type LogFunc func(format string, v ...interface{})

var levelFuncs = map[Level]LogFunc{
	DEBUG: func(format string, v ...interface{}) {
		if enabled && level <= DEBUG {
			logger.Printf("[DEBUG] "+format, v...)
		}
	},
	INFO: func(format string, v ...interface{}) {
		if enabled && level <= INFO {
			logger.Printf("[INFO] "+format, v...)
		}
	},
	WARN: func(format string, v ...interface{}) {
		if enabled && level <= WARN {
			logger.Printf("[WARN] "+format, v...)
		}
	},
	ERROR: func(format string, v ...interface{}) {
		if enabled && level <= ERROR {
			logger.Printf("[ERROR] "+format, v...)
		}
	},
}

func SetEnabled(e bool) { enabled = e }
func SetLevel(l Level)  { level = l }

func Log(lvl Level, format string, v ...interface{}) {
	if fn, ok := levelFuncs[lvl]; ok {
		fn(format, v...)
	}
}

func Debug(format string, v ...interface{}) { Log(DEBUG, format, v...) }
func Info(format string, v ...interface{})  { Log(INFO, format, v...) }
func Warn(format string, v ...interface{})  { Log(WARN, format, v...) }
func Error(format string, v ...interface{}) { Log(ERROR, format, v...) }

func JSON(prefix string, v interface{}) {
	if !enabled || level > DEBUG {
		return
	}

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(v); err != nil {
		Error("Error marshaling JSON: %v", err)
		return
	}

	Debug("%s:\n%s", prefix, buf.String())
}

func init() {
	SetEnabled(true)
}
