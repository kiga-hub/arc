package configuration

import (
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/spf13/viper"
)

const (
	//nacosServerPrefix nacos service field
	nacosServerPrefix = "nacosServer"
	//nacosServerIPAddr nacos service ip address
	nacosServerIPAddr = nacosServerPrefix + ".ipaddr"
	//nacosServerContextPath  nacos The request path of the service
	nacosServerContextPath = nacosServerPrefix + ".contextPath"
	//nacosServerPort nacos service port
	nacosServerPort = nacosServerPrefix + ".port"
)

//NacosServerConfig nacos service configuration
// type NacosServerConfig struct {
// 	IPAddr      string `toml:"ipaddr"`
// 	ContextPath string `toml:"contextPath"`
// 	Port        uint64 `toml:"port"`
// }

var defaultNacosServerConfig = constant.ServerConfig{
	ContextPath: "/nacos",
	IpAddr:      "127.0.0.1",
	Port:        8848,
}

// SetDefaultNacosServerConfig set default basic configuration
func SetDefaultNacosServerConfig() {
	viper.SetDefault(nacosServerIPAddr, defaultNacosServerConfig.IpAddr)
	viper.SetDefault(nacosServerContextPath, defaultNacosServerConfig.ContextPath)
	viper.SetDefault(nacosServerPort, defaultNacosServerConfig.Port)

}

// GetNacosServerConfig get basic configuration
func GetNacosServerConfig() *constant.ServerConfig {
	return &constant.ServerConfig{
		IpAddr:      viper.GetString(nacosServerIPAddr),
		ContextPath: viper.GetString(nacosServerContextPath),
		Port:        viper.GetUint64(nacosServerPort),
	}
}
