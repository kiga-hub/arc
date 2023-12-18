package pulsar

import (
	"context"
	"time"

	"github.com/kiga-hub/arc/logging"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Producer pulsar producer
type Producer struct {
	ctx        context.Context // context
	cancel     func()          // cancel
	logger     logging.ILogger // logger
	baseClient pulsar.Client   // client
	producer   pulsar.Producer // producer
}

// Close .
func (c *Producer) Close() error {
	c.cancel()
	c.producer.Close()
	c.baseClient.Close()
	return nil
}

// Send send message
func (c *Producer) Send(data []byte) {
	c.producer.SendAsync(c.ctx, &pulsar.ProducerMessage{
		Payload: data,
	}, func(_ pulsar.MessageID, _ *pulsar.ProducerMessage, err error) {
		if err != nil {
			c.logger.Error("pulsar async message", zap.Error(err))
		}
	})
}

// NewProducer .
// url pulsar://localhost:6600,localhost:6650
// topic topic
func NewProducer(logger logging.ILogger, url string, topic string) (*Producer, error) {
	c, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               url,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create pulsar connection failed")
	}
	cp, err := c.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	if err != nil {
		return nil, errors.Wrap(err, "create producer failed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Producer{
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
		baseClient: c,
		producer:   cp,
	}, nil
}
