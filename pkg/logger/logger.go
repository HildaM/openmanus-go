package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// InitLogger 初始化日志系统
func InitLogger() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.InfoLevel)
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	return log
}
