package cache

import (
	"fmt"
	"sync"
	"time"

	linkedList "github.com/emirpasic/gods/lists/singlylinkedlist"
	"go.uber.org/atomic"

	"github.com/kiga-hub/arc/logging"
)

// SearchRequest is a request for search data in the DataCache
type SearchRequest struct {
	ID       uint64
	TimeFrom int64 // us
	TimeTo   int64 // us
}

// DataCache maintain data
type DataCache struct {
	id         uint64
	lock       *sync.Mutex
	from       int64 //us
	to         int64 //us
	data       *linkedList.List
	totalSize  uint64
	expire     int64 //us
	lastSearch int64 //us
}

// NewDataCache create a new DataCache
func NewDataCache(id uint64, expire int64) *DataCache {
	return &DataCache{
		id:     id,
		lock:   new(sync.Mutex),
		from:   time.Now().UnixMicro(),
		to:     time.Now().UnixMicro(),
		data:   linkedList.New(),
		expire: expire, //us
	}
}

func (c *DataCache) input(dp IDataPoint) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.data.Add(dp)
	c.totalSize += dp.GetSize()
	c.to = dp.GetTime()
}

func (c *DataCache) search(request *SearchRequest) []IDataPoint {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.lastSearch = time.Now().UnixMicro()
	result := make([]IDataPoint, 0, 8)
	var dp IDataPoint
	it := c.data.Iterator()
	for it.Begin(); it.Next(); {
		item := it.Value().(IDataPoint)
		if item.GetTime() < request.TimeFrom {
			continue
		}
		if item.GetTime() > request.TimeTo {
			break
		}
		if dp == nil {
			dp = item
		} else {
			result = append(result, dp)
			dp = item
		}
	}
	if dp != nil {
		result = append(result, dp)
	}
	return result
}

func (c *DataCache) cleanTimeout(idleTimeout int) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	now := time.Now().UnixMicro()
	if idleTimeout > 0 && c.lastSearch > 0 && c.lastSearch < now-int64(idleTimeout) {
		c.data.Clear()
		return true
	}
	threshold := now - c.expire
	it := c.data.Iterator()
	index := -1
	for it.Begin(); it.Next(); {
		d := it.Value().(IDataPoint) // us
		if d.GetTime() < threshold {
			c.totalSize -= d.GetSize()
			c.from = d.GetTime()
			index = it.Index()
		} else {
			break
		}
	}
	if index >= 0 {
		for i := 0; i <= index; i++ {
			c.data.Remove(0)
		}
	}
	return false
}

// DataCacheStat contains the stat of the DataCache
type DataCacheStat struct {
	Count  int
	Size   uint64
	From   int64
	To     int64
	Expire int64
}

func (c *DataCache) stat() *DataCacheStat {
	c.lock.Lock()
	defer c.lock.Unlock()
	return &DataCacheStat{
		Count:  c.data.Size(),
		Size:   c.totalSize,
		From:   c.from,
		To:     c.to,
		Expire: c.expire,
	}
}

// DataCacheContainer handle DataCache of keys
type DataCacheContainer struct {
	caches        *sync.Map
	cleanTimer    *time.Ticker
	idleTimeout   int //minute
	running       *atomic.Bool
	expire        int64 //us
	isPrintStat   bool
	isSearchCache bool
	logger        logging.ILogger
}

// NewDataCacheContainer create a new DataCacheContainer
func NewDataCacheContainer(idleTimeout int, expire int64, isSearchCache bool, logger logging.ILogger) *DataCacheContainer {
	cc := &DataCacheContainer{
		caches:        new(sync.Map),
		cleanTimer:    time.NewTicker(time.Second),
		idleTimeout:   idleTimeout,
		running:       atomic.NewBool(false),
		expire:        expire,
		isSearchCache: isSearchCache,
		logger:        logger,
	}
	if expire <= 0 {
		cc.expire = 10 * 1e6 // default 10s
	}
	return cc
}

