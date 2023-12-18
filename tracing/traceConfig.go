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

// TraceConfig  Trace
type TraceConfig struct {
	JaegerCollectorAddr string `toml:"jaegerCollector"`
	JaegerQueryAddr     string `toml:"jaegerQuery"`
	TempoQueryAddr      string `toml:"tempoQuery"`
}

// SetDefaultTraceConfig set default Trace configuration
func SetDefaultTraceConfig() {
}

// GetTraceConfig get Trace configuration
func GetTraceConfig() *TraceConfig {
	return &TraceConfig{
		JaegerCollectorAddr: viper.GetString(traceJaegerCollectorAddr),
		JaegerQueryAddr:     viper.GetString(traceJaegerQueryAddr),
		TempoQueryAddr:      viper.GetString(traceTempoQueryAddr),
	}
}
