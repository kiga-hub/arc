package kafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// CreateConsumer -
func (k *Kafka) CreateConsumer() error {
	var err error

	if k.consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": k.config.BootStrapServers,
		"message.max.bytes": k.config.MessageMaxBytes, //  64*1024*1024
		"group.id":          k.config.GroupID,
	}); err != nil {
		return err
	}

	// SubscribeTopics - Subscribe to topic(s). Topics can be changed between calls to Subscribe*().
	if err = k.consumer.SubscribeTopics([]string{k.config.Topic}, nil); err != nil {
		return err
	}

	return nil
}

// ConsumerData -
func (k *Kafka) ConsumerData(t time.Duration) (*kafka.Message, kafka.ErrorCode, error) {
	msg, err := k.consumer.ReadMessage(t)
	if err != nil {
		return msg, err.(kafka.Error).Code(), err
	}
	return msg, 0, nil
}
