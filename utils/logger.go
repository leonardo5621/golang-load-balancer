package utils

import "go.uber.org/zap"

var Logger *zap.Logger

func InitLogger() *zap.Logger {
	Logger, _ = zap.NewProduction()
	return Logger
}
