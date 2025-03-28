package logger

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"sync"
)

// Códigos de color ANSI
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
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
			logger.Printf(colorBlue+"[DEBUG] "+format+colorReset, v...)
		}
	},
	INFO: func(format string, v ...interface{}) {
		if enabled && level <= INFO {
			logger.Printf(colorGreen+"[INFO] "+format+colorReset, v...)
		}
	},
	WARN: func(format string, v ...interface{}) {
		if enabled && level <= WARN {
			logger.Printf(colorYellow+"[WARN] "+format+colorReset, v...)
		}
	},
	ERROR: func(format string, v ...interface{}) {
		if enabled && level <= ERROR {
			logger.Printf(colorRed+"[ERROR] "+format+colorReset, v...)
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

	Debug("=== %s ===\n%s", prefix, buf.String())
}

func init() {
	SetEnabled(true)
	SetLevel(DEBUG)                                // Asegurar que DEBUG esté activado por defecto
	logger.SetFlags(log.Ltime | log.Lmicroseconds) // Mostrar timestamps más precisos
}
