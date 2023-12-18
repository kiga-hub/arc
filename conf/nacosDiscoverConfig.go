package conf

import "github.com/spf13/viper"

const (

	//NacosDiscoverNames The current program needs to call the service name
	//
	//goland:noinspection GoUnusedExportedFunction
	NacosDiscoverNames = "nacosdiscover.servicenames"
)

// NacosDiscoverNamesConfig nacos configuration
type NacosDiscoverNamesConfig struct {
	ServiceNames []string `toml:"servicenames"`
}

// GetNacosDiscoverConfig get nacos service discovery configuration
func GetNacosDiscoverConfig() *NacosDiscoverNamesConfig {

	return &NacosDiscoverNamesConfig{
		ServiceNames: viper.GetStringSlice(NacosDiscoverNames),
	}
}
