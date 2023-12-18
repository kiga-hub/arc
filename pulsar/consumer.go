package pulsar

import (
	"context"
	"os"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/pkg/errors"
)

// Consumer pulsar consumer client
type Consumer struct {
	ctx            context.Context             // context
	cancel         func()                      // cancel
	baseClient     pulsar.Client               // base client
	consumer       pulsar.Consumer             // consumer
	messageChannel chan pulsar.ConsumerMessage // store message channel
}

// Close .
func (c *Consumer) Close() error {
	c.cancel()
	c.consumer.Close()
	c.baseClient.Close()
	return nil
}

// ReceiverChannel receive data within the channel
func (c *Consumer) ReceiverChannel(out chan<- []byte) {
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				return
			case msg := <-c.messageChannel:
				out <- msg.Payload()
			}
		}
	}()
}

// NewConsumer .
// url pulsar://localhost:6600,localhost:6650
// topic topic
func NewConsumer(url string, topic string) (*Consumer, error) {
	c, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               url,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create pulsar connection failed")
	}
	hostname, err := os.Hostname()
	if err != nil {
		return nil, errors.Wrap(err, "get machine name failed")
	}
	mc := make(chan pulsar.ConsumerMessage, 1)
	cs, err := c.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: hostname,
		Type:             pulsar.Shared,
		MessageChannel:   mc,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create consumer failed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Consumer{
		ctx:            ctx,
		cancel:         cancel,
		baseClient:     c,
		consumer:       cs,
		messageChannel: mc,
	}, nil
}
