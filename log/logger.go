package log

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var appLogger *zap.Logger // Declaring a package-level logger variable

// InitLogger initializes the global Zap logger for the application.
// It sets up the logging level, encoding (development-friendly console format),
// output paths, and replaces the global Zap logger instance so it can be
// accessed from anywhere using zap.L() or zap.S().
func InitLogger(logLevel string, logFilePath string) (*zap.Logger, func(), error) {
	// Define the minimum logging level.
	// We'll use an AtomicLevel to allow dynamic changes at runtime.
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Invalid log level '%s', defaulting to DebugLevel.\n", logLevel)
		level = zap.DebugLevel // Default to DebugLevel for development setup if parsing fails
	}
	atomicLevel := zap.NewAtomicLevelAt(level)

	// Configure the encoder for a development-friendly console format.
	// This uses Zap's DevelopmentEncoderConfig for human-readable output.
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05") // ISO8601 timestamp
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder                  // INFO, WARN, ERROR with colors
	// Now setting EncodeCaller directly for all levels.
	// zapcore.ShortCallerEncoder includes file:line (e.g., main.go:42)
	// zapcore.FullCallerEncoder includes the full path (e.g., /path/to/main.go:42)
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	//encoderCfg.CallerKey = ""
	// Ensure stack trace information is not included
	encoderCfg.StacktraceKey = ""

	// Define where the logs will be written (stdout and an optional file).
	// We use MultiWriteSyncer to write to multiple destinations.
	stdoutSyncer := zapcore.Lock(os.Stdout) // Ensure thread-safe writing to stdout

	var writeSyncer zapcore.WriteSyncer
	fileSyncer, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file 'app.log': %v. Logging only to stdout.\n", err)
		writeSyncer = stdoutSyncer
	} else {
		fileWriteSyncer := zapcore.AddSync(fileSyncer) // Wrap the file for Zap's WriteSyncer
		writeSyncer = fileWriteSyncer                  // zapcore.NewMultiWriteSyncer(stdoutSyncer, fileWriteSyncer)
	}

	// Create a core that combines the development encoder, write syncer, and atomic level.
	core := zapcore.NewCore(
		//zapcore.NewJSONEncoder(encoderCfg),
		zapcore.NewConsoleEncoder(encoderCfg), // Use console encoder for development style
		writeSyncer,
		atomicLevel,
	)

	// Build the logger.
	// AddCaller() includes file and line number.
	// AddStacktrace() captures stack traces for ErrorLevel and above (default for development).
	appLogger = zap.New(core, zap.AddCaller())
	closer := func() {
		if appLogger != nil {
			appLogger.Info("Flushing global Zap logger buffers...")
			if err := appLogger.Sync(); err != nil {
				// Sync can return an error if the underlying WriteSyncer fails.
				// This often happens if stdout/stderr are closed prematurely.
				fmt.Fprintf(os.Stderr, "Error syncing logger: %v\n", err)
			}
		}
		if fileSyncer != nil {
			fileSyncer.Close()
		}
	}
	return appLogger, closer, nil
}

// CloseLogger ensures that any buffered log entries are flushed before the application exits.
// It should be called using defer in your main function.
func CloseLogger() {
	if appLogger != nil {
		appLogger.Info("Flushing global Zap logger buffers...")
		if err := appLogger.Sync(); err != nil {
			// Sync can return an error if the underlying WriteSyncer fails.
			// This often happens if stdout/stderr are closed prematurely.
			fmt.Fprintf(os.Stderr, "Error syncing logger: %v\n", err)
		}
	}
}

// SetLogLevel dynamically changes the global logger's minimum level at runtime.
func SetLogLevel(newLevel string) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(newLevel)); err != nil {
		zap.L().Warn("Invalid new log level, not changing", zap.String("requested_level", newLevel), zap.Error(err))
		return
	}
	zap.L().Info("Changing global log level", zap.String("old_level", zap.L().Level().String()), zap.String("new_level", level.String()))
	// To change the level of a logger built with a core that uses AtomicLevel,
	// you need to access the AtomicLevel instance directly.
	// Assuming the core was created with an AtomicLevel, you can often cast it.
	if al, ok := zap.L().Core().(interface{ SetLevel(zapcore.Level) }); ok {
		al.SetLevel(level)
	} else {
		zap.L().Error("Cannot dynamically change log level: Logger core does not support SetLevel.")
	}
}

// Expose the raw logger for specific needs if required (though zap.L() is preferred)
func GetLogger() *zap.Logger {
	return zap.L()
}

// Expose the sugared logger for convenience
func GetSugaredLogger() *zap.SugaredLogger {
	return zap.S()
}
