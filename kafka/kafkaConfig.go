package kafka

import "github.com/spf13/viper"

const (
	bootstrapServeres = "kafka.bootstrapserveres"
	clientID          = "kafka.clientid"
	topic             = "kafka.topic"
	messagemaxbytes   = "kafka.messagemaxbytes"
	kafkaenable       = "kafka.enable"
	groupID           = "kafka.groupid"
)

var defaultConfig = Config{
	Enable:           false,
	BootStrapServers: "localhost:9092",
	ClientID:         "arc",
	MessageMaxBytes:  67108864,
	Topic:            "arc",
	GroupID:          "arc",
}

// Config struct
type Config struct {
	Enable           bool   `toml:"enable"`
	BootStrapServers string `toml:"bootstrapserveres"` // BootStrapServers Broker
	ClientID         string `toml:"clientid"`          // GroupID for producer
	MessageMaxBytes  int    `toml:"messagemaxbytes"`   // MessageMaxBytes
	Topic            string `toml:"topic"`
	GroupID          string `toml:"groupid"`
}

//SetDefaultConfig -
func SetDefaultConfig() {
	viper.SetDefault(kafkaenable, defaultConfig.Enable)
	viper.SetDefault(bootstrapServeres, defaultConfig.BootStrapServers)
	viper.SetDefault(clientID, defaultConfig.ClientID)
	viper.SetDefault(messagemaxbytes, defaultConfig.MessageMaxBytes)
	viper.SetDefault(topic, defaultConfig.Topic)
	viper.SetDefault(groupID, defaultConfig.GroupID)
}

//GetConfig -
func GetConfig() *Config {
	return &Config{
		Enable:           viper.GetBool(kafkaenable),
		BootStrapServers: viper.GetString(bootstrapServeres),
		ClientID:         viper.GetString(clientID),
		MessageMaxBytes:  viper.GetInt(messagemaxbytes),
		Topic:            viper.GetString(topic),
		GroupID:          viper.GetString(groupID),
	}
}
