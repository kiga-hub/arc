package alertmanager

import "github.com/spf13/viper"

const (
	enable        = "alertmanager.enable"
	serveraddress = "alertmanager.serveraddress"
	serverport    = "alertmanager.serverport"
	url           = "alertmanager.url"
)

var defaultConfig = Config{
	Enable:        false,
	ServerAddress: "localhost",
	ServerPort:    7001,
	URL:           "/alertmanager/webhook",
}

// Config -
type Config struct {
	Enable        bool   `toml:"enable"`
	ServerAddress string `toml:"serveraddress"`
	ServerPort    int    `toml:"serverport"`
	URL           string `toml:"url"`
}

// SetDefaultConfig -
func SetDefaultConfig() {
	viper.SetDefault(enable, defaultConfig.Enable)
	viper.SetDefault(serveraddress, defaultConfig.ServerAddress)
	viper.SetDefault(serverport, defaultConfig.ServerPort)
	viper.SetDefault(url, defaultConfig.URL)
}

// GetConfig -
func GetConfig() *Config {
	return &Config{
		Enable:        viper.GetBool(enable),
		ServerAddress: viper.GetString(serveraddress),
		ServerPort:    viper.GetInt(serverport),
		URL:           viper.GetString(url),
	}
}
