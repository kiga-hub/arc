package tracing

import (
	"github.com/spf13/viper"
)

const (
	traceJaegerCollectorAddr = "trace.jaegerCollector"
	traceJaegerQueryAddr     = "trace.jaegerQuery"
	traceTempoQueryAddr      = "trace.tempoQuery"
)

//var defaultTraceConfig = TraceConfig{}

//TraceConfig  Trace配置
type TraceConfig struct {
	JaegerCollectorAddr string `toml:"jaegerCollector"`
	JaegerQueryAddr     string `toml:"jaegerQuery"`
	TempoQueryAddr      string `toml:"tempoQuery"`
}

//SetDefaultTraceConfig 设置默认Trace配置
func SetDefaultTraceConfig() {
}

//GetTraceConfig 获取Trace配置
func GetTraceConfig() *TraceConfig {
	return &TraceConfig{
		JaegerCollectorAddr: viper.GetString(traceJaegerCollectorAddr),
		JaegerQueryAddr:     viper.GetString(traceJaegerQueryAddr),
		TempoQueryAddr:      viper.GetString(traceTempoQueryAddr),
	}
}
