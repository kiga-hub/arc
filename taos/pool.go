package taos

import (
	"errors"
	"sync"

	"common/logging"
)

// Pool 定义一个结构体,这个实体类型可以作为整体单元被复制,可以作为参数或返回值,或被存储到数组
type Pool struct {
	//定义成员,互斥锁类型
	m sync.Mutex
	//定义成员,通道类型,通道传递的是 TaoClient 类型
	// resources chan *TaoClient
	resources *sync.Map
	//定义工厂成员,类型是func()(*TaoClient,error)
	//error是预定义类型,实际上是个interface接口类型
	factory func(config *Config) (IHandler, error)
	closed  bool
	config  *Config
	logger  logging.ILogger
}

// ErrPoolClosed 定义变量,函数返回的是error类型
var ErrPoolClosed = errors.New("池已经关闭了")

// NewPool - 定义New方法,创建一个池,返回的是Pool类型的指针
// 传入的参数是个函数类型func(*TaoClient,error)和池的大小
func NewPool(config *Config, logger logging.ILogger) (*Pool, error) {
	//使用结构体字面值给结构体成员赋值
	p := Pool{
		factory: NewTaoClient,
		// resources: make(chan *TaoClient, size),
		resources: new(sync.Map),
		config:    config,
		logger:    logger,
	}
	return &p, nil
}

// Acquire - 从池中请求获取一个资源,给Pool类型定义的方法
// 返回的值是*TaoClient类型
func (p *Pool) Acquire(key uint64) (IHandler, error) {
	//基于select的多路复用
	//select会等待case中有能够执行的,才会去执行,等待其中一个能执行就执行
	//default分支会在所有case没法执行时,默认执行,也叫轮询channel
	p.m.Lock()
	v, loaded := p.resources.Load(key)
	if !loaded {
		v, _ = p.resources.LoadOrStore(key, make(chan IHandler, p.config.PoolSize))
	}
	resources := v.(chan IHandler)
	p.m.Unlock()

	/*
		ticker := time.NewTicker(time.Millisecond * 10)
		defer func() {
			ticker.Stop()
		}()
		select {
		case r := <-resources:
			return r, nil
		//如果缓冲通道中没有了,就会执行这里
		case <-ticker.C:
			break
		}
	*/
	select {
	case r := <-resources:
		return r, nil
	//如果缓冲通道中没有了,就会执行这里
	default:
		break
	}

	p.logger.Debugw("taos create new release")
	return p.factory(p.config)

}

// Release 将一个使用后的资源放回池
// 传入的参数是*TaoClient类型
func (p *Pool) Release(r IHandler, key uint64) {
	//使用mutex互斥锁
	p.m.Lock()
	//解锁
	defer p.m.Unlock()
	v, ok := p.resources.Load(key)
	if !ok {
		return
	}

	resources := v.(chan IHandler)

	//如果池都关闭了
	if p.closed {
		//关掉资源
		r.Close()
		return
	}
	//select多路选择
	//如果放回通道的时候满了,就关闭这个资源
	select {
	case resources <- r:
	default:
		r.Close()
	}
}

// Close - 关闭资源池,关闭通道,将通道中的资源关掉
func (p *Pool) Close() {
	p.m.Lock()
	defer p.m.Unlock()
	p.closed = true
	//先关闭通道再清空资源
	p.logger.Debugw("taos close release")
	p.resources.Range(func(key, value interface{}) bool {
		resources := value.(chan IHandler)
		close(resources)
		//清空并关闭资源
		for r := range resources {
			r.Close()
		}
		return true
	})
}
