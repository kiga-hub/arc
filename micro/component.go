package micro

import (
	"context"
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/pangpanglabs/echoswagger/v2"

	platformConf "github.com/kiga-hub/arc/conf"
)

// EmptyComponent is the base implementation of IComponent
type EmptyComponent struct {
	IsPrint bool
}

const componentName = "Empty"

// Name of the component
func (c *EmptyComponent) Name() string {
	return componentName
}

// Status of the component
func (c *EmptyComponent) Status() *ComponentStatus {
	return &ComponentStatus{
		IsOK:   true,
		Params: map[string]interface{}{},
	}
}

// PreInit called before Init()
func (c *EmptyComponent) PreInit(ctx context.Context) error {
	_ = ctx
	if c.IsPrint {
		fmt.Println("PreInit")
	}
	return nil
}

// Init the component
func (c *EmptyComponent) Init(server *Server) error {
	_ = server
	if c.IsPrint {
		fmt.Println("Init")
	}
	return nil
}

// PostInit called after Init()
func (c *EmptyComponent) PostInit(ctx context.Context) error {
	_ = ctx
	if c.IsPrint {
		fmt.Println("PostInit")
	}
	return nil
}

// OnConfigChanged called when dynamic config changed
func (c *EmptyComponent) OnConfigChanged(*platformConf.NodeConfig) error { return nil }

// SetDynamicConfig called when get dynamic config for the first time
func (c *EmptyComponent) SetDynamicConfig(*platformConf.NodeConfig) error { return nil }

// GetSubscribeServiceList returns the service that the component need to subscribe
func (c *EmptyComponent) GetSubscribeServiceList() []string {
	return []string{}
}

// OnServiceChanged called when subscribe service changed
func (c *EmptyComponent) OnServiceChanged(services []model.SubscribeService, err error) {
	_ = services
	_ = err
}

// SetupHandler of echo if the component need
func (c *EmptyComponent) SetupHandler(root echoswagger.ApiRoot, base string) error {
	_ = root
	_ = base
	if c.IsPrint {
		fmt.Println("GetHandle")
	}
	return nil
}

// PreStart called before Start()
func (c *EmptyComponent) PreStart(ctx context.Context) error {
	_ = ctx
	if c.IsPrint {
		fmt.Println("PreStart")
	}
	return nil
}

// Start the component
func (c *EmptyComponent) Start(ctx context.Context) error {
	_ = ctx
	if c.IsPrint {
		fmt.Println("Start")
	}
	return nil
}

// PostStart called after Start()
func (c *EmptyComponent) PostStart(ctx context.Context) error {
	_ = ctx
	if c.IsPrint {
		fmt.Println("PostStart")
	}
	return nil
}

// PreStop called before Stop()
func (c *EmptyComponent) PreStop(ctx context.Context) error {
	_ = ctx
	if c.IsPrint {
		fmt.Println("PreStop")
	}
	return nil
}

// Stop the component
func (c *EmptyComponent) Stop(ctx context.Context) error {
	_ = ctx
	if c.IsPrint {
		fmt.Println("Stop")
	}
	return nil
}

// PostStop called after Stop()
func (c *EmptyComponent) PostStop(ctx context.Context) error {
	_ = ctx
	if c.IsPrint {
		fmt.Println("PostStop")
	}
	return nil
}
