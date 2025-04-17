package fadingMetricsCache

import (
	"fmt"
	"log/slog"
	"os"
)

var logger slog.Logger

func initLog(envVarName string) {
	switch os.Getenv(envVarName) {
	case "ERROR":
		slog.SetLogLoggerLevel(slog.LevelError)
	case "WARN":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "INFO":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "DEBUG":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	default:
	}
	logger = *slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func logFatalf(format string, args ...any) {
	logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func logErrorf(format string, args ...any) {
	logger.Error(fmt.Sprintf(format, args...))
}

func logWarnf(format string, args ...any) {
	logger.Warn(fmt.Sprintf(format, args...))
}

func logInfof(format string, args ...any) {
	logger.Info(fmt.Sprintf(format, args...))
}

func logDebugf(format string, args ...any) {
	slog.Debug(fmt.Sprintf(format, args...))
}
