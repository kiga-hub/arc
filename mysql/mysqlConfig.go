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

// defaultMysqlConfig 默认mysql数据库配置
var defaultMysqlConfig = Config{
	Host:         "localhost",
	Port:         3306,
	DB:           "kiga",
	User:         "root",
	LogMode:      false,
	MaxOpenConns: 16,
	MaxIdleConns: 4,
}

// Config  mysql数据库配置
type Config struct {
	Host         string `toml:"host"`     // 服务器地址
	Port         int    `toml:"port"`     // 端口, 例如3306
	DB           string `toml:"db"`       // 数据库名
	User         string `toml:"user"`     // 用户名
	Password     string `toml:"password"` // 密码
	TimeZone     string `toml:"timeZone"` // 时区
	LogMode      bool   `toml:"logmode"`
	MaxOpenConns int    `toml:"maxOpenConns"` // 连接池信息
	MaxIdleConns int    `toml:"maxIdleConns"`
}

// SetDefaultMysqlConfig 设置mysql数据库默认配置
func SetDefaultMysqlConfig() {
	viper.SetDefault(mysqlHost, defaultMysqlConfig.Host)
	viper.SetDefault(mysqlPort, defaultMysqlConfig.Port)
	viper.SetDefault(mysqlDB, defaultMysqlConfig.DB)
	viper.SetDefault(mysqlUser, defaultMysqlConfig.User)
	viper.SetDefault(mysqlLogMode, defaultMysqlConfig.LogMode)
	viper.SetDefault(mysqlMaxOpenConns, defaultMysqlConfig.MaxOpenConns)
	viper.SetDefault(mysqlMaxIdleConns, defaultMysqlConfig.MaxIdleConns)
}

// GetMysqlConfig 获取mysql数据库配置
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
