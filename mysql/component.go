package mysql

import (
	"context"

	"common/micro"

	"github.com/jinzhu/gorm"
)

// ElementKey is ElementKey for mysql
var ElementKey = micro.ElementKey("MysqlComponent")

// Component is Component for mysql
type Component struct {
	micro.EmptyComponent
	db *gorm.DB
}

// Name of the component
func (c *Component) Name() string {
	return "Mysql"
}

// PreInit called before Init()
func (c *Component) PreInit(ctx context.Context) error {
	// load config
	SetDefaultMysqlConfig()
	return nil
}

// Init the component
func (c *Component) Init(server *micro.Server) error {
	// init
	var err error
	mysqlConf := GetMysqlConfig()
	// spew.Dump(logConf)
	c.db, err = CreateDB(*mysqlConf)
	if err != nil {
		return err
	}
	server.RegisterElement(&ElementKey, c.db)
	return nil
}

// PostStop called after Stop()
func (c *Component) PostStop(ctx context.Context) error {
	// post stop
	return c.db.Close()
}
