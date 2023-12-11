package mongodb

import "github.com/spf13/viper"

const (
	mongoAddress  = "mongo.address"
	mongoDB       = "mongo.db"
	mongoUser     = "mongo.user"
	mongoPassword = "mongo.password"
	mongoOptions  = "mongo.options"
)

//Config  数据库配置
type Config struct {
	Address  string `toml:"host"`     // 服务器地址
	DB       string `toml:"db"`       // 数据库名
	User     string `toml:"user"`     // 用户名
	Password string `toml:"password"` // 密码
	Options  string `toml:"options"`  // 配置
}

//GetMongoConfig 获取数据库配置
func GetMongoConfig() *Config {
	return &Config{
		Address:  viper.GetString(mongoAddress),
		DB:       viper.GetString(mongoDB),
		User:     viper.GetString(mongoUser),
		Password: viper.GetString(mongoPassword),
		Options:  viper.GetString(mongoOptions),
	}
}
