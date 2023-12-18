package redis

import "github.com/spf13/viper"

const (
	redisAddress         = "redis.address"
	redisDB              = "redis.db"
	redisPassword        = "redis.password"
	redisMaxIdle         = "redis.maxIdle"
	redisMaxActive       = "redis.maxActive"
	redisIdleTimeout     = "redis.idleTimeout"
	redisConnTimeout     = "redis.connTimeout"
	redisWait            = "redis.wait"
	redisMaxConnLifetime = "redis.maxConnLifetime"
)

var defaultRedisConfig = Config{
	Address:         "localhost:6379",
	DB:              15,
	MaxIdle:         4,
	MaxActive:       16,
	IdleTimeout:     10,
	ConnTimeout:     10,
	Wait:            true,
	MaxConnLifetime: 3600,
}

// Config redis configuration
type Config struct {
	Address         string `toml:"address"`
	DB              int    `toml:"db"`
	Password        string `toml:"password"`
	MaxIdle         int    `toml:"maxIdle"`
	MaxActive       int    `toml:"maxActive"`
	IdleTimeout     int    `toml:"idleTimeout"`
	ConnTimeout     int    `toml:"connTimeout"`
	Wait            bool   `toml:"wait"`
	MaxConnLifetime int    `toml:"maxConnLifetime"`
}

// SetDefaultRedisConfig  Set the default configuration information for the Redis non-relational cache database
func SetDefaultRedisConfig() {
	viper.SetDefault(redisAddress, defaultRedisConfig.Address)
	viper.SetDefault(redisDB, defaultRedisConfig.DB)
	viper.SetDefault(redisPassword, defaultRedisConfig.Password)
	viper.SetDefault(redisMaxIdle, defaultRedisConfig.MaxIdle)
	viper.SetDefault(redisMaxActive, defaultRedisConfig.MaxActive)
	viper.SetDefault(redisIdleTimeout, defaultRedisConfig.IdleTimeout)
	viper.SetDefault(redisConnTimeout, defaultRedisConfig.ConnTimeout)
	viper.SetDefault(redisWait, defaultRedisConfig.Wait)
	viper.SetDefault(redisMaxConnLifetime, defaultRedisConfig.MaxConnLifetime)
}

// GetRedisConfig Get the configuration information of the Redis non-relational cache database
func GetRedisConfig() *Config {
	return &Config{
		Address:         viper.GetString(redisAddress),
		DB:              viper.GetInt(redisDB),
		Password:        viper.GetString(redisPassword),
		MaxIdle:         viper.GetInt(redisMaxIdle),
		MaxActive:       viper.GetInt(redisMaxActive),
		IdleTimeout:     viper.GetInt(redisIdleTimeout),
		ConnTimeout:     viper.GetInt(redisConnTimeout),
		Wait:            viper.GetBool(redisWait),
		MaxConnLifetime: viper.GetInt(redisMaxConnLifetime),
	}
}
