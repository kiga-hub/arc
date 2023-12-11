package pulsar

import "github.com/spf13/viper"

const (
	pulsarEnable = "pulsar.enable"
	pulsarURL    = "pulsar.url"
	pulsarTopic  = "pulsar.topic"
)

var defaultConfig = Config{
	Enable: false,
	URL:    "pulsar://localhost:6650",
	Topic:  "persistent://public/default/kiga.data",
}

// Config .
type Config struct {
	Enable bool   `toml:"enable"`
	URL    string `toml:"url"`
	Topic  string `toml:"topic"`
}

// SetDefaultConfig -
func SetDefaultConfig() {
	viper.SetDefault(pulsarEnable, defaultConfig.Enable)
	viper.SetDefault(pulsarURL, defaultConfig.URL)
	viper.SetDefault(pulsarTopic, defaultConfig.Topic)
}

// GetConfig -
func GetConfig() *Config {
	return &Config{
		Enable: viper.GetBool(pulsarEnable),
		URL:    viper.GetString(pulsarURL),
		Topic:  viper.GetString(pulsarTopic),
	}
}
