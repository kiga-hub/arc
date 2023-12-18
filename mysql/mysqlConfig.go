package mysql

import "github.com/spf13/viper"

const (
	mysqlHost         = "mysql.host"
	mysqlPort         = "mysql.port"
	mysqlDB           = "mysql.db"
	mysqlUser         = "mysql.user"
	mysqlPassword     = "mysql.password"
	mysqlTimeZone     = "mysql.timezone"
	mysqlLogMode      = "mysql.logmode"
	mysqlMaxOpenConns = "mysql.maxopenconns"
	mysqlMaxIdleConns = "mysql.maxidleconns"
)

// defaultMysqlConfig  mysql default configuration
var defaultMysqlConfig = Config{
	Host:         "localhost",
	Port:         3306,
	DB:           "kiga",
	User:         "root",
	LogMode:      false,
	MaxOpenConns: 16,
	MaxIdleConns: 4,
}

// Config  mysql configuration
type Config struct {
	Host         string `toml:"host"`
	Port         int    `toml:"port"`
	DB           string `toml:"db"`
	User         string `toml:"user"`
	Password     string `toml:"password"`
	TimeZone     string `toml:"timeZone"`
	LogMode      bool   `toml:"logmode"`
	MaxOpenConns int    `toml:"maxOpenConns"`
	MaxIdleConns int    `toml:"maxIdleConns"`
}

// SetDefaultMysqlConfig set default mysql configuration
func SetDefaultMysqlConfig() {
	viper.SetDefault(mysqlHost, defaultMysqlConfig.Host)
	viper.SetDefault(mysqlPort, defaultMysqlConfig.Port)
	viper.SetDefault(mysqlDB, defaultMysqlConfig.DB)
	viper.SetDefault(mysqlUser, defaultMysqlConfig.User)
	viper.SetDefault(mysqlLogMode, defaultMysqlConfig.LogMode)
	viper.SetDefault(mysqlMaxOpenConns, defaultMysqlConfig.MaxOpenConns)
	viper.SetDefault(mysqlMaxIdleConns, defaultMysqlConfig.MaxIdleConns)
}

// GetMysqlConfig get mysql configuration
func GetMysqlConfig() *Config {
	return &Config{
		Host:         viper.GetString(mysqlHost),
		Port:         viper.GetInt(mysqlPort),
		DB:           viper.GetString(mysqlDB),
		User:         viper.GetString(mysqlUser),
		Password:     viper.GetString(mysqlPassword),
		TimeZone:     viper.GetString(mysqlTimeZone),
		LogMode:      viper.GetBool(mysqlLogMode),
		MaxOpenConns: viper.GetInt(mysqlMaxOpenConns),
		MaxIdleConns: viper.GetInt(mysqlMaxIdleConns),
	}
}
