package redis

import (
	"context"

	"github.com/kiga-hub/arc/micro"

	"github.com/go-redis/redis"
)

// TODO RedisComponent should maintain redis client for all dbs, including creating new one if not exist yet

// ElementKey is ElementKey for redis
var ElementKey = micro.ElementKey("RedisComponent")

// Component is Component for redis
type Component struct {
	micro.EmptyComponent
	client *redis.Client
}

// Name of the component
func (c *Component) Name() string {
	return "Redis"
}

// PreInit called before Init()
func (c *Component) PreInit(ctx context.Context) error {
	_ = ctx
	// load config
	SetDefaultRedisConfig()
	return nil
}

// Init the component
func (c *Component) Init(server *micro.Server) error {
	// init
	//var err error
	redisConf := GetRedisConfig()

	c.client = redis.NewClient(&redis.Options{
		Addr:     redisConf.Address,
		Password: redisConf.Password,
		DB:       redisConf.DB,
	})
	server.RegisterElement(&ElementKey, c.client)
	return nil
}

// PostStop called after Stop()
func (c *Component) PostStop(ctx context.Context) error {
	_ = ctx
	// post stop
	return c.client.Close()
}
