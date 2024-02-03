package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(loggerLevel string) (*zap.SugaredLogger, error) {

	config := zap.NewProductionConfig()

	level, err := zap.ParseAtomicLevel(loggerLevel)

	if err != nil {
		return nil, err
	}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")

	config.Level = level

	// если оставить, то помимо сообщения выводится stasktrace (для уровня Error).
	// выглядит громоздко, информации в нашем случае не несет. Отключаем.
	config.DisableStacktrace = true

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	logger.Info("Логгер сконфигурирован")
	return sugar, err
}
