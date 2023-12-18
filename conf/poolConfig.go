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

// PoolConfig connection pool configuration
type PoolConfig struct {
	MinActive   int `toml:"min"`         // The minimum number of connections in the connection pool
	MaxActive   int `toml:"max"`         // The maximum number of connections in the connection pool
	IdleTimeout int `toml:"idleTimeout"` // Maximum idle time of connection.it will become invalid after this time.
}

// SetDefaultPoolConfig set default connection pool configuration
func SetDefaultPoolConfig() {
	viper.SetDefault(poolMin, defaultPoolConfig.MinActive)
	viper.SetDefault(poolMax, defaultPoolConfig.MaxActive)
	viper.SetDefault(poolIdleTimeout, defaultPoolConfig.IdleTimeout)
}

// GetPoolConfig get connection pool configuration
func GetPoolConfig() *PoolConfig {
	return &PoolConfig{
		MinActive:   viper.GetInt(poolMin),
		MaxActive:   viper.GetInt(poolMax),
		IdleTimeout: viper.GetInt(poolIdleTimeout),
	}
}
