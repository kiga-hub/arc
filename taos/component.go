package taos

import (
	"context"

	"github.com/kiga-hub/common/logging"
	"github.com/kiga-hub/common/micro"
)

// ElementKey is ElementKey for taos
var ElementKey = micro.ElementKey("TaosComponent")

// Component is Component for kafka
type Component struct {
	micro.EmptyComponent
	client *Client
}

// Name of the component
func (c *Component) Name() string {
	return "Taos"
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
	var err error
	conf := GetConfig()
	if !conf.Enable {
		return nil
	}
	c.client, err = NewClient(conf, server.GetElement(&micro.LoggingElementKey).(logging.ILogger))
	if err != nil {
		return err
	}
	server.RegisterElement(&ElementKey, c.client)
	return nil
}

// PostStop called after Stop()
func (c *Component) PostStop(ctx context.Context) error {
	// post stop
	if c.client != nil {
		c.client.Close()
	}
	return nil
}
