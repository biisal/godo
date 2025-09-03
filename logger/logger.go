package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

type LogLevel string

const (
	Debug LogLevel = "debug"
	Info  LogLevel = "info"
	Warn  LogLevel = "warn"
	Error LogLevel = "error"
)

var levelMap = map[LogLevel]int{
	Debug: 0,
	Info:  1,
	Warn:  2,
	Error: 3,
}

var levelPrefix = map[LogLevel]string{
	Debug: "[DEBUG]",
	Info:  "[INFO]",
	Warn:  "[WARN]",
	Error: "[ERROR]",
}

type Logger struct {
	fLogger   *log.Logger
	level     LogLevel
	file      *os.File
	callDepth int
}

func NewLogger(fileName, prefix string, level LogLevel) *Logger {
	return newLogger(fileName, prefix, level, false)
}

func NewLoggerWithConsole(fileName, prefix string, level LogLevel) *Logger {
	return newLogger(fileName, prefix, level, true)
}

func newLogger(fileName, prefix string, level LogLevel, toStdout bool) *Logger {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to set up logging: " + err.Error())
	}

	var writer io.Writer = f
	if toStdout {
		writer = io.MultiWriter(os.Stdout, f)
	}

	lg := log.New(writer, prefix, log.Ldate|log.Ltime|log.Lshortfile)

	return &Logger{
		fLogger:   lg,
		level:     level,
		file:      f,
		callDepth: 3}
}

func (l *Logger) Close() error {
	if l.file == nil {
		return nil
	}
	return l.file.Close()
}

func (l *Logger) hasPriority(msgLevel LogLevel) bool {
	msgPri, ok := levelMap[msgLevel]
	if !ok {
		l.fLogger.Println("[LOGGER] invalid message log level:", string(msgLevel))
		return false
	}
	appPri, ok := levelMap[l.level]
	if !ok {
		l.fLogger.Println("[LOGGER] invalid configured log level:", string(l.level))
		return false
	}
	return msgPri >= appPri
}

func (l *Logger) format(level LogLevel, msg string) string {
	p, ok := levelPrefix[level]
	if !ok {
		p = "[UNKNOWN]"
	}
	return fmt.Sprintf("%s %s", p, msg)
}

func (l *Logger) logln(level LogLevel, v ...any) {
	if !l.hasPriority(level) {
		return
	}
	msg := fmt.Sprint(v...)
	_ = l.fLogger.Output(l.callDepth, l.format(level, msg))
}

func (l *Logger) logf(level LogLevel, format string, v ...any) {
	if !l.hasPriority(level) {
		return
	}
	msg := fmt.Sprintf(format, v...)
	_ = l.fLogger.Output(l.callDepth, l.format(level, msg))
}

func (l *Logger) Debug(v ...any) { l.logln(Debug, v...) }
func (l *Logger) Info(v ...any)  { l.logln(Info, v...) }
func (l *Logger) Warn(v ...any)  { l.logln(Warn, v...) }
func (l *Logger) Error(v ...any) { l.logln(Error, v...) }

func (l *Logger) FDebug(format string, v ...any) { l.logf(Debug, format, v...) }
func (l *Logger) FInfo(format string, v ...any)  { l.logf(Info, format, v...) }
func (l *Logger) FWarn(format string, v ...any)  { l.logf(Warn, format, v...) }
func (l *Logger) FError(format string, v ...any) { l.logf(Error, format, v...) }
