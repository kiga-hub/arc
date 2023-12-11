package pulsar

import (
	"context"
	"time"

	"common/logging"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Producer pulsar 生产者客户端
type Producer struct {
	ctx        context.Context // 上下文
	cancel     func()          // 取消函数
	logger     logging.ILogger // 日志
	baseClient pulsar.Client   // 原始客户端
	producer   pulsar.Producer // 生产者
}

// Close .
func (c *Producer) Close() error {
	c.cancel()
	c.producer.Close()
	c.baseClient.Close()
	return nil
}

// Send 发送消息
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
// topic 主题
func NewProducer(logger logging.ILogger, url string, topic string) (*Producer, error) {
	c, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               url,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, errors.Wrap(err, "创建pulsar连接失败")
	}
	cp, err := c.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	if err != nil {
		return nil, errors.Wrap(err, "创建生产者失败")
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
