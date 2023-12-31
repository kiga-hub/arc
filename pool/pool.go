package pool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kiga-hub/arc/conf"
	error2 "github.com/kiga-hub/arc/error"
	"github.com/kiga-hub/arc/logging"
)

// ChannelPool store connection information
type ChannelPool struct {
	mu          sync.Mutex
	conns       chan *IdleConn
	idleTimeout time.Duration
	active      int
	logger      logging.ILogger
}

// TODO move this
var ids uint32 = 0
var busyCount int32 = 0
var vObjFactory func() IComputation
var total = 0

// IdleConn idle connection
type IdleConn struct {
	conn IComputation
	t    time.Time
}

// NewPool new 1 pool
func NewPool(poolConfig *conf.PoolConfig, voiceFactory func() IComputation, logger logging.ILogger) (*ChannelPool, error) {
	logger.Infow("NewPool init", "min", poolConfig.MinActive, "max", poolConfig.MaxActive)
	vObjFactory = voiceFactory

	if poolConfig.MinActive < 0 || poolConfig.MaxActive <= 0 || poolConfig.MinActive > poolConfig.MaxActive {
		logger.Errorw("NewPool init", "min", poolConfig.MinActive, "max", poolConfig.MaxActive)
		return nil, error2.ErrInvalidCapacitySettings
	}
	c := ChannelPool{
		conns:       make(chan *IdleConn, poolConfig.MaxActive),
		active:      poolConfig.MaxActive,
		idleTimeout: time.Duration(poolConfig.IdleTimeout) * time.Second,
		logger:      logger,
	}

	for i := 0; i < poolConfig.MinActive; i++ {
		workerObj := vObjFactory()
		id := atomic.AddUint32(&ids, 1)
		err := workerObj.InitWorker(id)
		if err != nil {
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.conns <- &IdleConn{conn: workerObj, t: time.Now()}
	}
	total = poolConfig.MinActive
	return &c, nil
}

// GetConns get all connection
func (c *ChannelPool) GetConns() chan *IdleConn {
	c.mu.Lock()
	defer c.mu.Unlock()

	conns := c.conns
	return conns
}

// Get get a connection from pool
func (c *ChannelPool) Get() (IComputation, error) {
	c.logger.Debugw("start to get object", "available", c.active)
	conns := c.GetConns()
	if conns == nil {
		c.logger.Error("start to get object fail with", error2.ErrPoolIsClosed)
		return nil, error2.ErrPoolIsClosed
	}
	for {
		select {
		case wrapConn := <-conns:
			if wrapConn == nil {
				c.logger.Error("start to get voice warpConn = nil")
				return nil, error2.ErrPoolIsClosed
			}
			// determine wherther it is time out. if it is timeout, discard it.
			if timeout := c.idleTimeout; timeout > 0 {
				if wrapConn.t.Add(timeout).Before(time.Now()) {
					// discard and close connection.
					err := c.Close(wrapConn.conn)
					if err != nil {
						c.logger.Error(" c.Close(wrapConn.conn)err:", err)
					}
					continue
				}
			}
			c.active--
			wrapConn.conn.OnBusy()
			x := atomic.AddInt32(&busyCount, 1)
			c.logger.Infow("info", "get", wrapConn.conn.GetID(), "busy", x, "total", total)
			return wrapConn.conn, nil
		default:
			if c.active < 1 {
				time.Sleep(1 * time.Millisecond)
				continue
			}
			c.mu.Lock()
			conn := vObjFactory()
			id := atomic.AddUint32(&ids, 1)
			err := conn.InitWorker(id)
			if err != nil {
				c.mu.Unlock()
				return conn, err
			}
			total++
			c.active--
			conn.OnBusy()
			x := atomic.AddInt32(&busyCount, 1)
			c.logger.Infow("info", "get", conn.GetID(), "busy", x, "total", total)
			c.mu.Unlock()
			return conn, nil
		}
	}
}

// Put put the connection back into the pool
func (c *ChannelPool) Put(conn IComputation) error {
	c.logger.Debug("Put")
	if conn == nil {
		c.logger.Error(error2.ErrConnectionIsNil)
		return error2.ErrConnectionIsNil
	}

	conn.OnIdle()
	x := atomic.AddInt32(&busyCount, -1)
	c.logger.Infow("info", "get", conn.GetID(), "busy", x, "total", total)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conns == nil {
		return c.Close(conn)
	}

	select {
	case c.conns <- &IdleConn{conn: conn, t: time.Now()}:
		c.active++
		return nil
	default:
		// The Connection pool is full.close the connection directly.
		return c.Close(conn)
	}
}

// Close close single connection
func (c *ChannelPool) Close(conn IComputation) error {
	c.logger.Debug("Close")
	if conn == nil {
		return error2.ErrConnectionIsNil
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	total--
	return conn.Release()
}

// Release releasef all connection in ther connection pool
func (c *ChannelPool) Release() error {
	c.logger.Debug("Release")
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.mu.Unlock()

	if conns == nil {
		return nil
	}
	close(conns)
	var err error
	for wrapConn := range conns {
		err = wrapConn.conn.Release()
	}
	total = 0
	return err
}

// Len existing connections in the connection pool
func (c *ChannelPool) Len() int {
	c.logger.Debug("Len")
	return len(c.GetConns())
}

// ApplyLen  The number of connection that can be applied for.
func (c *ChannelPool) ApplyLen() int {
	c.logger.Debug("ApplyLen")
	return c.active
}
