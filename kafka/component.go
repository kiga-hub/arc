package kafka

import (
	"context"

	"common/logging"
	"common/micro"
)

// ElementKey is ElementKey for kafka
var ElementKey = micro.ElementKey("KafkaComponent")

// Component is Component for kafka
type Component struct {
	micro.EmptyComponent
	server *Kafka
}

// Name of the component
func (c *Component) Name() string {
	return "Kafka"
}

// PreInit called before Init()
func (c *Component) PreInit(ctx context.Context) error {
	// load config
	SetDefaultConfig()
	return nil
}

// Init the component
func (c *Component) Init(server *micro.Server) error {
	// init
	//var err error
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
		go c.server.CreateProducerKeepalived(ctx)
	}
	return nil
}

// PostStop called after Stop()
func (c *Component) PostStop(ctx context.Context) error {
	// post stop
	if c.server != nil {
		c.server.Close()
	}
	return nil
}
