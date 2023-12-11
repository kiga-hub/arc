package tracing

import (
	"context"
	"io"

	"github.com/davecgh/go-spew/spew"
	"github.com/opentracing/opentracing-go"

	platformConf "github.com/kiga-hub/common/conf"
	"github.com/kiga-hub/common/micro"
	microConf "github.com/kiga-hub/common/micro/conf"
)

// ElementKey is ElementKey for tracing
var ElementKey = micro.ElementKey("TracingComponent")

// Component is Component for tracing
type Component struct {
	micro.EmptyComponent
	closer io.Closer
	enable bool
}

// Name of the component
func (c *Component) Name() string {
	return "Trace"
}

// PreInit called before Init()
func (c *Component) PreInit(ctx context.Context) error {
	// load config
	SetDefaultTraceConfig()
	return nil
}

// SetDynamicConfig called when get dynamic config for the first time
func (c *Component) SetDynamicConfig(config *platformConf.NodeConfig) error {
	c.enable = config.APM != nil && !config.APM.EnableTrace
	return nil
}

// Init the component
func (c *Component) Init(server *micro.Server) error {
	// init
	var err error
	var tracer opentracing.Tracer
	// setup tracer
	basicConf := microConf.GetBasicConfig()
	traceConf := GetTraceConfig()
	//logger := server.Get(&logging.Key).(logging.ILogger)

	if c.enable {
		server.RegisterElement(&ElementKey, opentracing.NoopTracer{})
		return nil
	}

	tracer, c.closer, err = CreateTracer(*basicConf, *traceConf, nil)
	if err != nil {
		panic("Could not initialize jaeger tracer: " + err.Error())
	}
	server.RegisterElement(&ElementKey, tracer)

	if basicConf.IsDevMode {
		spew.Dump(traceConf)
	}
	return nil
}

// PostStop called after Stop()
func (c *Component) PostStop(ctx context.Context) error {
	// post stop
	return c.closer.Close()
}
