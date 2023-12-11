package kafka

import (
	"context"
	"time"

	"github.com/kiga-hub/common/utils"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func (k *Kafka) createProducer() (err error) {
	k.producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": k.config.BootStrapServers,
		"client.id":         k.config.ClientID,
		"message.max.bytes": k.config.MessageMaxBytes, //  64*1024*1024
	})
	return
}

// CreateProducerKeepalived -
func (k *Kafka) CreateProducerKeepalived(ctx context.Context) {
	if err := k.createProducer(); err != nil {
		k.logger.Error(err)
	} else {
		k.logger.Infow("kafka producer started", "addr", k.config.BootStrapServers)
	}

	for {
		if k.isClose.Load() {
			return
		}

		select {
		case e := <-k.producer.Events():
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					k.logger.Errorw("Delivery failed", "topic_partition", ev.TopicPartition)
				}
				continue
			case kafka.Error:
				if ev.Code() == kafka.ErrAllBrokersDown {
					k.logger.Errorw("Kafka ErrAllBrokersDown", "code", ev.Code())
				}
			default:
				break
			}
		case <-ctx.Done():
			k.producer.Flush(1000)
			k.producer.Close()
			return
		}
	}
}

// ProduceData -
func (k *Kafka) ProduceData(topic string, key, value []byte) {
	k.producer.ProduceChannel() <- &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Timestamp: time.Now(), // now
		Key:       key,        // now
		Value:     value,      // json,
	}
}

// ProduceDataWithTimeKey -
func (k *Kafka) ProduceDataWithTimeKey(topic string, value []byte) {
	k.ProduceData(topic, utils.Str2byte(time.Now().String()), value)
}

// ProduceDataSimple -
func (k *Kafka) ProduceDataSimple(value []byte) {
	k.ProduceData(k.config.Topic, utils.Str2byte(time.Now().String()), value)
}
