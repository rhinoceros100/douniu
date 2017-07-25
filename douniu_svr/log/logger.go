package log

import (
	"fmt"
)

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Error(v ...interface{})
}

func init()  {
	loggerInst = &logger{}
}

var loggerInst Logger

func SetLogger(logger Logger)  {
	loggerInst = logger
}

func Debug(v ...interface{})  {
	loggerInst.Debug(v...)
}

func Info(v ...interface{})  {
	loggerInst.Info(v...)
}

func Error(v ...interface{})  {
	loggerInst.Error(v...)
}

type logger struct {
}

func (l *logger) Debug(v ...interface{})  {
	fmt.Println(v...)
}

func (l *logger) Info(v ...interface{}) {
	fmt.Println(v...)
}

func (l *logger) Error(v ...interface{}) {
	fmt.Println(v...)
}