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

//LogGroupConfig  日志配置
type LogGroupConfig struct {
	LogConfig
	Levels map[string][]string `toml:"levels"` // level-[]module
}

//SetDefaultLogGroupConfig 获取默认日志配置
func SetDefaultLogGroupConfig() {
	SetDefaultLogConfig()
	viper.SetDefault(logLevels, defaultLogGroupConfig.Levels)
}

//GetLogGroupConfig  获取日志配置
func GetLogGroupConfig() (*LogGroupConfig, error) {
	config := &LogGroupConfig{
		LogConfig: *GetLogConfig(),
	}
	err := viper.UnmarshalKey(logLevels, &config.Levels)
	return config, err
}
