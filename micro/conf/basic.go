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
	BasicAppName = "basic.appName"
	// BasicAppVersion is key for app version
	BasicAppVersion = "basic.appVersion"

	basicDevMode       = "basic.devMode"
	basicAPIRoot       = "basic.apiRoot"
	basicAPIPort       = "basic.apiPort"
	basicProf          = "basic.prof"
	basicDynamicConfig = "basic.dynamicConfig"
	basicCPUCount      = "basic.cpu"
	basicIsAPIRate     = "basic.isApiRate"
	basicIsAPIBody     = "basic.isApiBody"
	basicIsAPITimeout  = "basic.isApiTimeout"
	basicAPIRate       = "basic.apiRate"
	basicAPIBurst      = "basic.apiBurst"
	basicAPIExpiresIn  = "basic.apiExpires"
	basicAPIBodyLimit  = "basic.apiBodyLimit"
	basicAPITimeout    = "basic.apiTimeout"
	basicInSwarm       = "basic.inSwarm"
	basicWorkLoad      = "basic.workLoad"
)

var defaultBasicConfig = BasicConfig{
	Zone:            "needSetThis",
	Node:            "needSetThis",
	Machine:         "needSetThis",
	Service:         "needSetThis",
	IsDevMode:       false,
	APIRoot:         "/api/need_set_this/v1",
	APIPort:         80,
	IsProf:          false,
	IsDynamicConfig: false,
	CPUCount:        -1,
	IsAPIRate:       true,
	IsAPIBody:       true,
	IsAPITimeout:    true,
	APIRate:         100,
	APIBurst:        30,
	APIExpiresIn:    10,
	APIBodyLimit:    "10MB",
	APITimeout:      15000,
	InSwarm:         true,
	WorkLoad:        0,
}

// BasicConfig 基本配置
type BasicConfig struct {
	Zone            string  `toml:"zone" json:"zone,omitempty"`                       // deployment environment zone code
	Node            string  `toml:"node" json:"node,omitempty"`                       // node
	Machine         string  `toml:"machine" json:"machine,omitempty"`                 // machine
	Service         string  `toml:"service" json:"service,omitempty"`                 // service
	Instance        string  `json:"instance,omitempty"`                               // instance
	AppName         string  `json:"app_name,omitempty"`                               // app name
	AppVersion      string  `json:"app_version,omitempty"`                            // app version
	IsDevMode       bool    `toml:"devMode" json:"is_dev_mode,omitempty"`             // dev mode
	APIRoot         string  `toml:"apiRoot" json:"api_root,omitempty"`                // restful api root path
	APIPort         int     `toml:"apiPort" json:"api_port,omitempty"`                // api export port
	IsProf          bool    `toml:"prof" json:"is_prof,omitempty"`                    // open pProf
	IsDynamicConfig bool    `toml:"dynamicConfig" json:"is_dynamic_config,omitempty"` // use dynamic config
	CPUCount        int     `toml:"cpu" json:"cpu_count,omitempty"`                   // cpu count
	IsAPIRate       bool    `toml:"isApiRate" json:"is_api_rate,omitempty"`           // whether to enable rate limit
	IsAPIBody       bool    `toml:"isApiBody" json:"is_api_body,omitempty"`           // whether to enable body limit
	IsAPITimeout    bool    `toml:"isApiTimeout" json:"is_api_timeout,omitempty"`     // whether to enable timeout limit
	APIRate         float64 `toml:"apiRate" json:"rate,omitempty"`                    // rate
	APIBurst        int     `toml:"apiBurst" json:"burst,omitempty"`                  // burst value. the number if times increased when the request reaches the rate limit.
	APIExpiresIn    int     `toml:"apiExpires" json:"expires_in,omitempty"`           // expire time
	APIBodyLimit    string  `toml:"apiBodyLimit" json:"body_limit,omitempty"`         // query body limit. eg: 10MB
	APITimeout      int     `toml:"apiTimeout" json:"timeout,omitempty"`              // api timeout. ms
	InSwarm         bool    `toml:"inSwarm" json:"inSwarm,omitempty"`                 // in swarm
	WorkLoad        int     `toml:"workLoad" json:"work_load,omitempty"`              // work load
}

// SetDefaultBasicConfig set default basic config
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

// GetBasicConfig get basic config
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
