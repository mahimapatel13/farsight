package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger that provides structured logging.
type Logger struct {
	zapLogger *zap.Logger
	sugar     *zap.SugaredLogger
	level     zap.AtomicLevel // Added field to store the level
}

// NewLogger creates a new logger instance.
// By default, it writes to stdout and includes timestamps, log levels, and caller information.
func NewLogger() *Logger {
	// Create a new development encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create the atomic level and store it
	level := zap.NewAtomicLevelAt(zapcore.InfoLevel)

	// Create a console encoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// Create a core that writes to stdout
	core := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		level, // Use the atomic level reference
	)

	// Create a logger with the core and add caller info
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{
		zapLogger: zapLogger,
		sugar:     zapLogger.Sugar(),
		level:     level, // Store the level reference
	}
}

// WithField adds a field to the logger.
func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{
		zapLogger: l.zapLogger.With(zap.Any(key, value)),
		sugar:     l.zapLogger.Sugar().With(key, value),
		level:     l.level, // Maintain the level reference
	}
}

// Debug logs a message at debug level with optional key-value pairs.
func (l *Logger) Debug(msg string, keysAndValues ...any) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Info logs a message at info level with optional key-value pairs.
func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Warn logs a message at warn level with optional key-value pairs.
func (l *Logger) Warn(msg string, keysAndValues ...any) {
	l.sugar.Warnw(msg, keysAndValues...)
}

// Error logs a message at error level with optional key-value pairs.
func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.sugar.Errorw(msg, keysAndValues...)
}

// Fatal logs a message at fatal level with optional key-value pairs and then exits.
func (l *Logger) Fatal(msg string, keysAndValues ...any) {
	l.sugar.Fatalw(msg, keysAndValues...)
}

// SetLevel sets the logging level.
func (l *Logger) SetLevel(level string) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "fatal":
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	l.level.SetLevel(zapLevel) // Now we can actually set the level
}

// GetLevel returns the current logging level as a string.
func (l *Logger) GetLevel() string {
	level := l.level.Level()
	switch level {
	case zapcore.DebugLevel:
		return "debug"
	case zapcore.InfoLevel:
		return "info"
	case zapcore.WarnLevel:
		return "warn"
	case zapcore.ErrorLevel:
		return "error"
	case zapcore.FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

// WithError adds an error field to the logger.
func (l *Logger) WithError(err error) *Logger {
	return l.WithField("error", err.Error())
}

