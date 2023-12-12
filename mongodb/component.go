package mongodb

import (
	"context"

	"github.com/kiga-hub/arc/micro"

	"go.mongodb.org/mongo-driver/mongo"
)

// ElementKey is ElementKey for mongodb
var ElementKey = micro.ElementKey("MongodbComponent")

// MongoComponent is Component for mongodb
type MongoComponent struct {
	micro.EmptyComponent
	db *mongo.Database
}

// Name of the component
func (c *MongoComponent) Name() string {
	return "Mongo"
}

// PreInit called before Init()
func (c *MongoComponent) PreInit(ctx context.Context) error {
	_ = ctx
	return nil
}

// Init the component
func (c *MongoComponent) Init(server *micro.Server) error {
	// init
	var err error
	conf := GetMongoConfig()
	client, err := CreateDB(*conf)
	if err != nil {
		return err
	}

	// 通过连接获取配置文件中的数据库对象
	c.db = client.Database(conf.DB)

	server.RegisterElement(&ElementKey, c.db)
	return nil
}

// PostStop called after Stop()
func (c *MongoComponent) PostStop(ctx context.Context) error {
	return c.db.Client().Disconnect(ctx)
}
