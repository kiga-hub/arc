package machinery

import (
	"context"
	"fmt"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/opentracing/opentracing-go"

	"github.com/kiga-hub/arc/logging"
	"github.com/kiga-hub/arc/redis"
)

// Invoke invoke
type Invoke func(ctx context.Context, taskType, taskData string) (string, error)

// IWorker work interface
type IWorker interface {
	Work(ctx context.Context, taskType, taskData string) (string, error)
	GetQueueName() string
	Start()
	Stop()
	Release() error
}

// InitMQInitMachineryServerWorker initialize the machinary message queue service work.
func InitMQInitMachineryServerWorker(
	queueName string, invoke Invoke, redis *redis.Config,
	concurrencyNum int, errorsChan chan<- error,
	tracer opentracing.Tracer, logger logging.ILogger,
) (*machinery.Worker, error) {
	server, err := InitMachineryServer(queueName, redis)
	if err != nil {
		return nil, err
	}

	taskMaps := map[string]interface{}{
		"WorkerInvoke": func(ctx context.Context, taskType, taskData string) (string, error) {
			// NOTE machinery will bring a noopSpan in the context if no other span set
			if tracer == nil {
				return invoke(ctx, taskType, taskData)
			}
			sig := tasks.SignatureFromContext(ctx)
			if len(sig.Headers) == 0 {
				return invoke(ctx, taskType, taskData)
			}
			spanContext, err := tracer.Extract(opentracing.TextMap, sig.Headers)
			// if err == opentracing.ErrSpanContextNotFound {
			if err != nil || spanContext == nil {
				return invoke(ctx, taskType, taskData)
			}
			span := tracer.StartSpan(
				"WorkerInvoke",
				opentracing.ChildOf(spanContext),
			)
			defer span.Finish()
			ctx = opentracing.ContextWithSpan(ctx, span)
			return invoke(ctx, taskType, taskData)
		},
	}
	err = server.RegisterTasks(taskMaps)
	if err != nil {
		return nil, err
	}

	consumerTag := "machinery_task"
	logger.Infow("consumer", "concurrency_num", concurrencyNum)
	worker := server.NewWorker(consumerTag, concurrencyNum)
	worker.SetPostTaskHandler(genPostTaskHandler(logger))
	worker.SetErrorHandler(genErrorHandle(logger))
	worker.SetPreTaskHandler(genPreTaskHandler(logger))
	worker.LaunchAsync(errorsChan)
	return worker, nil
}

// genErrorHandle  handle error
func genErrorHandle(logger logging.ILogger) func(err error) {
	return func(err error) {
		logger.Error(err)
	}
}

// genPreTaskHandler pre-task handler
func genPreTaskHandler(logger logging.ILogger) func(signature *tasks.Signature) {
	return func(signature *tasks.Signature) {
		logger.Infow("start task", "group", signature.GroupUUID, "uuid", signature.UUID)
	}
}
func genPostTaskHandler(logger logging.ILogger) func(signature *tasks.Signature) {
	return func(signature *tasks.Signature) {
		logger.Infow("finished task", "group", signature.GroupUUID, "uuid", signature.UUID)
	}
}

// InitMachineryServer initialize Machinery service
func InitMachineryServer(queueName string, redis *redis.Config) (*machinery.Server, error) {
	queueConfig := &config.Config{
		Broker:          fmt.Sprintf("redis://:%s@%s/0", redis.Password, redis.Address),
		DefaultQueue:    queueName,
		ResultBackend:   fmt.Sprintf("redis://:%s@%s/1", redis.Password, redis.Address),
		ResultsExpireIn: 3600 * 2, // 2 hours
		Redis: &config.RedisConfig{
			MaxIdle:                redis.MaxIdle,
			MaxActive:              redis.MaxActive,
			IdleTimeout:            redis.IdleTimeout,
			ConnectTimeout:         redis.ConnTimeout,
			Wait:                   redis.Wait,
			ReadTimeout:            15,
			WriteTimeout:           15,
			NormalTasksPollPeriod:  1000,
			DelayedTasksPollPeriod: 20,
		},
		NoUnixSignals: false,
	}
	return machinery.NewServer(queueConfig)
}

// GetQueueName Get Queue Name
//
//goland:noinspection GoUnusedExportedFunction
func GetQueueName(workerID, taskType string, reserved int) string {
	queue := fmt.Sprintf("%s_%s", taskType, workerID)
	if reserved == 23123 {
		queue = fmt.Sprintf("%s_kiga", taskType)
	}
	return queue
}
