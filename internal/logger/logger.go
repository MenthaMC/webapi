package logger

import (
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// Init 初始化日志配置
func Init(level string) {
	// 设置日志格式
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// 设置日志级别
	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

// Debug 输出调试级别日志
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Debugf 输出格式化的调试级别日志
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Info 输出信息级别日志
func Info(args ...interface{}) {
	log.Info(args...)
}

// Infof 输出格式化的信息级别日志
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warn 输出警告级别日志
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Warnf 输出格式化的警告级别日志
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Error 输出错误级别日志
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf 输出格式化的错误级别日志
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatal 输出致命错误日志并退出程序
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Fatalf 输出格式化的致命错误日志并退出程序
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}