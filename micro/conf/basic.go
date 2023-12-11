package conf

import (
	"github.com/spf13/viper"
)

const (
	basicZone = "basic.zone"
	basicNode = "basic.node"
	// BasicMachine is machine code (same to private cluster)
	BasicMachine = "basic.machine"
	basicService = "basic.service"

	// BasicInstance is key for Host name
	BasicInstance = "basic.instance"
	// BasicAppName is key for app name
	BasicAppName = "basic.appname"
	// BasicAppVersion is key for app version
	BasicAppVersion = "basic.appversion"

	basicDevMode       = "basic.devmode"
	basicAPIRoot       = "basic.apiroot"
	basicAPIPort       = "basic.apiport"
	basicProf          = "basic.prof"
	basicDynamicConfig = "basic.dynamicconfig"
	basicCPUCount      = "basic.cpu"

	basicIsAPIRate    = "basic.isapirate"
	basicIsAPIBody    = "basic.isapibody"
	basicIsAPITimeout = "basic.isapitimeout"

	basicAPIRate      = "basic.apirate"
	basicAPIBurst     = "basic.apiburst"
	basicAPIExpiresIn = "basic.apiexpires"
	basicAPIBodyLimit = "basic.apibodylimit"
	basicAPITimeout   = "basic.apitimeout"

	basicInSwarm  = "basic.inswarm"
	basicWorkLoad = "basic.workLoad"
)

var defaultBasicConfig = BasicConfig{
	Zone:    "needsetthis",
	Node:    "needsetthis",
	Machine: "needsetthis",
	Service: "needsetthis",

	IsDevMode:       false,
	APIRoot:         "/api/needsetthis/v1",
	APIPort:         80,
	IsProf:          false,
	IsDynamicConfig: false,
	CPUCount:        -1,

	IsAPIRate:    true,
	IsAPIBody:    true,
	IsAPITimeout: true,

	APIRate:      100,
	APIBurst:     30,
	APIExpiresIn: 10,
	APIBodyLimit: "10MB",
	APITimeout:   15000,

	InSwarm:  true,
	WorkLoad: 0,
}

//BasicConfig 基本配置
type BasicConfig struct {
	Zone    string `toml:"zone" json:"zone,omitempty"`       // 部署环境编码
	Node    string `toml:"node" json:"node,omitempty"`       // 节点编码
	Machine string `toml:"machine" json:"machine,omitempty"` // 主机编码
	Service string `toml:"service" json:"service,omitempty"` // 服务

	Instance   string `json:"instance,omitempty"`    // 实例
	AppName    string `json:"app_name,omitempty"`    // 应用名称
	AppVersion string `json:"app_version,omitempty"` // 应用版本
	// BuildVersion string `json:"build_version"`
	// BuildTime    string `json:"build_time"`
	// GitRevision  string `json:"git_revision"`
	// GitBranch    string `json:"git_branch"`
	// GoVersion    string `json:"go_version"`

	IsDevMode       bool   `toml:"devmode" json:"is_dev_mode,omitempty"`             // 是否开发模式
	APIRoot         string `toml:"apiroot" json:"api_root,omitempty"`                // restful api根路径
	APIPort         int    `toml:"apiport" json:"api_port,omitempty"`                // api暴露端口
	IsProf          bool   `toml:"prof" json:"is_prof,omitempty"`                    // 是否调试性能
	IsDynamicConfig bool   `toml:"dynamicconfig" json:"is_dynamic_config,omitempty"` // 是否响应动态配置
	CPUCount        int    `toml:"cpu" json:"cpu_count,omitempty"`                   // 线程限制

	IsAPIRate    bool    `toml:"isapirate" json:"is_api_rate,omitempty"`       // 是否启用限速
	IsAPIBody    bool    `toml:"isapibody" json:"is_api_body,omitempty"`       // 是否启用限制Body大小
	IsAPITimeout bool    `toml:"isapitimeout" json:"is_api_timeout,omitempty"` // 是否启用超时
	APIRate      float64 `toml:"apirate" json:"rate,omitempty"`                // 速率
	APIBurst     int     `toml:"apiburst" json:"burst,omitempty"`              // 突发值,请求队列达到速率限制时,增加的次数
	APIExpiresIn int     `toml:"apiexpires" json:"expires_in,omitempty"`       // 过期时间,单位秒
	APIBodyLimit string  `toml:"bapiodylimit" json:"body_limit,omitempty"`     // 请求body大小限制,单位MB
	APITimeout   int     `toml:"apitimeout" json:"timeout,omitempty"`          // 服务端超时,单位毫秒

	InSwarm  bool `toml:"inswarm" json:"inswarm,omitempty"`    // 是否在Swarm内
	WorkLoad int  `toml:"workLoad" json:"work_load,omitempty"` // 对外工作负载数量
}

