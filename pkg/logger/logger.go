package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	mu  sync.Mutex
	out io.Writer
}

func New() *Logger {
	return &Logger{out: os.Stdout}
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log("WARN", format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log("DEBUG", format, args...)
}

func (l *Logger) log(level, format string, args ...interface{}) {
	var b strings.Builder
	b.WriteString(time.Now().Format(time.RFC3339))
	b.WriteString(" ")
	b.WriteString(level)
	b.WriteString(" ")
	b.WriteString(fmt.Sprintf(format, args...))
	b.WriteString("\n")

	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = io.WriteString(l.out, b.String())
}
