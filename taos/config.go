package taos

import "github.com/spf13/viper"

const (
	taosenable    = "taos.enable"
	taosdevice    = "taos.device"
	taoshost      = "taos.host"
	taosport      = "taos.port"
	taosuser      = "taos.user"
	taospassword  = "taos.password"
	taosname      = "taos.name"
	taospoolsize  = "taos.poolsize"
	taosprecision = "taos.precision"
)

var defaultTaosConfig = Config{
	Enable:    true,
	Device:    "taosSql",
	Host:      "localhost",
	Port:      6030,
	User:      "root",
	Password:  "taosdata",
	Name:      "kiga",
	PoolSize:  64,
	Precision: "ms",
}

// Config struct
type Config struct {
	Enable    bool   `toml:"enable"`
	Device    string `toml:"device"`
	Host      string `toml:"host"`
	Port      int    `toml:"port"`
	User      string `toml:"user"`
	Password  string `toml:"password"`
	Name      string `toml:"name"`
	PoolSize  uint   `toml:"poolsize"`
	Precision string `toml:"precision"`
}

// SetDefaultConfig -
func SetDefaultConfig() {
	viper.SetDefault(taosenable, defaultTaosConfig.Enable)
	viper.SetDefault(taosdevice, defaultTaosConfig.Device)
	viper.SetDefault(taoshost, defaultTaosConfig.Host)
	viper.SetDefault(taosport, defaultTaosConfig.Port)
	viper.SetDefault(taosuser, defaultTaosConfig.User)
	viper.SetDefault(taospassword, defaultTaosConfig.Password)
	viper.SetDefault(taosname, defaultTaosConfig.Name)
	viper.SetDefault(taospoolsize, defaultTaosConfig.PoolSize)
	viper.SetDefault(taosprecision, defaultTaosConfig.Precision)
}

// GetConfig -
func GetConfig() *Config {
	return &Config{
		Enable:    viper.GetBool(taosenable),
		Device:    viper.GetString(taosdevice),
		Host:      viper.GetString(taoshost),
		Port:      viper.GetInt(taosport),
		User:      viper.GetString(taosuser),
		Password:  viper.GetString(taospassword),
		Name:      viper.GetString(taosname),
		PoolSize:  viper.GetUint(taospoolsize),
		Precision: viper.GetString(taosprecision),
	}
}
