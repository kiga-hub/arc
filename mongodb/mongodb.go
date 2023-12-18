package mongodb

import (
	"context"

	_ "github.com/go-sql-driver/mysql" //_ Import the required drivers
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateDB create db
func CreateDB(config Config) (*mongo.Client, error) {

	var auth string
	if config.User != "" && config.Password != "" {
		auth = config.User + ":" + config.Password + "@"
	}

	var ops string
	if config.Options != "" {
		ops = "?" + config.Options
	}

	connection := "mongodb://" + auth + config.Address + "/" + config.DB + ops
	clientOptions := options.Client().ApplyURI(connection)

	ctx := context.TODO()
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}
