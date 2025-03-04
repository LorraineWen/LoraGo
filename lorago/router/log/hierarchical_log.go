package log

import (
	"fmt"
	"io"
	"os"
	"time"
)

/*
*@Author: LorraineWen
*支持不同颜色的日志
*支持日志格式自定义
*支持分级日志，比如error级别的日志，info级别的日志，debug级别的日志
*debug级别下，三种打印都可以生效，info级别下，只有debug和info有效，error级别下只有error有效
 */
//定义日志级别
type LoggerLevel int

const (
	LevelDebug LoggerLevel = iota
	LevelInfo
	LevelError
)

type Logger struct {
	Formatter HierarchicalLogFormatter
	Outs      []io.Writer
	Level     LoggerLevel
}

func New() *Logger {
	return &Logger{}
}

type HierarchicalLogFormatter struct {
	Color bool
	Level LoggerLevel
}

func NewLogger() *Logger {
	logger := New()
	out := os.Stdout
	logger.Outs = append(logger.Outs, out)
	logger.Level = LevelDebug
	logger.Formatter = HierarchicalLogFormatter{}
	return logger
}

func (l *Logger) Info(msg any) {
	l.Print(LevelInfo, msg)
}

func (l *Logger) Debug(msg any) {
	l.Print(LevelDebug, msg)
}

func (l *Logger) Error(msg any) {
	l.Print(LevelError, msg)
}
func (l *Logger) Print(level LoggerLevel, msg any) {
	if l.Level > level {
		//如果路由传递进来的级别大于Info等函数自定义的级别，那么对应函数就不打印
		return
	}
	l.Formatter.Level = level
	formatter := l.Formatter.formatter(msg)
	for _, out := range l.Outs {
		fmt.Fprint(out, formatter)
	}
}
func (f *HierarchicalLogFormatter) formatter(msg any) string {
	now := time.Now()
	return fmt.Sprintf("[msgo] %v | level=%s | msg=%#v \n",
		now.Format("2006/01/02 - 15:04:05"),
		f.Level.Level(), msg,
	)
}

func (level LoggerLevel) Level() string {
	switch level {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}
