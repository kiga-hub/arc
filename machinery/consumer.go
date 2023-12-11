package machinery

import (
	"fmt"
	"runtime/debug"

	"github.com/RichardKnop/machinery/v1"
	"github.com/opentracing/opentracing-go"

	"common/logging"
	"common/redis"
)

// ConsumerUnit  消费者组件
type ConsumerUnit struct {
	logicWorker   IWorker
	redis         *redis.Config
	concurrentNum int
	errorsChan    chan<- error
	tracer        opentracing.Tracer
	logger        logging.ILogger
	mqWorker      *machinery.Worker
}

// NewConsumerUnit  创建一个消费者单元
func NewConsumerUnit(
	logicWorker IWorker, redis *redis.Config, concurrentNum int,
	errorsChan chan<- error, tracer opentracing.Tracer, logger logging.ILogger,
) (*ConsumerUnit, error) {
	unit := &ConsumerUnit{
		logicWorker:   logicWorker,
		redis:         redis,
		concurrentNum: concurrentNum,
		errorsChan:    errorsChan,
		tracer:        tracer,
		logger:        logger,
	}
	return unit, nil
}

// Start 启动
func (unit *ConsumerUnit) Start() error {
	unit.logicWorker.Start()
	mqWorker, err := InitMQInitMachineryServerWorker(
		unit.logicWorker.GetQueueName(),
		unit.logicWorker.Work,
		unit.redis,
		unit.concurrentNum,
		unit.errorsChan,
		unit.tracer,
		unit.logger,
	)
	if err != nil {
		return err
	}
	unit.mqWorker = mqWorker
	return err
}

// Stop 停止
func (unit *ConsumerUnit) Stop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered %v\n", r)
			debug.PrintStack()
		}
	}()
	unit.mqWorker.Quit()
	unit.logicWorker.Stop()
}

// Resume 重新开始
func (unit *ConsumerUnit) Resume(errorsChan chan<- error) {
	unit.mqWorker.LaunchAsync(errorsChan)
}

// Release  释放
func (unit *ConsumerUnit) Release() error {
	return unit.logicWorker.Release()
}

// Consumer 类
type Consumer struct {
	units       []*ConsumerUnit
	redisConfig *redis.Config
	tracer      opentracing.Tracer
	logger      logging.ILogger
	errorsChan  chan<- error
	running     bool
}

// NewConsumer 创建消费者对象
func NewConsumer(redis *redis.Config, errorsChan chan<- error, tracer opentracing.Tracer, logger logging.ILogger) *Consumer {
	return &Consumer{
		units:       []*ConsumerUnit{},
		redisConfig: redis,
		errorsChan:  errorsChan,
		tracer:      tracer,
		logger:      logger,
	}
}

// Register 注册
func (c *Consumer) Register(logicWorker IWorker, concurrentNum int) error {
	unit, err := NewConsumerUnit(logicWorker, c.redisConfig, concurrentNum, c.errorsChan, c.tracer, c.logger)
	if err == nil {
		c.units = append(c.units, unit)
	}
	return err
}

// Start 启动
func (c *Consumer) Start() error {
	if c.running {
		return nil
	}
	c.logger.Info("start units")
	c.running = true
	for _, unit := range c.units {
		err := unit.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

// Stop 停止
func (c *Consumer) Stop() {
	if !c.running {
		return
	}
	c.logger.Info("stop units")
	c.running = false
	for _, unit := range c.units {
		unit.Stop()
	}
}

// Release 释放
func (c *Consumer) Release() {
	for _, unit := range c.units {
		err := unit.Release()
		if err != nil {
			fmt.Println(err)
		}
	}
}
