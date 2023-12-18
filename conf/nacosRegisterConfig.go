package conf

import (
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"
)

const (
	//NacosRegister nacos service field
	NacosRegister = "nacosRegisterr"
	//NacosRegisterServiceIP The ip address of the service to be registered with nacos.
	NacosRegisterServiceIP = NacosRegister + ".ip"
	//NacosRegisterServicePort The service port for registering with nacos.
	NacosRegisterServicePort = NacosRegister + ".port"
	//NacosRegisterWeight The weight of the registered service.
	NacosRegisterWeight = NacosRegister + ".weight"
	//NacosRegisterClusterName The name of group. default:DEFAULT
	NacosRegisterClusterName = NacosRegister + ".clusterName"
	//NacosRegisterEnable is it enabled
	NacosRegisterEnable = NacosRegister + ".enable"
	//NacosRegisterHealthy is it healthy
	NacosRegisterHealthy = NacosRegister + ".healthy"
	//NacosRegisterEphemeral Ephemeral optioanl
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

// SetDefaultNacosRegisterConfig set default basic configuration
//
//goland:noinspection GoUnusedExportedFunction
func SetDefaultNacosRegisterConfig() {
	viper.SetDefault(NacosRegisterServiceIP, defaultNacosRegisterConfig.Weight)
	viper.SetDefault(NacosRegisterServicePort, defaultNacosRegisterConfig.Weight)
	viper.SetDefault(NacosRegisterWeight, defaultNacosRegisterConfig.Weight)
	viper.SetDefault(NacosRegisterClusterName, defaultNacosRegisterConfig.ClusterName)
	viper.SetDefault(NacosRegisterEnable, defaultNacosRegisterConfig.Enable)
	viper.SetDefault(NacosRegisterHealthy, defaultNacosRegisterConfig.Healthy)
	viper.SetDefault(NacosRegisterEphemeral, defaultNacosRegisterConfig.Ephemeral)

}

// GetNacosRegisterConfig get basic configuration
//
//goland:noinspection GoUnusedExportedFunction
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
