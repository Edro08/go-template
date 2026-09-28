package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"
)

// ILogger
// ------------------------------------------------------------------------------------------------
type ILogger interface {
	Debug(title string, keys ...any)
	Info(title string, keys ...any)
	Warn(title string, keys ...any)
	Error(title string, keys ...any)
	Fatal(title string, keys ...any)
	Slog() *slog.Logger
	With(args ...any) *Logger
	WithGroup(name string) *Logger
	Close() error
}

// ------------------------------------------------------------------------------------------------
// Implementation Methods
// ------------------------------------------------------------------------------------------------

func (l *Logger) Debug(title string, keys ...any) {
	l.write(DEBUG, title, keys...)
}

func (l *Logger) Info(title string, keys ...any) {
	l.write(INFO, title, keys...)
}

func (l *Logger) Warn(title string, keys ...any) {
	l.write(WARN, title, keys...)
}

func (l *Logger) Error(title string, keys ...any) {
	l.write(ERROR, title, keys...)
}

func (l *Logger) Fatal(title string, keys ...any) {
	l.write(FATAL, title, keys...)
	os.Exit(1)
}

// Slog retorna la instancia interna de *slog.Logger de la librería estándar.
func (l *Logger) Slog() *slog.Logger {
	return l.slog
}

// With retorna un nuevo Logger derivado que incluye atributos contextuales fijos.
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		opts:   l.opts,
		slog:   l.slog.With(args...),
		closer: l.closer,
	}
}

// WithGroup retorna un nuevo Logger derivado que agrupa los atributos bajo un grupo.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{
		opts:   l.opts,
		slog:   l.slog.WithGroup(name),
		closer: l.closer,
	}
}

// ToSlog convierte el Level interno al Level de log/slog.
func (l Level) toSlog() slog.Level {
	switch l {
	case DEBUG:
		return slog.LevelDebug
	case INFO:
		return slog.LevelInfo
	case WARN:
		return slog.LevelWarn
	case ERROR:
		return slog.LevelError
	case FATAL:
		return slog.LevelError + 4
	default:
		return slog.LevelInfo
	}
}

// write delega la creación del registro a log/slog preservando la ubicación exacta de la llamada original.
func (l *Logger) write(level Level, title string, keysVals ...any) {
	slogLevel := level.toSlog()
	ctx := context.Background()

	//	if !l.slog.Enabled(ctx, slogLevel) {
	//		return
	//	}

	var pcs [1]uintptr
	// 3 profundidad: [0] runtime.Callers, [1] l.write, [2] l.Info/Debug/etc, [3] código del usuario
	runtime.Callers(3, pcs[:])

	record := slog.NewRecord(time.Now(), slogLevel, title, pcs[0])
	record.Add(keysVals...)

	_ = l.slog.Handler().Handle(ctx, record)
}
