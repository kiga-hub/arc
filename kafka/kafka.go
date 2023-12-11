package kafka

import (
	"common/logging"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/atomic"
)

// ErrCodeTimedOut - 超时错误码
var ErrCodeTimedOut = kafka.ErrTimedOut

// Kafka -
type Kafka struct {
	isClose *atomic.Bool
	config  *Config
	logger  logging.ILogger

	producer *kafka.Producer
	consumer *kafka.Consumer
}

// New -
func New(config *Config, logger logging.ILogger) *Kafka {
	return &Kafka{
		config:  config,
		isClose: atomic.NewBool(false),
		logger:  logger,
	}
}

// Close -
func (k *Kafka) Close() {
	k.isClose.Store(true)

	if k.producer != nil {
		k.producer.Flush(1000)
		k.producer.Close()
	}

	if k.consumer != nil {
		k.consumer.Close()
	}
}
