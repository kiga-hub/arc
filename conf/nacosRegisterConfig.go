package conf

import (
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"
)

const (
	//NacosRegister nacos服务端字段
	NacosRegister = "nacosRegisterr"
	//NacosRegisterServiceIP 向nacos进行服务注册的服务ip地址
	NacosRegisterServiceIP = NacosRegister + ".ip"
	//NacosRegisterServicePort 向nacos进行服务注册的服务端口
	NacosRegisterServicePort = NacosRegister + ".port"
	//NacosRegisterWeight 注册服务的权重
	NacosRegisterWeight = NacosRegister + ".weight"
	//NacosRegisterClusterName 群组的名字 default:DEFAULT
	NacosRegisterClusterName = NacosRegister + ".clustername"
	//NacosRegisterEnable 是否开启
	NacosRegisterEnable = NacosRegister + ".enable"
	//NacosRegisterHealthy 是否健康
	NacosRegisterHealthy = NacosRegister + ".healthy"
	//NacosRegisterEphemeral Ephemeral 可选
	NacosRegisterEphemeral = NacosRegister + ".ephemeral"
)

var defaultNacosRegisterConfig = vo.RegisterInstanceParam{
	Ip:          "127.0.0.1",
	Port:        80,
	Weight:      10,
	ClusterName: "a",
	Enable:      true,
	Healthy:     true,
	Ephemeral:   true,
}

//SetDefaultNacosRegisterConfig 设置默认基本配置
func SetDefaultNacosRegisterConfig() {
	viper.SetDefault(NacosRegisterServiceIP, defaultNacosRegisterConfig.Weight)
	viper.SetDefault(NacosRegisterServicePort, defaultNacosRegisterConfig.Weight)
	viper.SetDefault(NacosRegisterWeight, defaultNacosRegisterConfig.Weight)
	viper.SetDefault(NacosRegisterClusterName, defaultNacosRegisterConfig.ClusterName)
	viper.SetDefault(NacosRegisterEnable, defaultNacosRegisterConfig.Enable)
	viper.SetDefault(NacosRegisterHealthy, defaultNacosRegisterConfig.Healthy)
	viper.SetDefault(NacosRegisterEphemeral, defaultNacosRegisterConfig.Ephemeral)

}

//GetNacosRegisterConfig 获取基本配置
func GetNacosRegisterConfig() *vo.RegisterInstanceParam {

	return &vo.RegisterInstanceParam{
		Ip:          viper.GetString(NacosRegisterServiceIP),
		Port:        viper.GetUint64(NacosRegisterServicePort),
		Weight:      viper.GetFloat64(NacosRegisterWeight),
		ClusterName: viper.GetString(NacosRegisterClusterName),
		Enable:      viper.GetBool(NacosRegisterEnable),
		Healthy:     viper.GetBool(NacosRegisterHealthy),
		Ephemeral:   viper.GetBool(NacosRegisterEphemeral),
	}
}
