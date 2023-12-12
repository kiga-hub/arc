package configuration

import (
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/spf13/viper"
)

const (
	//nacosClientPrefix nacos客户端字段
	nacosClientPrefix = "nacosClient"
	//nacosClientNamespaceID   nacos命名空间的id 可以为空
	nacosClientNamespaceID = nacosClientPrefix + ".namespaceID"
	//nacosClientTimeoutMs     nacos超时时间(毫秒)
	nacosClientTimeoutMs = nacosClientPrefix + ".timeoutMs"
	//nacosClientNotLoadCacheAtStart  在开始时不加载缓存
	nacosClientNotLoadCacheAtStart = nacosClientPrefix + ".notLoadCacheAtStart"
	//nacosClientLogDir nacos服务日志存放目录
	nacosClientLogDir = nacosClientPrefix + ".logDir"
	//nacosClientCacheDir nacos服务缓存存放目录
	nacosClientCacheDir = nacosClientPrefix + ".cacheDir"
	//nacosClientLogLevel 日志默认级别，值必须是：debug,info,warn,error，默认值是info
	nacosClientLogLevel = nacosClientPrefix + ".loglevel"
)

var defaultNacosClientConfig = constant.ClientConfig{
	NamespaceId:         "",
	TimeoutMs:           15000,
	NotLoadCacheAtStart: true,
	LogDir:              "/tmp/nacos/log",
	CacheDir:            "/tmp/nacos/cache",
	LogLevel:            "warn",
}

// SetDefaultNacosClientConfig 设置默认基本配置
func SetDefaultNacosClientConfig() {
	viper.SetDefault(nacosClientNamespaceID, defaultNacosClientConfig.NamespaceId)
	viper.SetDefault(nacosClientTimeoutMs, defaultNacosClientConfig.TimeoutMs)
	viper.SetDefault(nacosClientNotLoadCacheAtStart, defaultNacosClientConfig.NotLoadCacheAtStart)
	viper.SetDefault(nacosClientLogDir, defaultNacosClientConfig.LogDir)
	viper.SetDefault(nacosClientCacheDir, defaultNacosClientConfig.CacheDir)
	viper.SetDefault(nacosClientLogLevel, defaultNacosClientConfig.LogLevel)
}

// GetNacosClientConfig 获取基本配置
func GetNacosClientConfig() *constant.ClientConfig {

	return &constant.ClientConfig{
		NamespaceId:         viper.GetString(nacosClientNamespaceID),
		TimeoutMs:           viper.GetUint64(nacosClientTimeoutMs),
		NotLoadCacheAtStart: viper.GetBool(nacosClientNotLoadCacheAtStart),
		LogDir:              viper.GetString(nacosClientLogDir),
		CacheDir:            viper.GetString(nacosClientCacheDir),
		LogLevel:            viper.GetString(nacosClientLogLevel),
	}
}
