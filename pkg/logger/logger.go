package logger

import (
	"log"
	"os"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
	Debug(msg string)
	Warn(msg string)
}

type SimpleLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	warnLogger  *log.Logger
}

func NewSimpleLogger() *SimpleLogger {
	return &SimpleLogger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:  log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *SimpleLogger) Info(msg string) {
	l.infoLogger.Println(msg)
}

func (l *SimpleLogger) Error(msg string) {
	l.errorLogger.Println(msg)
}

func (l *SimpleLogger) Debug(msg string) {
	l.debugLogger.Println(msg)
}

func (l *SimpleLogger) Warn(msg string) {
	l.warnLogger.Println(msg)
}
