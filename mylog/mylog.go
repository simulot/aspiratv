package mylog

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Logger interface {
	Printf(string, ...interface{})
}

type Level int

const (
	LevelFatal Level = -2
	LevelError       = iota
	LevelInfo
	LevelTrace
	LevelDebug
)

var levelStrings = map[string]Level{
	"FATAL": LevelFatal,
	"ERROR": LevelError,
	"INFO":  LevelInfo,
	"TRACE": LevelTrace,
	"DEBUG": LevelDebug,
}

var prefixes = map[Level]string{
	LevelFatal: "[FATAL] ",
	LevelError: "[ERROR] ",
	LevelInfo:  "[INFO ] ",
	LevelTrace: "[TRACE] ",
	LevelDebug: "[DEBUG] ",
}

type MyLog struct {
	logLevel                  Level
	consoleLogger, fileLogger Logger
}

// NewLog return a MyLog structure
func NewLog(lvl string, consoleLogger, fileLogger Logger) (*MyLog, error) {
	var (
		level Level
		ok    bool
	)

	if level, ok = levelStrings[strings.ToUpper(lvl)]; !ok {
		return nil, fmt.Errorf("Invalid log level '%s'", lvl)
	}

	// intercept all direct call to log as in http package
	log.SetOutput(ioutil.Discard)

	return &MyLog{
		logLevel:      level,
		consoleLogger: consoleLogger,
		fileLogger:    fileLogger,
	}, nil
}

// Fatal prepare the output of FATAL message
func (l *MyLog) Fatal() logcontext {
	return logcontext{l, LevelFatal}
}

// Error prepare the output of ERROR message
func (l *MyLog) Error() logcontext {
	return logcontext{l, LevelError}
}

// Info prepare the output of INFO message
func (l *MyLog) Info() logcontext {
	return logcontext{l, LevelInfo}
}

// Trace prepare the output of TRACE message
func (l *MyLog) Trace() logcontext {
	return logcontext{l, LevelTrace}
}

// Debug prepare the output of DEBUG message
func (l *MyLog) Debug() logcontext {
	return logcontext{l, LevelDebug}
}

// IsDebug return true if log level is DEBUG
func (l *MyLog) IsDebug() bool {
	if l == nil {
		return true
	}
	return l.logLevel >= LevelDebug
}

// logcontext get the level of current message
type logcontext struct {
	mylog *MyLog
	lvl   Level
}

// Printf print message on configured writers
// When a log file writer is provided, only errors are written on
// console writer
// When the message is FATAL, the message is written on writers and the
// program exits
// If the logger isn't initialized, it logs to the console
func (c logcontext) Printf(fmt string, args ...interface{}) {
	if c.mylog == nil {
		if c.lvl == LevelFatal {
			log.Fatalf(prefixes[c.lvl]+fmt, args...)
		} else {
			log.Printf(prefixes[c.lvl]+fmt, args...)
		}
		return
	}
	if c.lvl <= LevelError && c.mylog.consoleLogger != nil {
		c.mylog.consoleLogger.Printf(prefixes[c.lvl]+fmt, args...)
	}
	if c.mylog.fileLogger != nil && c.lvl <= c.mylog.logLevel {
		c.mylog.fileLogger.Printf(prefixes[c.lvl]+fmt, args...)
	}
	if c.lvl == LevelFatal {
		if c.mylog.consoleLogger == nil {
			log.Printf(prefixes[c.lvl]+fmt, args...)
		}
		os.Exit(1)
	}
}
