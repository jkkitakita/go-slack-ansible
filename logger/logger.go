package logger

import (
	"go-slack-ansible/config"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Logger outputs logs to LOG_PATH
var Logger = newLogger()

// Shutdown should be called before shutdown app
func Shutdown() error {
	return Logger.Sync()
}

func newLogger() *zap.Logger {
	zapConfig := zap.NewDevelopmentConfig()
	// encoderConfig := zapConfig.EncoderConfig
	encoderConfig := zapConfig.EncoderConfig
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.LevelKey = "severity"
	encoderConfig.NameKey = "receiver"
	encoderConfig.CallerKey = "caller"
	encoderConfig.MessageKey = "message"
	encoderConfig.StacktraceKey = "trace"
	encoderConfig.LineEnding = zapcore.DefaultLineEnding
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	enc := zapcore.NewJSONEncoder(encoderConfig)
	w := zapcore.AddSync(
		&lumberjack.Logger{
			Filename:   viper.GetString(config.LogPath),
			MaxSize:    viper.GetInt(config.LogRotateMaxSize), // MB
			MaxBackups: viper.GetInt(config.LogRotateMaxBackups),
			MaxAge:     viper.GetInt(config.LogRotateMaxDays),
		},
	)
	return zap.New(
		zapcore.NewCore(enc, w, zapConfig.Level),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.WarnLevel),
	)
}
