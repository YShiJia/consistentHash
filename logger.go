/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-27 00:28:23
 */

package csHash

import (
	"go.uber.org/zap/zapcore"
)

type LoggerLevel zapcore.Level

const (
	DebugLevel = zapcore.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zapcore.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zapcore.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel = zapcore.ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel = zapcore.DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zapcore.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zapcore.FatalLevel
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	DPanic(msg string, fields ...any)
	Panic(msg string, fields ...any)
	Fatal(msg string, fields ...any)
}
