package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New создаёт новый логгер
func New(level, format string) (*zap.Logger, error) {
	// Парсим уровень логирования
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		lvl = zapcore.InfoLevel
	}

	// Выбираем encoder в зависимости от формата
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Создаём core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		lvl,
	)

	// Создаём логгер с caller info
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

// Must создаёт логгер или паникует
func Must(level, format string) *zap.Logger {
	logger, err := New(level, format)
	if err != nil {
		panic(err)
	}
	return logger
}