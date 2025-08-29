package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// Logger 全局日志器
var Logger *zap.Logger

// InitLogger 显式初始化日志器
func InitLogger() {
	// 控制台输出编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "file",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 带颜色
		EncodeTime:     zapcore.ISO8601TimeEncoder,       // ISO8601格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短文件路径
	}

	// 使用控制台编码器
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 创建核心 - 仅输出到控制台
	core := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout), // 只输出到控制台
		zap.InfoLevel,              // info 的级别
	)

	// 创建日志器
	Logger = zap.New(core,
		zap.AddCaller(),                   // 添加调用者信息
		zap.AddCallerSkip(1),              // 跳过封装函数
		zap.AddStacktrace(zap.ErrorLevel), // 错误级别添加堆栈
	)

	// 重定向标准log库输出到zap
	zap.RedirectStdLog(Logger)
}

// 快捷日志方法 =======================================

func LogDebug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func LogInfo(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func LogWarn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func LogError(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func LogFatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// 上下文字段快捷方法 =================================

func Str(key, value string) zap.Field {
	return zap.String(key, value)
}

func Strings(key string, value []string) zap.Field {
	return zap.Strings(key, value)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func Err(err error) zap.Field {
	return zap.Error(err)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

func Float64s(key string, values []float64) zap.Field {
	return zap.Float64s(key, values)
}

// SyncLogger 安全关闭日志器
func SyncLogger() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
