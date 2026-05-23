//go:build wasip1

package client

import (
	"fmt"
	"os"
	"strings"
)

// Logger writes structured log messages to the host.
type Logger struct{}

// NewLogger returns a new Logger.
func NewLogger() *Logger {
	return &Logger{}
}

// Debug writes a debug log message.
func (l *Logger) Debug(msg string, kvs ...string) {
	fmt.Fprintln(os.Stderr, l.format("debug", msg, kvs))
}

// Info writes an info log message.
func (l *Logger) Info(msg string, kvs ...string) {
	fmt.Fprintln(os.Stderr, l.format("info", msg, kvs))
}

// Warn writes a warn log message.
func (l *Logger) Warn(msg string, kvs ...string) {
	fmt.Fprintln(os.Stderr, l.format("warn", msg, kvs))
}

// Error writes an error log message.
func (l *Logger) Error(msg string, kvs ...string) {
	fmt.Fprintln(os.Stderr, l.format("error", msg, kvs))
}

func (l *Logger) format(level, msg string, kvs []string) string {
	var b strings.Builder
	b.WriteString("level=")
	b.WriteString(level)
	b.WriteString(" msg=")
	_, _ = fmt.Fprintf(&b, "%q", msg)
	for i := 0; i+1 < len(kvs); i += 2 {
		b.WriteByte(' ')
		b.WriteString(kvs[i])
		b.WriteByte('=')
		val := kvs[i+1]
		if strings.ContainsAny(val, " \t\r\n\"") {
			_, _ = fmt.Fprintf(&b, "%q", val)
		} else {
			b.WriteString(val)
		}
	}
	return b.String()
}
