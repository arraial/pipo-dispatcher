package common

import (
	"sync"

	"go.uber.org/zap"
)

var (
	logger_instance *zap.SugaredLogger
	logger_once     sync.Once
)

// returns logger singleton instance
func GetLogger() *zap.SugaredLogger {
	logger_once.Do(func() {
		logger_instance = initLogger()
	})
	return logger_instance
}

func initLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	return logger.Sugar()
}
