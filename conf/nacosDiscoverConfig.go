package conf

import "github.com/spf13/viper"

const (

	//NacosDiscoverNames 当前程序需要调用服务名称
	//
	//goland:noinspection GoUnusedExportedFunction
	NacosDiscoverNames = "nacosdiscover.servicenames"
)

// NacosDiscoverNamesConfig nacos的配置
type NacosDiscoverNamesConfig struct {
	ServiceNames []string `toml:"servicenames"`
}

// GetNacosDiscoverConfig 获取nacos服务发现配置
func GetNacosDiscoverConfig() *NacosDiscoverNamesConfig {

	return &NacosDiscoverNamesConfig{
		ServiceNames: viper.GetStringSlice(NacosDiscoverNames),
	}
}
