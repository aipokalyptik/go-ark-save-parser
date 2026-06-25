package arklog

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Level string

const (
	Parser  Level = "parser"
	Info    Level = "info"
	API     Level = "api"
	Error   Level = "error"
	Debug   Level = "debug"
	Warning Level = "warning"
	Save    Level = "save"
	Objects Level = "objects"
	All     Level = "all"
)

type Logger struct {
	mu      sync.RWMutex
	out     io.Writer
	enabled map[Level]bool
}

func New(out io.Writer) *Logger {
	if out == nil {
		out = io.Discard
	}
	return &Logger{
		out: out,
		enabled: map[Level]bool{
			Parser:  false,
			Info:    false,
			API:     false,
			Error:   false,
			Debug:   false,
			Warning: false,
			Save:    false,
			Objects: false,
			All:     false,
		},
	}
}

var Default = New(os.Stdout)

func SetLevel(level Level, enabled bool) {
	Default.SetLevel(level, enabled)
}

func Log(level Level, message string) {
	Default.Log(level, message)
}

func (l *Logger) SetLevel(level Level, enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled[level] = enabled
}

func (l *Logger) Enabled(level Level) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.enabled[All] || l.enabled[level]
}

func (l *Logger) Log(level Level, message string) {
	if !l.Enabled(level) {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintf(l.out, "[%s] %s\n", level, message)
}

func (l *Logger) Parser(message string) {
	l.Log(Parser, message)
}

func (l *Logger) Info(message string) {
	l.Log(Info, message)
}

func (l *Logger) API(message string) {
	l.Log(API, message)
}

func (l *Logger) Error(message string) {
	l.Log(Error, message)
}

func (l *Logger) Debug(message string) {
	l.Log(Debug, message)
}

func (l *Logger) Warning(message string) {
	l.Log(Warning, message)
}

func (l *Logger) Save(message string) {
	l.Log(Save, message)
}

func (l *Logger) Objects(message string) {
	l.Log(Objects, message)
}
