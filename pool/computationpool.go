package pool

import (
	"sync"

	"github.com/kiga-hub/arc/conf"
	"github.com/kiga-hub/arc/logging"
)

// IComputation computation interface
type IComputation interface {
	InitWorker(id uint32) error
	GetID() uint32
	OnIdle()
	OnBusy()
	Release() error
}

// ComputationPool  define computation pool
type ComputationPool struct {
	logger     logging.ILogger
	ObjectPool *ChannelPool
	WorkerTask func(taskType, taskData string) (string, error)
}

// TODO move this
var getVoiceLock = new(sync.Mutex)

// InitComputationPool init computation pool
//
//goland:noinspection GoUnusedExportedFunction
func InitComputationPool(poolConfig *conf.PoolConfig, computationFactory func() IComputation, initFunc func() error, logger logging.ILogger) (*ComputationPool, error) {
	logger.Info("InitWorkerPool")
	err := initFunc()
	if err != nil {
		return nil, err
	}

	newPool, err := NewPool(poolConfig, computationFactory, logger)

	if err != nil {
		return nil, err
	}

	return &ComputationPool{
		logger:     logger,
		ObjectPool: newPool,
	}, nil
}

// Get -
func (p *ComputationPool) Get() (IComputation, error) {
	p.logger.Debug("Get")
	getVoiceLock.Lock()
	defer getVoiceLock.Unlock()

	v, err := p.ObjectPool.Get()
	if err != nil || v == nil {
		// TODO generate error here
		p.logger.Error("get object from pool failed", err)
	}

	return v, err
}

// Process  -
func (p *ComputationPool) Process(taskType, taskData string) (string, error) {
	p.logger.Debug("Process")
	return p.WorkerTask(taskType, taskData)
}

// Release -
func (p *ComputationPool) Release(computation IComputation) error {
	p.logger.Debug("Release")
	// return queue
	return p.ObjectPool.Put(computation)
}

// Destroy -
func (p *ComputationPool) Destroy() error {
	return p.ObjectPool.Release()
}
