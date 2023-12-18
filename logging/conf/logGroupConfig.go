package conf

import (
	"github.com/spf13/viper"
)

const (
	logLevels = "log.levels"
)

var defaultLogGroupConfig = LogGroupConfig{
	Levels:    map[string][]string{},
	LogConfig: defaultLogConfig,
}

// LogGroupConfig log group configuration
type LogGroupConfig struct {
	LogConfig
	Levels map[string][]string `toml:"levels"` // level-[]module
}

// SetDefaultLogGroupConfig set default log configuration
func SetDefaultLogGroupConfig() {
	SetDefaultLogConfig()
	viper.SetDefault(logLevels, defaultLogGroupConfig.Levels)
}

// GetLogGroupConfig  get log configuration
func GetLogGroupConfig() (*LogGroupConfig, error) {
	config := &LogGroupConfig{
		LogConfig: *GetLogConfig(),
	}
	err := viper.UnmarshalKey(logLevels, &config.Levels)
	return config, err
}
