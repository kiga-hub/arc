package alertmanager

import (
	"context"

	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/micro"
)

// ElementKey is ElementKey for alert manager
var ElementKey = micro.ElementKey("AlertManagerComponent")

// Component is Component for alert manager
type Component struct {
	micro.EmptyComponent
	server *AlertManager
}

// Name of the component
func (c *Component) Name() string {
	return "Alertmanager"
}

// PreInit called before Init()
func (c *Component) PreInit(ctx context.Context) error {
	SetDefaultConfig()
	return nil
}

// Init the component
func (c *Component) Init(server *micro.Server) error {

	conf := GetConfig()
	if !conf.Enable {
		return nil
	}

	c.server = New(conf, server.GetElement(&micro.LoggingElementKey).(logging.ILogger))
	server.RegisterElement(&ElementKey, c.server)
	return nil
}

// Start the component
func (c *Component) Start(ctx context.Context) error {

	if c.server != nil {
		c.server.Start()
	}

	return nil
}

// PostStop called after Stop()
func (c *Component) PostStop(ctx context.Context) error {
	if c.server != nil {
		c.server.Close()
	}

	return nil
}
