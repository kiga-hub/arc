package conf

import "github.com/spf13/viper"

const (
	poolMin         = "pool.min"
	poolMax         = "pool.max"
	poolIdleTimeout = "pool.idleTimeout"
)

var defaultPoolConfig = PoolConfig{
	MinActive:   1,
	MaxActive:   1,
	IdleTimeout: 86400,
}

//PoolConfig   连接池配置
type PoolConfig struct {
	MinActive   int `toml:"min"`         //连接池中拥有的最小连接数
	MaxActive   int `toml:"max"`         //连接池中拥有的最大的连接数
	IdleTimeout int `toml:"idleTimeout"` //连接最大空闲时间，超过该事件则将失效
}

//SetDefaultPoolConfig 设置默认连接池配置
func SetDefaultPoolConfig() {
	viper.SetDefault(poolMin, defaultPoolConfig.MinActive)
	viper.SetDefault(poolMax, defaultPoolConfig.MaxActive)
	viper.SetDefault(poolIdleTimeout, defaultPoolConfig.IdleTimeout)
}

//GetPoolConfig 获取连接池配置
func GetPoolConfig() *PoolConfig {
	return &PoolConfig{
		MinActive:   viper.GetInt(poolMin),
		MaxActive:   viper.GetInt(poolMax),
		IdleTimeout: viper.GetInt(poolIdleTimeout),
	}
}
