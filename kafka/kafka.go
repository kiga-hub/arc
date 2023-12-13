package kafka

import (
	confluentincKafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kiga-hub/arc/logging"
	"go.uber.org/atomic"
)

// ErrCodeTimedOut - 超时错误码
var ErrCodeTimedOut = confluentincKafka.ErrTimedOut

// Kafka -
type Kafka struct {
	isClose *atomic.Bool
	config  *Config
	logger  logging.ILogger

	producer *confluentincKafka.Producer
	consumer *confluentincKafka.Consumer
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
