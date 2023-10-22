//go:build js && wasm

package client

import (
	"os" //nolint:gci
	"syscall/js"
)

// Logger binds to the looking glass logger in the browser.
type Logger struct {
	name string
}

// NewLogger returns a new logger.
func NewLogger() *Logger {
	name := os.Args[0]

	return &Logger{
		name: name,
	}
}

// Debug writes a debug log message.
func (l *Logger) Debug(msg string, kvs ...string) {
	js.Global().Call("log.debug", msg, toSliceAny(kvs))
}

// Info writes a info log message.
func (l *Logger) Info(msg string, kvs ...string) {
	js.Global().Call("log.info", msg, toSliceAny(kvs))
}

// Error writes a error log message.
func (l *Logger) Error(msg string, kvs ...string) {
	js.Global().Call("log.error", msg, toSliceAny(kvs))
}

func toSliceAny[T any](a []T) []any {
	arr := make([]any, len(a))
	for i, s := range a {
		arr[i] = s
	}
	return arr
}
