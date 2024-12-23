package utils

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

const (
	logFilePath       = "logs/app.log"
	logMaxSizeMB      = 10
	logMaxBackups     = 3
	logMaxAgeDays     = 7
	logCompression    = true
	humanReadableTime = "02/01/2006 03:04 PM"
)

func InitLogger() {
	Logger = zap.New(createCore(), zap.AddStacktrace(zapcore.FatalLevel))
}

func createCore() zapcore.Core {
	consoleEncoder := createConsoleEncoder()
	fileEncoder := createFileEncoder()

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    logMaxSizeMB,
		MaxBackups: logMaxBackups,
		MaxAge:     logMaxAgeDays,
		Compress:   logCompression,
	})

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
		zapcore.NewCore(fileEncoder, fileWriter, zapcore.DebugLevel),
	)
}

func createConsoleEncoder() zapcore.Encoder {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(humanReadableTime))
	}
	return zapcore.NewConsoleEncoder(config)
}

func createFileEncoder() zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	return zapcore.NewJSONEncoder(config)
}
