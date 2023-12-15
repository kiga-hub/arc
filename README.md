# Arc Mircro service

Go language-based microservice framework

## Microservice and Component

### Component 

- kafka
- redis
- mysql
- mongodb
- pool
- log
- config
- http
- grpc
- pulsar
- tracing
- metrics
- error
- micro
- protobuf

## Config


```go
// BasicConfig basic config
type BasicConfig struct {
	Zone       string `toml:"zone" json:"zone,omitempty"`       // Dev env code
	Node       string `toml:"node" json:"node,omitempty"`       // node code
	Machine    string `toml:"machine" json:"machine,omitempty"` // machine code
	Service    string `toml:"service" json:"service,omitempty"` // service
	Instance   string `json:"instance,omitempty"`               // instance
	AppName    string `json:"app_name,omitempty"`               // app name
	AppVersion string `json:"app_version,omitempty"`            // app version
    ...
}
```

## Init Arc

```go
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
```

## Start Server

```go
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
```

## How to use

```go
import(
    "github.com/kiga-hub/arc"
)
```

example:

```go
// SetupHandler of echo if the component need
func (c *ArcStorageComponent) SetupHandler(root echoswagger.ApiRoot, base string) error {
	basicConf := microConf.GetBasicConfig()
	selfServiceName := basicConf.Service
	c.handler.SetupWeb(root, base, selfServiceName)

	return nil
}
```

define a component:

```go
// ArcStorageElementKey is Element Key for arc storage
var ArcStorageElementKey = micro.ElementKey("ArcStorageComponent")
```
