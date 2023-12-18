package machinery

import (
	"fmt"
	"runtime/debug"

	"github.com/RichardKnop/machinery/v1"
	"github.com/opentracing/opentracing-go"

	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/redis"
)

// ConsumerUnit  consumer unit
type ConsumerUnit struct {
	logicWorker   IWorker
	redis         *redis.Config
	concurrentNum int
	errorsChan    chan<- error
	tracer        opentracing.Tracer
	logger        logging.ILogger
	mqWorker      *machinery.Worker
}

// NewConsumerUnit  create a new consumer unit
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

// Start startup
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

// Stop stop
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

// Resume resume
func (unit *ConsumerUnit) Resume(errorsChan chan<- error) {
	unit.mqWorker.LaunchAsync(errorsChan)
}

// Release  release
func (unit *ConsumerUnit) Release() error {
	return unit.logicWorker.Release()
}

// Consumer consumer
type Consumer struct {
	units       []*ConsumerUnit
	redisConfig *redis.Config
	tracer      opentracing.Tracer
	logger      logging.ILogger
	errorsChan  chan<- error
	running     bool
}

// NewConsumer create new consumer object
//
//goland:noinspection GoUnusedExportedFunction
func NewConsumer(redis *redis.Config, errorsChan chan<- error, tracer opentracing.Tracer, logger logging.ILogger) *Consumer {
	return &Consumer{
		units:       []*ConsumerUnit{},
		redisConfig: redis,
		errorsChan:  errorsChan,
		tracer:      tracer,
		logger:      logger,
	}
}

// Register register
func (c *Consumer) Register(logicWorker IWorker, concurrentNum int) error {
	unit, err := NewConsumerUnit(logicWorker, c.redisConfig, concurrentNum, c.errorsChan, c.tracer, c.logger)
	if err == nil {
		c.units = append(c.units, unit)
	}
	return err
}

// Start start
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

// Stop stop
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

// Release release
func (c *Consumer) Release() {
	for _, unit := range c.units {
		err := unit.Release()
		if err != nil {
			fmt.Println(err)
		}
	}
}