//SetDefaultBasicConfig 设置默认基本配置
func SetDefaultBasicConfig() {
	viper.SetDefault(basicZone, defaultBasicConfig.Zone)
	viper.SetDefault(basicNode, defaultBasicConfig.Node)
	viper.SetDefault(BasicMachine, defaultBasicConfig.Machine)
	viper.SetDefault(basicService, defaultBasicConfig.Service)

	viper.SetDefault(basicDevMode, defaultBasicConfig.IsDevMode)
	viper.SetDefault(basicAPIRoot, defaultBasicConfig.APIRoot)
	viper.SetDefault(basicAPIPort, defaultBasicConfig.APIPort)
	viper.SetDefault(basicProf, defaultBasicConfig.IsProf)
	viper.SetDefault(basicDynamicConfig, defaultBasicConfig.IsDynamicConfig)
	viper.SetDefault(basicCPUCount, defaultBasicConfig.CPUCount)

	viper.SetDefault(basicIsAPIRate, defaultBasicConfig.IsAPIRate)
	viper.SetDefault(basicIsAPIBody, defaultBasicConfig.IsAPIBody)
	viper.SetDefault(basicIsAPITimeout, defaultBasicConfig.IsAPITimeout)

	viper.SetDefault(basicAPIRate, defaultBasicConfig.APIRate)
	viper.SetDefault(basicAPIBurst, defaultBasicConfig.APIBurst)
	viper.SetDefault(basicAPIExpiresIn, defaultBasicConfig.APIExpiresIn)
	viper.SetDefault(basicAPIBodyLimit, defaultBasicConfig.APIBodyLimit)
	viper.SetDefault(basicAPITimeout, defaultBasicConfig.APITimeout)

	viper.SetDefault(basicInSwarm, defaultBasicConfig.InSwarm)
	viper.SetDefault(basicWorkLoad, defaultBasicConfig.WorkLoad)
}

//GetBasicConfig 获取基本配置
func GetBasicConfig() *BasicConfig {
	return &BasicConfig{
		Zone:            viper.GetString(basicZone),
		Node:            viper.GetString(basicNode),
		Machine:         viper.GetString(BasicMachine),
		Service:         viper.GetString(basicService),
		Instance:        viper.GetString(BasicInstance),
		IsDevMode:       viper.GetBool(basicDevMode),
		AppName:         viper.GetString(BasicAppName),
		AppVersion:      viper.GetString(BasicAppVersion),
		APIRoot:         viper.GetString(basicAPIRoot),
		APIPort:         viper.GetInt(basicAPIPort),
		IsProf:          viper.GetBool(basicProf),
		IsDynamicConfig: viper.GetBool(basicDynamicConfig),
		CPUCount:        viper.GetInt(basicCPUCount),
		IsAPIRate:       viper.GetBool(basicIsAPIRate),
		IsAPIBody:       viper.GetBool(basicIsAPIBody),
		IsAPITimeout:    viper.GetBool(basicIsAPITimeout),
		APIRate:         viper.GetFloat64(basicAPIRate),
		APIBurst:        viper.GetInt(basicAPIBurst),
		APIExpiresIn:    viper.GetInt(basicAPIExpiresIn),
		APIBodyLimit:    viper.GetString(basicAPIBodyLimit),
		APITimeout:      viper.GetInt(basicAPITimeout),
		InSwarm:         viper.GetBool(basicInSwarm),
		WorkLoad:        viper.GetInt(basicWorkLoad),
	}
}
