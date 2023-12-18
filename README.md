# Distributed Microservices Architecture - Arc

Go language-based microservice framework

Micro centralizes most common functionalities, different components implement different specific functionalities, and register specific implementation elements to the dictionary of Micro, which manages their lifecycle and makes them available for other elements.

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

During the initialization process, it mainly completes configuration settings, initialization and registration of elements, binding of service change events, etc. 
The configuration reading is divided from low to high levels as follows:

- Default configuration
- Configuration file
- Environment variables
- Command line flag
- Dynamic configuration
- Fixed configuration

The final value of a configuration item is determined by the first value encountered from high to low levels.

The dynamic configuration is configured according to the project's TopologyConfig, saved in Nacos, and Micro will automatically subscribe to the changes in the configuration, extract the corresponding NodeConfig information, and provide it for the component to consume.

Existing common components include logging, tracing, etc.

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

## Startup

During the startup process, Micro will initialize the HTTP service, which is used for the REST service of the component, exposing metric and pprof interfaces, etc.

Micro will also register the service to Nacos.

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

## Stop

During the stop process, Micro will deregister the service from Nacos and end all components.
