package logger

import (
	"log"
	"os"
)

type LogLevel string

const (
	Info  LogLevel = "info"
	Debug LogLevel = "debug"
	Warn  LogLevel = "warn"
	Error LogLevel = "error"
)

var levelMap = map[LogLevel]int{
	Info:  0,
	Debug: 1,
	Warn:  2,
	Error: 3,
}

type Logger struct {
	fLogger *log.Logger
	level   LogLevel
	file    *os.File
}

func NewLogger(fileName, prefix string, level LogLevel) *Logger {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to set up logging : " + err.Error())
	}
	fLogger := log.New(f, prefix, log.Ldate|log.Ltime|log.Lshortfile)
	return &Logger{
		fLogger: fLogger,
		level:   level,
		file:    f,
	}

}
func (l *Logger) Close() {
	l.file.Close()
}

func (l *Logger) hasPriority(level LogLevel) bool {
	if priority, ok := levelMap[level]; ok {
		appLoggerPriority, ok := levelMap[l.level]
		if !ok {
			l.fLogger.Println("invalid log level given")
			return false
		}
		if priority >= appLoggerPriority {
			return true
		}
	}
	return false

}

func (l *Logger) Info(lstring ...any) {
	if !l.hasPriority("info") {
		return
	}
	l.fLogger.Println(lstring...)
}

func (l *Logger) Debug(lstring ...any) {
	if !l.hasPriority("debug") {
		return
	}
	l.fLogger.Println(lstring...)
}

func (l *Logger) Warn(lstring ...any) {
	if !l.hasPriority("warn") {
		return
	}
	l.fLogger.Println(lstring...)
}

func (l *Logger) Error(lstring ...any) {
	if !l.hasPriority("error") {
		return
	}
	l.fLogger.Println(lstring...)
}

func (l *Logger) FInfo(format string, v ...any) {
	if !l.hasPriority("info") {
		return
	}
	l.fLogger.Printf(format, v...)
}

func (l *Logger) FDebug(format string, v ...any) {
	if !l.hasPriority("debug") {
		return
	}
	l.fLogger.Printf(format, v...)
}

func (l *Logger) FWarn(format string, v ...any) {
	if !l.hasPriority("warn") {
		return
	}
	l.fLogger.Printf(format, v...)
}

func (l *Logger) FError(format string, v ...any) {
	if !l.hasPriority("error") {
		return
	}
	l.fLogger.Printf(format, v...)
}
