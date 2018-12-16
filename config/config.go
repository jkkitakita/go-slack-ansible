package config

import "github.com/spf13/viper"

// These are used for getting config value
const (
	IntervalSeconds     = "INTERVAL_SECONDS"
	ShutdownWaitSeconds = "SHUTDOWN_WAIT_SECONDS"
	LogPath             = "LOG_PATH"
	LogRotateMaxSize    = "LOG_ROTATE_MAX_SIZE"
	LogRotateMaxBackups = "LOG_ROTATE_MAX_BACKUPS"
	LogRotateMaxDays    = "LOG_ROTATE_MAX_DAYS"
)

func init() {
	viper.SetDefault(IntervalSeconds, 1)     // seconds
	viper.SetDefault(ShutdownWaitSeconds, 1) // seconds
	viper.SetDefault(LogPath, "./log/app.log")
	viper.SetDefault(LogRotateMaxSize, 100) // MB
	viper.SetDefault(LogRotateMaxBackups, 3)
	viper.SetDefault(LogRotateMaxDays, 7)
	viper.SetEnvPrefix("SAMPLE")
	viper.AutomaticEnv() // e.g. SAMPLE_LOG_PATH is binded to LOG_PATH
}
