package pulsar

import (
	"context"

	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/micro"
)

// ProducerElementKey pulsar message queue producer module
var ProducerElementKey = micro.ElementKey("PulsarProducerComponent")

// ProducerComponent pulsar producer module
type ProducerComponent struct {
	micro.EmptyComponent
	client *Producer
}

// Name of the component
func (c *ProducerComponent) Name() string {
	return "PulsarProducer"
}

// PreInit called before Init()
func (c *ProducerComponent) PreInit(ctx context.Context) error {
	_ = ctx
	// load config
	SetDefaultConfig()
	return nil
}

// Init the component
func (c *ProducerComponent) Init(server *micro.Server) error {
	// init
	var err error
	config := GetConfig()
	c.client, err = NewProducer(server.GetElement(&micro.LoggingElementKey).(logging.ILogger),
		config.URL, config.Topic)
	if err != nil {
		return err
	}
	server.RegisterElement(&ProducerElementKey, c.client)
	return nil
}

// PostStop called after Stop()
func (c *ProducerComponent) PostStop(ctx context.Context) error {
	_ = ctx
	// post stop
	return c.client.Close()
}
