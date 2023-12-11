package pulsar

import (
	"context"

	"common/micro"
)

// ConsumerElementKey pulsar消息队列消费者模块
var ConsumerElementKey = micro.ElementKey("PulsarConsumerComponent")

// ConsumerComponent pulsar消费者模块
type ConsumerComponent struct {
	micro.EmptyComponent
	client *Consumer
}

// Name of the component
func (c *ConsumerComponent) Name() string {
	return "PulsarConsumer"
}

// PreInit called before Init()
func (c *ConsumerComponent) PreInit(ctx context.Context) error {
	// load config
	SetDefaultConfig()
	return nil
}

// Init the component
func (c *ConsumerComponent) Init(server *micro.Server) error {
	// init
	var err error
	config := GetConfig()
	c.client, err = NewConsumer(config.URL, config.Topic)
	if err != nil {
		return err
	}
	server.RegisterElement(&ConsumerElementKey, c.client)
	return nil
}

// PostStop called after Stop()
func (c *ConsumerComponent) PostStop(ctx context.Context) error {
	// post stop
	return c.client.Close()
}
