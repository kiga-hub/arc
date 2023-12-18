package mongodb

import "github.com/spf13/viper"

const (
	mongoAddress  = "mongo.address"
	mongoDB       = "mongo.db"
	mongoUser     = "mongo.user"
	mongoPassword = "mongo.password"
	mongoOptions  = "mongo.options"
)

// Config  database configuration
type Config struct {
	Address  string `toml:"host"`     // Server address
	DB       string `toml:"db"`       // database name
	User     string `toml:"user"`     // user name
	Password string `toml:"password"` // password
	Options  string `toml:"options"`  // config
}

// GetMongoConfig get database configuration
func GetMongoConfig() *Config {
	return &Config{
		Address:  viper.GetString(mongoAddress),
		DB:       viper.GetString(mongoDB),
		User:     viper.GetString(mongoUser),
		Password: viper.GetString(mongoPassword),
		Options:  viper.GetString(mongoOptions),
	}
}