// Start the container
func (cc *DataCacheContainer) Start(isPrintStat bool) {
	cc.logger.Info("DataCacheContainer starting")
	go cc.daemon()
	cc.isPrintStat = isPrintStat
	if isPrintStat {
		go cc.statDaemon()
	}
	cc.running.Store(true)
}

// Stop the container
func (cc *DataCacheContainer) Stop() {
	if !cc.running.Load() {
		return
	}
	cc.running.Store(false)
	cc.logger.Debug("DataCacheContainer stopped")
}

// Input data to the container
func (cc *DataCacheContainer) Input(dp IDataPoint) {
	if !cc.running.Load() {
		return
	}

	value, loaded := cc.caches.LoadOrStore(dp.GetID(), NewDataCache(dp.GetID(), cc.expire))
	nc := value.(*DataCache)
	if !loaded {
		cc.logger.Debugf("DataCacheContainer starting cache for id %d...\n", dp.GetID())
	}

	nc.input(dp)
}

// Search data in the container
func (cc *DataCacheContainer) Search(request *SearchRequest) ([]IDataPoint, error) {
	if !cc.running.Load() {
		return nil, fmt.Errorf("cache not run")
	}

	value, loaded := cc.caches.LoadOrStore(request.ID, NewDataCache(request.ID, cc.expire))
	nc := value.(*DataCache)
	if !loaded {
		cc.logger.Debugf("DataCacheContainer starting cache for id %d...\n", request.ID)
	}

	return nc.search(request), nil
}

func (cc *DataCacheContainer) daemon() {
	cc.logger.Debugf("DataCacheContainer run search")
	for {
		<-cc.cleanTimer.C
		if !cc.running.Load() {
			return
		}
		cc.caches.Range(func(key, value interface{}) bool {
			nc := value.(*DataCache)
			timeouts := cc.idleTimeout * 60 * 1e6
			if !cc.isSearchCache {
				timeouts = 0
			}
			if ok := nc.cleanTimeout(timeouts); ok {
				cc.logger.Debugf("DataCacheContainer stopping cache for id %d...\n", key.(uint64))
				cc.caches.Delete(key)
			}
			return true
		})
	}
}

// GetStat -
func (cc *DataCacheContainer) GetStat() map[uint64]*DataCacheStat {
	stat := map[uint64]*DataCacheStat{}
	cc.caches.Range(func(key, value interface{}) bool {
		stat[key.(uint64)] = value.(*DataCache).stat()
		return true
	})
	return stat
}

// PrintStat -
func (cc *DataCacheContainer) PrintStat() {
	output := fmt.Sprintf("========= Cache @ %s ==========\n", time.Now().Format("15:04:05"))
	var totalSize uint64
	var totalCount uint64
	for k, v := range cc.GetStat() {
		if v.Size == 0 {
			output += fmt.Sprintf("%012X: [0B]\n", k)
			continue
		}
		ft := time.UnixMicro(v.From).Format("15:04:05")
		tt := time.UnixMicro(v.To).Format("15:04:05")
		output += fmt.Sprintf("%012X: [%s.%03d-%s.%03d][%s]\n", k, ft, v.From%1e6, tt, v.To%1e6, cc.printSize(v.Size))
		totalSize += v.Size
		totalCount++
	}
	output += fmt.Sprintf("Total: [%d][%s]\n", totalCount, cc.printSize(totalSize))
	output += "=====================================\n"
	fmt.Print(output)
}

func (cc *DataCacheContainer) printSize(size uint64) string {
	var s = float64(size)
	var g float64 = 1024 * 1024 * 1024
	var m float64 = 1024 * 1024
	var k float64 = 1024
	if s > g {
		return fmt.Sprintf("%.3fGB", s/g)
	} else if s > m {
		return fmt.Sprintf("%.3fMB", s/m)
	} else if s > k {
		return fmt.Sprintf("%.3fKB", s/k)
	}
	return fmt.Sprintf("%.0fB", s)
}

func (cc *DataCacheContainer) statDaemon() {
	// print stat with progress bar, etc.
	statTimer := time.NewTicker(time.Second * 5)
	for {
		<-statTimer.C
		if !cc.running.Load() {
			return
		}
		cc.PrintStat()
	}
}
