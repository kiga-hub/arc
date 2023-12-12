package configuration

import (
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/spf13/viper"
)

const (
	//nacosServerPrefix nacos服务端字段
	nacosServerPrefix = "nacosServer"
	//nacosServerIPAddr nacos服务的ip地址
	nacosServerIPAddr = nacosServerPrefix + ".ipaddr"
	//nacosServerContextPath  nacos服务的请求路径
	nacosServerContextPath = nacosServerPrefix + ".contextPath"
	//nacosServerPort nacos服务的端口
	nacosServerPort = nacosServerPrefix + ".port"
)

//NacosServerConfig nacos服务的配置
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

// SetDefaultNacosServerConfig 设置默认nacos服务的配置
func SetDefaultNacosServerConfig() {
	viper.SetDefault(nacosServerIPAddr, defaultNacosServerConfig.IpAddr)
	viper.SetDefault(nacosServerContextPath, defaultNacosServerConfig.ContextPath)
	viper.SetDefault(nacosServerPort, defaultNacosServerConfig.Port)

}

// GetNacosServerConfig 获取nacos服务的基本配置
func GetNacosServerConfig() *constant.ServerConfig {
	return &constant.ServerConfig{
		IpAddr:      viper.GetString(nacosServerIPAddr),
		ContextPath: viper.GetString(nacosServerContextPath),
		Port:        viper.GetUint64(nacosServerPort),
	}
}
