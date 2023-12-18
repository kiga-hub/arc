package configuration

import (
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/spf13/viper"
)

const (
	//nacosClientPrefix nacos client field
	nacosClientPrefix = "nacosClient"
	//nacosClientNamespaceID   nacos namespace id
	nacosClientNamespaceID = nacosClientPrefix + ".namespaceID"
	//nacosClientTimeoutMs     nacos client timeout
	nacosClientTimeoutMs = nacosClientPrefix + ".timeoutMs"
	//nacosClientNotLoadCacheAtStart do not load cache at start
	nacosClientNotLoadCacheAtStart = nacosClientPrefix + ".notLoadCacheAtStart"
	//nacosClientLogDir nacos service log storage directory
	nacosClientLogDir = nacosClientPrefix + ".logDir"
	//nacosClientCacheDir nacos service cache storage directory
	nacosClientCacheDir = nacosClientPrefix + ".cacheDir"
	//nacosClientLogLevel The default log level is info.The value must be in [debug,info,warn,error]
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

// SetDefaultNacosClientConfig set default basic configuration
func SetDefaultNacosClientConfig() {
	viper.SetDefault(nacosClientNamespaceID, defaultNacosClientConfig.NamespaceId)
	viper.SetDefault(nacosClientTimeoutMs, defaultNacosClientConfig.TimeoutMs)
	viper.SetDefault(nacosClientNotLoadCacheAtStart, defaultNacosClientConfig.NotLoadCacheAtStart)
	viper.SetDefault(nacosClientLogDir, defaultNacosClientConfig.LogDir)
	viper.SetDefault(nacosClientCacheDir, defaultNacosClientConfig.CacheDir)
	viper.SetDefault(nacosClientLogLevel, defaultNacosClientConfig.LogLevel)
}

// GetNacosClientConfig get basic configuration
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
