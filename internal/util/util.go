package util

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SetupLogger sets the global zap instance to our customized version. This
// allows any other class to just call:
// zap.L() or zap.S() to get a Logger or a SugaredLogger (respectively).
func SetupLogger(levelStr string) func() {
	logger := GetLogger(levelStr)

	return zap.ReplaceGlobals(logger)
}

// GetLogger returns a zap.Logger that has colored output for console output.
func GetLogger(levelStr string) *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var level zapcore.Level
	// Determine the level error. If there is one, let's log it
	levelErr := level.Set(levelStr)

	config.Level.SetLevel(level)
	// config.DisableCaller = !level.Enabled(zapcore.DebugLevel)

	// The stack trace is suuuper verbose, it logs on warning or greater
	// config.DisableStacktrace = !level.Enabled(zapcore.DebugLevel)
	config.DisableStacktrace = true
	logger, _ := config.Build()

	if levelErr != nil {
		logger.Fatal("Invalid log-level provided", zapcore.Field{Key: "log-level", String: levelStr})
	}

	return logger
}
