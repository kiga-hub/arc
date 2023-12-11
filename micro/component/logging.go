package component

import (
	"context"
	"time"

	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"

	platformConf "github.com/kiga-hub/common/conf"
	"github.com/kiga-hub/common/logging"
	"github.com/kiga-hub/common/logging/conf"
	"github.com/kiga-hub/common/micro"
	microConf "github.com/kiga-hub/common/micro/conf"
)

// LoggingComponent is Component for logging
type LoggingComponent struct {
	micro.EmptyComponent
	zlog   *zap.Logger
	enable bool
}

// Name of the component
func (c *LoggingComponent) Name() string {
	return "Logger"
}

// PreInit called before Init()
func (c *LoggingComponent) PreInit(ctx context.Context) error {
	// load config
	conf.SetDefaultLogConfig()
	c.enable = true
	return nil
}

// SetDynamicConfig called when get dynamic config for the first time
func (c *LoggingComponent) SetDynamicConfig(config *platformConf.NodeConfig) error {
	c.enable = config.APM != nil && config.APM.EnableLog
	return nil
}

// Init the component
func (c *LoggingComponent) Init(server *micro.Server) error {
	// init
	var err error
	// setup logger
	basicConf := microConf.GetBasicConfig()
	// spew.Dump(basicConf)
	logConf := conf.GetLogConfig()
	if !c.enable {
		logConf.GraylogAddr = ""
		logConf.LokiAddr = ""
	}

	// spew.Dump(logConf)
	c.zlog, err = logging.CreateLogger(basicConf, logConf)
	if err != nil {
		return err
	}
	logger := c.zlog.Sugar()
	//logger.Info("Using config file: ", viper.ConfigFileUsed())
	server.RegisterElement(&micro.LoggingElementKey, logger)

	if basicConf.IsDevMode {
		spew.Dump(logConf)
	}
	return nil
}

// PostStop called after Stop()
func (c *LoggingComponent) PostStop(ctx context.Context) error {
	// post stop
	err := c.zlog.Sync()
	time.Sleep(time.Second * 3)
	return err
}
