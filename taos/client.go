package taos

import (
	"fmt"

	"common/logging"
)

// Client -
type Client struct {
	config *Config
	pool   *Pool
	logger logging.ILogger
}

// NewClient -
func NewClient(config *Config, logger logging.ILogger) (*Client, error) {
	var c Client
	c.config = config
	pool, err := NewPool(config, logger)
	if err != nil {
		return nil, err
	}

	logger.Infow(fmt.Sprintf("%s:%s@/tcp(%s:%d)/",
		config.User,
		config.Password,
		config.Host,
		config.Port))

	c.pool = pool
	c.logger = logger
	return &c, nil
}

// Close -
func (c *Client) Close() {
	c.pool.Close()
}

// GetConfig -
func (c *Client) GetConfig() *Config {
	return c.config
}

// Ping -
func (c *Client) Ping() error {
	tao, err := c.pool.Acquire(0)
	if err != nil {
		return err
	}
	defer c.pool.Release(tao, 0)
	return tao.Ping()
}

// ExecByPoolKey -
func (c *Client) ExecByPoolKey(sql string, key uint64) (int64, error) {
	tao, err := c.pool.Acquire(key)
	if err != nil {
		return 0, err
	}
	defer c.pool.Release(tao, key)
	return tao.Exec(sql)
}

// Exec -
func (c *Client) Exec(sql string) (int64, error) {
	return c.ExecByPoolKey(sql, 0)
}

// CreateDatabase -
func (c *Client) CreateDatabase() (int64, error) {
	tao, err := c.pool.Acquire(0)
	if err != nil {
		return 0, err
	}
	defer c.pool.Release(tao, 0)
	sql := fmt.Sprintf("create database if not exists %s precision '%s'", c.config.Name, c.config.Precision)
	c.logger.Debugw(sql)
	if _, err := c.Exec(sql); err != nil {
		return 0, err
	}
	sql = "use " + c.config.Name
	c.logger.Debugw(sql)
	return c.Exec(sql)
}

// DropDatabase -
func (c *Client) DropDatabase() (int64, error) {
	tao, err := c.pool.Acquire(0)
	if err != nil {
		return 0, err
	}
	defer c.pool.Release(tao, 0)
	sql := fmt.Sprintf("drop database if exists %s", c.config.Name)
	c.logger.Debugw(sql)
	return tao.Exec(sql)
}

// QueryByPoolKey -
func (c *Client) QueryByPoolKey(sql string, key uint64) ([]map[string]interface{}, error) {
	tao, err := c.pool.Acquire(key)
	if err != nil {
		return nil, err
	}
	defer c.pool.Release(tao, key)
	return tao.Query(sql)
}

// Query -
func (c *Client) Query(sql string) ([]map[string]interface{}, error) {
	return c.QueryByPoolKey(sql, 0)
}
